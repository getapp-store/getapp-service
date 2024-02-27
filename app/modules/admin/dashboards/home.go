package dashboards

import (
	"fmt"
	"go.uber.org/zap"
	"html/template"
	"net/http"
	"time"

	tracker "ru/kovardin/getapp/app/modules/tracker/models"
	"ru/kovardin/getapp/pkg/database"
	"ru/kovardin/getapp/pkg/logger"
)

type Home struct {
	log         *logger.Logger
	templates   map[string]*template.Template
	conversions *database.Repository[tracker.Conversion]
	trackers    *database.Repository[tracker.Tracker]
	database    *database.Database
}

func NewHome(
	log *logger.Logger,
	trackers *database.Repository[tracker.Tracker],
	conversions *database.Repository[tracker.Conversion],
	database *database.Database,
) *Home {
	return &Home{
		log:         log,
		trackers:    trackers,
		conversions: conversions,
		database:    database,
		templates: map[string]*template.Template{
			"home": template.Must(template.ParseFiles(
				"templates/admin/home.gohtml",
				"templates/admin/conversions.gohtml",
				"templates/admin/impressions.gohtml",
				"templates/admin/ecpms.gohtml",
				"templates/admin/payments.gohtml",
				"templates/admin/subscriptions.gohtml",
			)),
		},
	}
}

type Data struct {
	Trackers []Tracker
	Data     []time.Time
}

type Tracker struct {
	Name string
	Data []int
}

type Row struct {
	Name      string
	TrackerId int
	Date      time.Time
	Cnt       int
}

// https://apache.github.io/echarts-handbook/en/basics/download
// https://getbootstrap.com/docs/5.0/layout/containers/
func (h *Home) Dashboard(w http.ResponseWriter, r *http.Request) {
	// trackers conversions

	end := time.Now()
	start := end.AddDate(0, -2, 0)

	rows := []Row{}

	h.database.DB().Debug().Table("conversions").
		Select("t.name as name, tracker_id, date(conversions.created_at) as date, count(*) as cnt").
		Joins("LEFT JOIN trackers t on t.id = conversions.tracker_id").
		Where("date_trunc('day', conversions.created_at)::date > ? AND date_trunc('day', conversions.created_at)::date <= ?", start, end).
		Group("t.name, tracker_id, date(conversions.created_at)").
		Order("date(conversions.created_at) DESC").
		Scan(&rows)

	// prepare xaxis
	dates := []time.Time{}
	for d := start; d.After(end) == false; d = d.AddDate(0, 0, 1) {
		dates = append(dates, time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC))
	}

	for _, d := range dates {
		fmt.Println(d)
	}

	// prepare index
	index := map[string]map[time.Time]int{}
	for _, r := range rows {
		t, ok := index[r.Name]
		if !ok {
			t = map[time.Time]int{}
		}

		t[r.Date] = r.Cnt
		index[r.Name] = t
	}

	// prepare data
	trackers := []Tracker{}
	for k, v := range index {
		tracker := Tracker{
			Name: k,
			Data: []int{},
		}

		for _, d := range dates {
			val := v[d]

			tracker.Data = append(tracker.Data, val)
		}

		trackers = append(trackers, tracker)
	}

	data := Data{
		Trackers: trackers,
		Data:     dates,
	}

	// prepare data for graps
	if err := h.templates["home"].ExecuteTemplate(w, "home", struct {
		Conversions Data
	}{
		Conversions: data,
	}); err != nil {
		h.log.Error("pay error", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
	}
}
