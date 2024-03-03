package chart

import "time"

type Item struct {
	Name  string
	Id    int
	Date  time.Time
	Value float64
}

type Metric struct {
	Name string
	Data []float64
}

type Indx = map[string]map[time.Time]float64
type Dtx = []time.Time
type Mtx = []Metric

func Dates(start, end time.Time) []time.Time {
	dates := []time.Time{}
	for d := start; d.After(end) == false; d = d.AddDate(0, 0, 1) {
		dates = append(dates, time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC))
	}

	return dates
}

func Hours(start, end time.Time) []time.Time {
	dates := []time.Time{}
	for d := start; d.After(end) == false; d = d.Add(time.Hour) {
		dates = append(dates, time.Date(d.Year(), d.Month(), d.Day(), d.Hour(), 0, 0, 0, time.UTC))
	}

	return dates
}

func Index(items []Item) Indx {
	index := Indx{}
	for _, r := range items {
		t, ok := index[r.Name]
		if !ok {
			t = map[time.Time]float64{}
		}

		t[r.Date] = r.Value
		index[r.Name] = t
	}

	return index
}

func Metrics(dates Dtx, index Indx) Mtx {
	metrics := Mtx{}
	for k, v := range index {
		tracker := Metric{
			Name: k,
			Data: []float64{},
		}

		for _, d := range dates {
			val := v[d]

			tracker.Data = append(tracker.Data, val)
		}

		metrics = append(metrics, tracker)
	}

	return metrics
}
