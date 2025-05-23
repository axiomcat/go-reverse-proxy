package metrics

import (
	"fmt"
	"sync"

	"github.com/axiomcat/reverse-proxy/logger"
)

type Metrics struct {
	RequestTimes        []int64
	RequestCount        int
	MetricsRequestCount int
}

var instance *Metrics
var once sync.Once

func CreateInstance() {
	logger := logger.GetInstance(0)
	once.Do(func() {
		instance = &Metrics{}
	})
	logger.Debug("Metrics instance created")
}

func GetInstance() *Metrics {
	return instance
}

func (m *Metrics) getRequestTimeAvg() int64 {
	if len(m.RequestTimes) == 0 {
		return 0.0
	}
	var sum int64
	for _, val := range m.RequestTimes {
		sum += val
	}
	return sum / int64(len(m.RequestTimes))
}

func (m *Metrics) GetMetrics() string {
	m.MetricsRequestCount += 1
	metrics := ""
	metrics += fmt.Sprintf("http_request_time_avg_milliseconds %d\n", m.getRequestTimeAvg())
	metrics += fmt.Sprintf("http_request_total_count %d\n", m.RequestCount)
	metrics += fmt.Sprintf("metrics_request_count %d\n", m.MetricsRequestCount)
	return metrics
}
