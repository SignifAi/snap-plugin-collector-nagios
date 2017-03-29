package nagios

import (
	/*	"encoding/json" */
	"fmt"
	/*
		"io"
		"os" */

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
)

const (
	Name      = "Nagios"
	Namespace = "nagios"
	Version   = 1
)

type NagiosPlugin struct {
	statusFile string
}

func (NagiosPlugin) GetConfigPolicy() (*plugin.ConfigPolicy, error) {
	policy := plugin.NewConfigPolicy()
	policy.AddNewStringRule([]string{"nagios"},
		"status_file",
		true)

	return policy, nil
}

func (nagios *NagiosPlugin) GetMetricTypes(pluginConfig plugin.Config) ([]plugin.Metric, error) {
	metricDefinitions := []plugin.Metric{}

	// Host State
	metricDefinitions = append(metricDefinitions, plugin.Metric{
		Description: "A host's state (supercedes service state)",
		Namespace:   plugin.NewNamespace("nagios").AddDynamicElement("hostname", "The hostname for the service").AddStaticElement("state"),
		Unit:        "string",
		Version:     Version,
	})

	// Service State
	metricDefinitions = append(metricDefinitions, plugin.Metric{
		Description: "A service's state",
		Namespace:   plugin.NewNamespace("nagios").AddDynamicElement("hostname", "The hostname for the service").AddStaticElement("services").AddDynamicElement("service_name", "The service's name").AddStaticElement("state"),
		Unit:        "string",
		Version:     Version,
	})

	// Acknowledgement (host)
	metricDefinitions = append(metricDefinitions, plugin.Metric{
		Description: "A host's acknowledgment status",
		Namespace:   plugin.NewNamespace("nagios").AddDynamicElement("hostname", "The hostname being acknowledged").AddStaticElement("acknowledged"),
		Unit:        "boolean",
		Version:     Version,
	})

	// Acknowledgement (service)
	metricDefinitions = append(metricDefinitions, plugin.Metric{
		Description: "A service's acknowledgment status",
		Namespace:   plugin.NewNamespace("nagios").AddDynamicElement("hostname", "The hostname of the host for which the service is acknowledged").AddStaticElement("services").AddDynamicElement("service_name", "The service's name").AddStaticElement("acknowledged"),
		Unit:        "boolean",
		Version:     Version,
	})
	return metricDefinitions, nil
}

func matchMetric(metric plugin.Metric, hostname string, service string, acked bool) bool {
	// It's outright invalid to match host against host:service and vice versa
	if service == "" {
		if metric.Namespace[2].Value == "services" {
			return false
		}
	} else {
		if metric.Namespace[2].Value == "state" {
			return false
		}
	}

	var matches bool = false

	for namePartIndex := range metric.Namespace {
		namePart := metric.Namespace[namePartIndex]
		if namePart.IsDynamic() {
			if namePart.Name == "hostname" && (namePart.Value == hostname || namePart.Value == "*") {
				matches = true
			} else if namePart.Name == "service_name" && (namePart.Value == service || namePart.Value == "*") {
				matches = true
			} else if namePart.Name == "service_name" && (namePart.Value != service && namePart.Value != "*") {
				// i.e. if the host matches and the service does not
				matches = false
			} else if namePart.Name == "acknowledged" && !acked {
				// i.e. this is an acked metric but we want to match a state one
				matches = false
			} else if namePart.Name == "state" && acked {
				// i.e. this is a state metric but we want to match an acked one
				matches = false
			}
		}
	}

	return matches
}

func (nagios *NagiosPlugin) CollectMetrics(metrics []plugin.Metric) ([]plugin.Metric, error) {
	var returnedMetrics []plugin.Metric
	for metricIndex := range metrics {
		metric := metrics[metricIndex]
		fmt.Println("idk", metric.Namespace.Strings())
	}
	return returnedMetrics, nil
}
