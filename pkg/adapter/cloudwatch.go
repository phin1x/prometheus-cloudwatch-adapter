package adapter

import (
	"fmt"
	"math"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/davecgh/go-spew/spew"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/prompb"
)

var (
	ignoresLabels = map[string]interface{}{
		model.MetricsPathLabel: nil,
		"endpoint":             nil,
		"metrics_path":         nil,
		"id":                   nil,
	}
)

/*
Cloudwatch API Limits:
Each PutMetricData request is limited to 40 KB in size for HTTP POST requests. You can processWriteRequest a payload compressed by gzip. Each request is also limited to no more than 20 different metrics.
PutMetricData can handle 150 transactions per second (TPS), which is the maximum number of operation requests you can make per second without being throttled.
*/

func (a *adapter) processWriteRequest(req *prompb.WriteRequest) error {
	metrics := make([]*cloudwatch.MetricDatum, 0)
	for _, ts := range req.Timeseries {
		m, err := a.toMetricDatum(&ts)
		if err != nil {
			return err
		}
		metrics = append(metrics, m...)
	}

	if len(metrics) == 0 {
		return nil
	}

	inp := &cloudwatch.PutMetricDataInput{
		Namespace:  aws.String(a.config.CloudwatchNamespace),
		MetricData: make([]*cloudwatch.MetricDatum, 0, 20),
	}

	counter := 0
	for _, metricDatum := range metrics {
		inp.MetricData = append(inp.MetricData, metricDatum)

		counter++
		if counter == 20 {
			if err := a.sendToCloudWatch(inp); err != nil {
				return err
			}

			counter = 0
			inp.MetricData = make([]*cloudwatch.MetricDatum, 0, 20)
		}
	}

	if err := a.sendToCloudWatch(inp); err != nil {
		return err
	}

	totalMetrics := len(metrics)
	metricsCount.Set(float64(totalMetrics))
	if a.config.Debug {
		a.logger.Info(fmt.Sprintf("wrote %d samples", totalMetrics))
	}

	return nil
}

func (a *adapter) sendToCloudWatch(inp *cloudwatch.PutMetricDataInput) error {
	if len(inp.MetricData) == 0 {
		return nil
	}

	if a.config.Debug {
		a.logger.Info("put metrics to cloudwatch")
	}

	putMetricAPICallsCount.Inc()
	timer := prometheus.NewTimer(putMetricAPICallsDuration)
	defer timer.ObserveDuration()

	if _, err := a.cw.PutMetricData(inp); err != nil {
		putMetricAPICallErrorsCount.Inc()

		if a.config.Debug {
			if awsErr, ok := err.(awserr.Error); ok && awsErr.Message() != "" {
				a.logger.Info(awsErr.Message())
			}

			a.logger.Info(spew.Sdump(inp))
		}

		return err
	}

	return nil
}

func (a *adapter) toMetricDatum(ts *prompb.TimeSeries) ([]*cloudwatch.MetricDatum, error) {
	if len(ts.Labels) > 10 {
		return nil, nil
	}

	metricName := ""
	dimensions := make([]*cloudwatch.Dimension, 0, 10)
	for _, label := range ts.Labels {
		// skip POD containers, we don't want pause containers
		if label.Name == "container" && label.Value == "POD" {
			return nil, nil
		}

		if label.Name == model.MetricNameLabel {
			metricName = label.Value
			continue
		}

		if _, ok := ignoresLabels[label.Name]; ok {
			continue
		}

		d := &cloudwatch.Dimension{}
		d.SetName(label.Name)
		d.SetValue(label.Value)
		dimensions = append(dimensions, d)
	}

	if metricName == "" {
		return nil, nil
	}

	metrics := make([]*cloudwatch.MetricDatum, 0)
	for _, sample := range ts.Samples {
		if math.IsNaN(sample.Value) || math.IsInf(sample.Value, 0) {
			continue
		}

		datum := &cloudwatch.MetricDatum{}
		datum.SetMetricName(metricName)
		datum.SetDimensions(dimensions)
		datum.SetTimestamp(time.Unix(0, sample.Timestamp*1e6))
		datum.SetValue(sample.Value)
		metrics = append(metrics, datum)
	}

	return metrics, nil
}
