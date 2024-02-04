package database

import (
	"context"
	"fmt"
	"gorm.io/gorm/logger"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var AutoMigratedModels = []interface{}{}

type Config struct {
	Host     string
	User     string
	Password string
	Database string
}

type Database struct {
	log *zap.Logger
	db  *gorm.DB
}

func NewDatabase(config Config, log *zap.Logger, lc fx.Lifecycle) *Database {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=5432 sslmode=disable TimeZone=Europe/Moscow",
		config.Host, config.User, config.Password, config.Database)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger:                                   logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatal("error on open db", zap.Error(err))
	}

	d := &Database{
		log: log,
		db:  db,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			d.migrate()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return nil
		},
	})

	return d
}

func (d *Database) DB() *gorm.DB {
	return d.db
}

func (d *Database) migrate() {
	if err := d.DB().AutoMigrate(AutoMigratedModels...); err != nil {
		d.log.Error("error on migrate models", zap.Error(err))
	}
}
