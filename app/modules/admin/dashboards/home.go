package dashboards

import (
	"html/template"
	"net/http"
	"time"

	"go.uber.org/zap"

	tracker "ru/kovardin/getapp/app/modules/tracker/models"
	"ru/kovardin/getapp/pkg/database"
	"ru/kovardin/getapp/pkg/logger"
	"ru/kovardin/getapp/pkg/utils/chart"
)

type Home struct {
	log       *logger.Logger
	templates map[string]*template.Template
	trackers  *database.Repository[tracker.Tracker]
	database  *database.Database
}

func NewHome(
	log *logger.Logger,
	trackers *database.Repository[tracker.Tracker],
	database *database.Database,
) *Home {
	return &Home{
		log:      log,
		trackers: trackers,
		database: database,
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
	Metrics chart.Mtx
	Data    chart.Dtx
}

// https://apache.github.io/echarts-handbook/en/basics/download
// https://getbootstrap.com/docs/5.0/layout/containers/
func (h *Home) Dashboard(w http.ResponseWriter, r *http.Request) {
	// trackers conversions

	end := time.Now()

	// prepare data for graphs
	if err := h.templates["home"].ExecuteTemplate(w, "home", struct {
		Conversions Data
		Impressions Data
		Ecpms       Data
	}{
		Conversions: h.conversions(end.AddDate(0, -1, 0), end),
		Impressions: h.impressions(end.AddDate(0, -1, 0), end),
		Ecpms:       h.ecpms(end.Add(-time.Hour*24*2), end),
	}); err != nil {
		h.log.Error("pay error", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (h *Home) conversions(start, end time.Time) Data {
	rows := []chart.Item{}

	h.database.DB().Table("conversions").
		Select("t.name as name, tracker_id as id, date(conversions.created_at) as date, count(*) as value").
		Joins("LEFT JOIN trackers t on t.id = conversions.tracker_id").
		Where("date_trunc('day', conversions.created_at)::date > ? AND date_trunc('day', conversions.created_at)::date <= ?", start, end).
		Group("t.name, tracker_id, date(conversions.created_at)").
		Order("date(conversions.created_at) DESC").
		Scan(&rows)

	dates := chart.Dates(start, end)
	index := chart.Index(rows)
	metrics := chart.Metrics(dates, index)

	return Data{
		Metrics: metrics,
		Data:    dates,
	}
}

func (h *Home) impressions(start, end time.Time) Data {
	rows := []chart.Item{}

	h.database.DB().Table("impressions").
		Select("u.name as name, unit_id as id, date(impressions.created_at) as date, count(*) as value").
		Joins("LEFT JOIN units u on u.id = impressions.unit_id").
		Where("date_trunc('day', impressions.created_at)::date > ? AND date_trunc('day', impressions.created_at)::date <= ?", start, end).
		Group("u.name, unit_id, date(impressions.created_at)").
		Order("date(impressions.created_at) DESC").
		Scan(&rows)

	dates := chart.Dates(start, end)
	index := chart.Index(rows)
	metrics := chart.Metrics(dates, index)

	return Data{
		Metrics: metrics,
		Data:    dates,
	}
}

func (h *Home) ecpms(start, end time.Time) Data {
	rows := []chart.Item{}

	h.database.DB().Table("cpms").
		Select("u.name as name, unit_id as id, date_trunc('hour', cpms.created_at)::timestamp as date, sum(amount) as value").
		Joins("LEFT JOIN units u on u.id = cpms.unit_id").
		Where("date_trunc('day', cpms.created_at)::date > ? AND date_trunc('hour', cpms.created_at)::date <= ? AND amount > 0", start, end).
		Group("u.name, unit_id, date_trunc('hour', cpms.created_at)").
		Order("date_trunc('hour', cpms.created_at) DESC").
		Scan(&rows)

	dates := chart.Hours(start, end)
	index := chart.Index(rows)
	metrics := chart.Metrics(dates, index)

	return Data{
		Metrics: metrics,
		Data:    dates,
	}
}
