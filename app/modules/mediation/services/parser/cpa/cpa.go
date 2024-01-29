package cpa

type CPA struct {
}

func New() *CPA {
	return &CPA{
		// тут могут быть разные CPA сетки
	}
}

func (m *CPA) Start() {
	// get from api
	// pars response
	// count cpm and save to db
}

func (m *CPA) Stop() {

}
