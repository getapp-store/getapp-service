package repos

import (
	"database/sql"
	"errors"
	"go.uber.org/zap"
	"time"

	"ru/kovardin/getapp/app/modules/mediation/models"
	"ru/kovardin/getapp/pkg/database"
	"ru/kovardin/getapp/pkg/logger"
)

type Cpms struct {
	log *logger.Logger
	db  *database.Database
}

func New(log *logger.Logger, db *database.Database) *Cpms {
	return &Cpms{
		log: log,
		db:  db,
	}
}

func (c *Cpms) CpmsByNetwork(from, to time.Time) ([]models.CpmByNetwork, error) {
	rows, err := c.db.DB().Raw(cpmsByNetwork, from, to).Rows()

	if errors.Is(err, sql.ErrNoRows) {
		return []models.CpmByNetwork{}, nil
	}

	if err != nil {
		return nil, err
	}

	items := []models.CpmByNetwork{}

	for rows.Next() {
		item := models.CpmByNetwork{}
		if err := rows.Scan(&item.Network, &item.Cpm); err != nil {
			c.log.Error("error on scap cpm", zap.Error(err))
			continue
		}

		items = append(items, item)
	}

	return items, nil
}

const cpmsByNetwork = `
	select network_id, avg(amount) 
	from cpms 
	where date > ? and date <= ?
	group by network_id
`
