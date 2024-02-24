package database

import (
	"gorm.io/gorm"
)

type Repository[T any] struct {
	db *Database
}

func NewRepository[T any](db *Database) *Repository[T] {
	return &Repository[T]{
		db: db,
	}
}

func (s *Repository[T]) Find(condition Condition) ([]T, error) {
	var models []T
	var model T

	err := Query(condition).Build(
		s.db.DB().Model(model),
	).Order("created_at DESC").Find(&models).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return models, err
	}

	return models, nil
}

func (s *Repository[T]) First(condition Condition) (T, error) {
	var model T
	err := Query(condition).Build(
		s.db.DB().Model(model),
	).Order("created_at DESC").First(&model).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return model, err
	}

	return model, nil
}

func (s *Repository[T]) Create(item *T) error {
	return s.db.DB().Create(item).Error
}

func (s *Repository[T]) Save(item *T) error {
	return s.db.DB().Save(item).Error
}

func (s *Repository[T]) Update(item *T, field, from, to string) error {
	return s.db.DB().Model(item).Update(field, to).Where(field+" = ?", from).Error
}

func (s *Repository[T]) Delete(id int) error {
	var model T
	err := s.db.DB().Delete(model, id).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	return nil
}
