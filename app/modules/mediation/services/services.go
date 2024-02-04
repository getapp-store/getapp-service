package services

import (
	"ru/kovardin/getapp/app/modules/mediation/services/bigo"
	"ru/kovardin/getapp/app/modules/mediation/services/cpa"
	"ru/kovardin/getapp/app/modules/mediation/services/mytarget"
	"ru/kovardin/getapp/app/modules/mediation/services/yandex"
)

type Services struct {
	mytarget *mytarget.MyTarget
	yandex   *yandex.Yandex
	cpa      *cpa.CPA
	bigo     *bigo.Bigo
}

func New(target *mytarget.MyTarget, yandex *yandex.Yandex, cpa *cpa.CPA, bigo *bigo.Bigo) *Services {
	return &Services{
		mytarget: target,
		yandex:   yandex,
		cpa:      cpa,
		bigo:     bigo,
	}
}

func (s *Services) Start() {
	go s.mytarget.Start()
	go s.yandex.Start()
	go s.cpa.Start()
	go s.bigo.Start()
}

func (s *Services) Stop() {
	go s.mytarget.Stop()
	go s.yandex.Stop()
	go s.cpa.Stop()
	go s.bigo.Stop()
}
