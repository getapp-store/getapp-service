package dashboards

import (
	"net/http"
)

type Tracker struct {
}

func NewTracker() *Tracker {
	return &Tracker{}
}

// https://blog.cubieserver.de/2020/how-to-render-standalone-html-snippets-with-go-echarts/
func (d *Tracker) Dashboard(w http.ResponseWriter, r *http.Request) {

	w.Write([]byte(`:)`))
}
