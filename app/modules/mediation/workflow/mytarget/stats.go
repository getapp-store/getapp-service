package mytarget

type Stats struct {
	Items []Item `json:"items"`
	Total struct {
		Histogram []Data `json:"histogram"`
		Total     Data   `json:"total"`
	} `json:"total"`
}

type Item struct {
	Rows []struct {
		Date      string `json:"date"`
		Currency  string `json:"currency"`
		Histogram []Data `json:"histogram"`
		Total     Data   `json:"total"`
	} `json:"rows"`
	Total struct {
		Histogram []Data `json:"histogram"`
		Total     Data   `json:"total"`
	} `json:"total"`
	ID int `json:"id"`
}

type Data struct {
	Clicks           int     `json:"clicks"`
	Shows            int     `json:"shows"`
	Goals            int     `json:"goals"`
	Noshows          int     `json:"noshows"`
	Requests         int     `json:"requests"`
	RequestedBanners int     `json:"requested_banners"`
	ResponsedBlocks  int     `json:"responsed_blocks"`
	ResponsedBanners int     `json:"responsed_banners"`
	Amount           string  `json:"amount"`
	Responses        int     `json:"responses"`
	Cpm              string  `json:"cpm"`
	Ctr              int     `json:"ctr"`
	FillRate         float64 `json:"fill_rate"`
	ShowRate         float64 `json:"show_rate"`
	Vtr              int     `json:"vtr"`
	Vr               int     `json:"vr"`
}
