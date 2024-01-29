package parser

import (
	"ru/kovardin/getapp/app/modules/mediation/services/parser/bigo"
	"ru/kovardin/getapp/app/modules/mediation/services/parser/cpa"
	"ru/kovardin/getapp/app/modules/mediation/services/parser/mytarget"
	"ru/kovardin/getapp/app/modules/mediation/services/parser/yandex"
)

type Parser struct {
	mytarget *mytarget.MyTarget
	yandex   *yandex.Yandex
	cpa      *cpa.CPA
	bigo     *bigo.Bigo
}

func New(target *mytarget.MyTarget, yandex *yandex.Yandex, cpa *cpa.CPA, bigo *bigo.Bigo) *Parser {
	return &Parser{
		mytarget: target,
		yandex:   yandex,
		cpa:      cpa,
		bigo:     bigo,
	}
}

func (p *Parser) Start() {
	go p.mytarget.Start()
	go p.yandex.Start()
	go p.cpa.Start()
	go p.bigo.Start()
}

func (p *Parser) Stop() {
	go p.mytarget.Stop()
	go p.yandex.Stop()
	go p.cpa.Stop()
	go p.bigo.Stop()
}
