package adapter

import "github.com/prometheus/client_golang/prometheus"

const (
	metricNamespace = "cloudwatch"
)

var (
	putMetricAPICallsCount = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: metricNamespace,
		Name:      "put_metrics_api_call_count",
	})

	putMetricAPICallErrorsCount = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: metricNamespace,
		Name:      "put_metrics_api_call_errors_count",
	})

	putMetricAPICallsDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: metricNamespace,
		Name:      "put_metrics_duration",
	})

	metricsCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: metricNamespace,
		Name:      "metrics_count",
	})
)

func init() {
	prometheus.MustRegister(metricsCount)
	prometheus.MustRegister(putMetricAPICallsCount)
	prometheus.MustRegister(putMetricAPICallErrorsCount)
	prometheus.MustRegister(putMetricAPICallsDuration)
}
