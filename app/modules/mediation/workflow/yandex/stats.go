package yandex

type Stats struct {
	Result string `json:"result"`
	Data   struct {
		Currencies []struct {
			ID   string `json:"id"`
			Code string `json:"code"`
		} `json:"currencies"`
		TotalRows  int  `json:"total_rows"`
		IsLastPage bool `json:"is_last_page"`
		Measures   struct {
			EcpmPartnerWoNds struct {
				Index    int    `json:"index"`
				Unit     string `json:"unit"`
				Currency string `json:"currency"`
				Title    string `json:"title"`
				Type     string `json:"type"`
			} `json:"ecpm_partner_wo_nds"`
			PartnerWoNds struct {
				Index    int    `json:"index"`
				Unit     string `json:"unit"`
				Currency string `json:"currency"`
				Type     string `json:"type"`
				Title    string `json:"title"`
			} `json:"partner_wo_nds"`
			Impressions struct {
				Type  string `json:"type"`
				Title string `json:"title"`
				Unit  string `json:"unit"`
				Index int    `json:"index"`
			} `json:"impressions"`
		} `json:"measures"`
		Dimensions struct {
			Date struct {
				Title string `json:"title"`
				Type  string `json:"type"`
				Index int    `json:"index"`
			} `json:"date"`
			ComplexBlockID struct {
				Index int    `json:"index"`
				Type  string `json:"type"`
				Title string `json:"title"`
			} `json:"complex_block_id"`
		} `json:"dimensions"`
		Points      []Point    `json:"points"`
		ReportTitle string     `json:"report_title"`
		Periods     [][]string `json:"periods"`
		Totals      struct {
			Num2 []struct {
				Impressions      float64 `json:"impressions"`
				EcpmPartnerWoNds float64 `json:"ecpm_partner_wo_nds"`
				PartnerWoNds     float64 `json:"partner_wo_nds"`
			} `json:"2"`
		} `json:"totals"`
	} `json:"data"`
}

type Point struct {
	Dimensions struct {
		ComplexBlockID string   `json:"complex_block_id"`
		Date           []string `json:"date"`
	} `json:"dimensions"`
	Measures []struct {
		Impressions      float64 `json:"impressions"`
		EcpmPartnerWoNds float64 `json:"ecpm_partner_wo_nds"`
		PartnerWoNds     float64 `json:"partner_wo_nds"`
	} `json:"measures"`
}
