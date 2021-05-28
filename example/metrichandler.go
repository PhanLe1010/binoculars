package main

import (
	"github.com/sirupsen/logrus"

	"github.com/rancher/binoculars/client"
)

type MetricName string

const (
	MetricNameStorageCapacity = MetricName("storage_capacity")
	MetricNameEnabledFeature  = MetricName("enabled_feature")

	tagKeyInstanceUUID = "instance_uuid"
	instanceUUID       = "123e4567-e89b-12d3-a456-426614174000"
)

type MetricDefinition struct {
	Name   MetricName
	Tags   func() map[string]string
	Fields func() map[string]interface{}
}

var MetricDefinitions = map[MetricName]MetricDefinition{}

type MyMetricHandler struct {
	// fill in the fields needed to fetch data. I.g., Kubernetes client
}

func NewMyMetricHandler() *MyMetricHandler {
	h := &MyMetricHandler{}
	// Initialize MetricDefinitions map
	MetricDefinitions[MetricNameStorageCapacity] = MetricDefinition{
		Name:   MetricNameStorageCapacity,
		Tags:   h.GetTagsForStorageCapacityMetric,
		Fields: h.GetFieldsForStorageCapacityMetric,
	}
	MetricDefinitions[MetricNameEnabledFeature] = MetricDefinition{
		Name:   MetricNameEnabledFeature,
		Tags:   h.GetTagsForEnabledFeatureMetric,
		Fields: h.GetFieldsForEnabledFeatureMetric,
	}
	return h
}

func (h *MyMetricHandler) GatherMetrics() []client.Metric {
	var metrics []client.Metric

	for _, md := range MetricDefinitions {
		m := client.Metric{
			Name:   string(md.Name),
			Tags:   md.Tags(),
			Fields: md.Fields(),
		}
		metrics = append(metrics, m)
	}

	return metrics
}

func (h *MyMetricHandler) HandleResponse(response *client.ServerResponse, err error) {
	if err != nil {
		logrus.WithError(err).Error("failed to send metrics")
		return
	}
	logrus.Info("Successfully sent metrics")
}

func (h *MyMetricHandler) GetTagsForStorageCapacityMetric() map[string]string {
	return map[string]string{
		tagKeyInstanceUUID: instanceUUID,
	}
}

func (h *MyMetricHandler) GetFieldsForStorageCapacityMetric() map[string]interface{} {
	return map[string]interface{}{
		"value": 1000000000000,
	}
}

func (h *MyMetricHandler) GetTagsForEnabledFeatureMetric() map[string]string {
	return map[string]string{
		tagKeyInstanceUUID: instanceUUID,
	}
}

func (h *MyMetricHandler) GetFieldsForEnabledFeatureMetric() map[string]interface{} {
	return map[string]interface{}{
		"data_locality": 1,
		"auto_salvage":  1,
	}
}
