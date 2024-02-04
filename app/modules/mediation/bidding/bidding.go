package bidding

import (
	"math/rand"
	"sort"
	"time"

	"ru/kovardin/getapp/app/modules/mediation/models"
	"ru/kovardin/getapp/app/modules/mediation/repos"
	"ru/kovardin/getapp/pkg/database"
)

const (
	e        = 0.3
	addition = 100.0
	minimal  = 0.001
)

type Cpms interface {
	CpmsByNetwork(from, to time.Time) ([]models.CpmByNetwork, error)
}

type Bidding struct {
	cpms     *database.Repository[models.Cpm]
	networks Cpms
}

func New(cpms *database.Repository[models.Cpm], networks *repos.Cpms) *Bidding {
	return &Bidding{
		cpms:     cpms,
		networks: networks,
	}
}

// Cpm цена тысячи показов для этого запроса
func (b *Bidding) Cpm(unit models.Unit) (float64, error) {
	cpm, err := b.cpms.First(database.Condition{
		In: map[string]any{
			"unit_id":      unit.ID,
			"placement_id": unit.PlacementId,
			"network_id":   unit.NetworkId,
		},
		Sorting: database.Sorting{
			Sort:  "date",
			Order: "desc",
		},
	})

	if err != nil {
		return 0, err
	}

	if cpm.Amount == 0 {
		cpm.Amount = minimal
	}

	return cpm.Amount, nil
}

func (b *Bidding) Bid(unit models.Unit) (float64, error) {
	return b.Cpm(unit)
}

// Bandit ставка для выбора победителя. В этой логике в качестве ставки используется
// cpm. Этот не подойдет для более сложной логики оценки ставки для пользователя. Придется
// разделить подбор бида и оценку для тестирования
func (b *Bidding) Bandit(bid float64, unit models.Unit) (float64, error) {
	to := time.Now()
	from := to.Add(-time.Hour * 24 * 3)

	// Сначала выбираем список средних CPM по сеткам
	// на самом деле, тут нужно знать что мы поставили
	// в запросах по другим сеткам. Но так тоже подойдет.
	// Этого достаточно, чтобы понять что текущий запрос по сетке,
	// которая пока мало приносит денег
	cmpsByNetwork, err := b.networks.CpmsByNetwork(from, to)

	if err != nil {
		return 0, err
	}

	sort.Slice(cmpsByNetwork, func(i, j int) bool {
		return cmpsByNetwork[i].Cpm > cmpsByNetwork[j].Cpm
	})

	// запрос по самой дорогой сетке, тут ничего не тестим
	if cmpsByNetwork[0].Network == unit.NetworkId {
		return bid, nil
	}

	// только одна сетка в списке
	n := len(cmpsByNetwork)
	if n <= 1 {
		return bid, nil
	}

	p := e / float64(n-1)

	coin := rand.Float64()

	if coin < p {
		return bid + addition, nil
	}

	return bid, nil
}
