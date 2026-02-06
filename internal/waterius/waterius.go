package waterius

import (
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type DBData struct {
	Ch0     float64 `json:"ch0"`
	Ch1     float64 `json:"ch1"`
	Battery int     `json:"battery"`
}

type DB struct {
	mu   sync.RWMutex
	data map[string]*DBData
}

func NewDB() *DB {
	return &DB{
		data: make(map[string]*DBData),
	}
}

func (db *DB) Insert(room string, dbData *DBData) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.data[room] = dbData
}

func (db *DB) Sel(room string) *DBData {
	db.mu.RLock()
	defer db.mu.RUnlock()

	values, ok := db.data[room]
	if !ok {
		return &DBData{}
	}

	return &DBData{
		Ch0:     values.Ch0,
		Ch1:     values.Ch1,
		Battery: values.Battery,
	}
}

func (db *DB) SelAll() map[string]*DBData {
	db.mu.RLock()
	defer db.mu.RUnlock()

	cp := make(map[string]*DBData, len(db.data))
	maps.Copy(cp, db.data)

	return cp
}

type PrometheusCollector struct {
	db *DB
}

func NewPrometheusCollector(db *DB) *PrometheusCollector {
	return &PrometheusCollector{
		db: db,
	}
}

func (c *PrometheusCollector) Describe(ch chan<- *prometheus.Desc) {
	for room, _ := range c.db.SelAll() {
		nameCh0 := fmt.Sprintf("%s_ch0", room)
		nameCh1 := fmt.Sprintf("%s_ch1", room)
		nameBattery := fmt.Sprintf("%s_battery", room)

		ch <- prometheus.NewDesc(nameCh0, nameCh0, nil, nil)
		ch <- prometheus.NewDesc(nameCh1, nameCh1, nil, nil)
		ch <- prometheus.NewDesc(nameBattery, nameBattery, nil, nil)
	}
}

func (c *PrometheusCollector) Collect(ch chan<- prometheus.Metric) {
	for room, values := range c.db.SelAll() {

		nameCh0 := fmt.Sprintf("%s_ch0", room)
		nameCh1 := fmt.Sprintf("%s_ch1", room)
		nameBattery := fmt.Sprintf("%s_battery", room)

		ch0 := prometheus.NewMetricWithTimestamp(
			time.Now(),
			prometheus.MustNewConstMetric(
				prometheus.NewDesc(nameCh0, nameCh0, nil, nil),
				prometheus.GaugeValue,
				values.Ch0,
			),
		)

		ch1 := prometheus.NewMetricWithTimestamp(
			time.Now(),
			prometheus.MustNewConstMetric(
				prometheus.NewDesc(nameCh1, nameCh1, nil, nil),
				prometheus.GaugeValue,
				values.Ch1,
			),
		)

		battery := prometheus.NewMetricWithTimestamp(
			time.Now(),
			prometheus.MustNewConstMetric(
				prometheus.NewDesc(nameBattery, nameBattery, nil, nil),
				prometheus.GaugeValue,
				float64(values.Battery),
			),
		)

		ch <- ch0
		ch <- ch1
		ch <- battery
	}
}

type HTTPHandler struct {
	db *DB
}

func NewHTTPHandler(db *DB) *HTTPHandler {
	return &HTTPHandler{
		db: db,
	}
}
func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.Split(r.URL.Path, "/")
	if len(path) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("invalid path"))

		return
	}

	room := path[len(path)-1]

	decoder := json.NewDecoder(r.Body)
	defer func() {
		_ = r.Body.Close()
	}()

	data := struct {
		Ch0     float64 `json:"ch0"`
		Ch1     float64 `json:"ch1"`
		Battery int     `json:"battery"`
	}{}

	err := decoder.Decode(&data)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(fmt.Sprintf("error decoding body: %s", err)))

		return
	}

	h.db.Insert(room, &DBData{
		Ch0:     data.Ch0,
		Ch1:     data.Ch1,
		Battery: data.Battery,
	})

	writtenData := h.db.Sel(room)
	writtenJSONData, _ := json.Marshal(writtenData)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(writtenJSONData)
}
