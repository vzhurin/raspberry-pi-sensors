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

type DB struct {
	mu   sync.RWMutex
	data map[string][2]float64
}

func NewDB() *DB {
	return &DB{
		data: make(map[string][2]float64),
	}
}

func (db *DB) Insert(room string, ch0 float64, ch1 float64) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.data[room] = [...]float64{ch0, ch1}
}

func (db *DB) Sel(room string) [2]float64 {
	db.mu.RLock()
	defer db.mu.RUnlock()

	values, ok := db.data[room]
	if !ok {
		return [2]float64{0, 0}
	}

	return values
}

func (db *DB) SelAll() map[string][2]float64 {
	db.mu.RLock()
	defer db.mu.RUnlock()

	cp := make(map[string][2]float64, len(db.data))
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

		ch <- prometheus.NewDesc(nameCh0, nameCh0, nil, nil)
		ch <- prometheus.NewDesc(nameCh1, nameCh1, nil, nil)
	}
}

func (c *PrometheusCollector) Collect(ch chan<- prometheus.Metric) {
	for room, values := range c.db.SelAll() {

		nameCh0 := fmt.Sprintf("%s_ch0", room)
		nameCh1 := fmt.Sprintf("%s_ch1", room)

		ch0 := prometheus.NewMetricWithTimestamp(
			time.Now(),
			prometheus.MustNewConstMetric(
				prometheus.NewDesc(nameCh0, nameCh0, nil, nil),
				prometheus.GaugeValue,
				values[0],
			),
		)

		ch1 := prometheus.NewMetricWithTimestamp(
			time.Now(),
			prometheus.MustNewConstMetric(
				prometheus.NewDesc(nameCh1, nameCh1, nil, nil),
				prometheus.GaugeValue,
				values[1],
			),
		)

		ch <- ch0
		ch <- ch1
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
		CH0 float64 `json:"ch0"`
		CH1 float64 `json:"ch1"`
	}{}

	err := decoder.Decode(&data)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(fmt.Sprintf("error decoding body: %s", err)))

		return
	}

	h.db.Insert(room, data.CH0, data.CH1)

	writtenData := h.db.Sel(room)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(fmt.Sprintf("%v", writtenData)))
}
