package nagios

import (
	/*	"encoding/json" */
	//"fmt"
	/*
		"io"
		"os" */
	"errors"
	"os"
	"time"
	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
)

const (
	Name      = "Nagios"
	Namespace = "nagios"
	Version   = 1
)

var (
	HostStateCode2String = map[string]string{
		"0": "UP",
		"1": "DOWN",
	}
	ServiceStateCode2String = map[string]string{
		"0": "OK",
		"1": "WARNING",
		"2": "CRITICAL",
		"3": "UNKNOWN",
	}
)

type NagiosPlugin struct {
	statusFile string
}

func (NagiosPlugin) GetConfigPolicy() (plugin.ConfigPolicy, error) {
	policy := plugin.NewConfigPolicy()
	policy.AddNewStringRule([]string{"nagios"},
		"status_file",
		true)

	return *policy, nil
}

func (nagios NagiosPlugin) GetMetricTypes(pluginConfig plugin.Config) ([]plugin.Metric, error) {
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

func HostStatusToMetric(hostname string, valueOf string, status map[string]string) (plugin.Metric, error) {
	var metricValue interface{}
	switch valueOf {
	case "state":
		var stateVar string
		if status["state_type"] == "0" {
			// Soft -- use last_hard_state to avoid flapping too much...
			stateVar = "last_hard_state"
		} else {
			stateVar = "current_state"
		}
		metricValue = HostStateCode2String[status[stateVar]]
	case "acknowledged":
		metricValue = status["problem_has_been_acknowledged"]
		if metricValue.(string) == "0" {
			metricValue = false
		} else  {
			metricValue = true
		}
	}

	metricName := plugin.NewNamespace("nagios").AddDynamicElement("hostname", "The hostname for the service").AddStaticElement(valueOf)	
	metricName[1].Value = hostname
	return plugin.Metric{
		Namespace: metricName,
		Data: metricValue,
		Timestamp: time.Now(),
		Version: Version,
	}, nil
}

func HostServiceStatusToMetric(hostname string, service string, valueOf string, status map[string]string) (plugin.Metric, error) {
	var metricValue interface{}
	switch valueOf {
	case "state":
		var stateVar string
		if status["state_type"] == "0" {
			// Soft -- use last_hard_state to avoid flapping too much...
			stateVar = "last_hard_state"
		} else {
			stateVar = "current_state"
		}
		metricValue = ServiceStateCode2String[status[stateVar]]
	case "acknowledged":
		metricValue = status["problem_has_been_acknowledged"]
		if metricValue.(string) == "0" {
			metricValue = false
		} else {
			metricValue = true
		}
	}

	metricName := plugin.NewNamespace("nagios").AddDynamicElement("hostname", "The hostname for the service").AddStaticElement("services").AddDynamicElement("service_name", "The service's name").AddStaticElement(valueOf)
	metricName[1].Value = hostname
	metricName[3].Value = service
	return plugin.Metric{
		Namespace: metricName,
		Data: metricValue,
		Timestamp: time.Now(),
		Version: Version,
	}, nil
}

func (nagios NagiosPlugin) CollectMetrics(metrics []plugin.Metric) (returnedMetrics []plugin.Metric, err error) {
	statusFilename, err := metrics[0].Config.GetString("status_file")
	statusFile, err := os.Open(statusFilename)
	hoststatuses, servicestatuses, err := NagiosStatusMaps(statusFile)
	if err == nil {
		for _, metric := range metrics {
			var hostname, serviceName, valueOf string
			var isService bool = false
			for _, namePart := range metric.Namespace {
				if namePart.IsDynamic() {
					switch namePart.Name {
					case "hostname":
						hostname = namePart.Value
					case "service_name":
						serviceName = namePart.Value
						isService = true
					}
				} else {
					switch namePart.Value {
					case "acknowledged":
						valueOf = "acknowledged"
					case "state":
						valueOf = "state"
					}
				}
			}

			if valueOf != "" {
				if isService {
					if hostname == "*" {
						for _hostname, serviceMap := range servicestatuses {
							if serviceName == "*" {
								for _serviceName, serviceData := range serviceMap {
									newMetric, err := HostServiceStatusToMetric(_hostname, _serviceName, valueOf, serviceData)
									if err == nil {
										returnedMetrics = append(returnedMetrics, newMetric)
									}
								}
							} else {
								if serviceData, ok := serviceMap[serviceName]; ok {
									newMetric, err := HostServiceStatusToMetric(_hostname, serviceName, valueOf, serviceData)
									if err == nil {
										returnedMetrics = append(returnedMetrics, newMetric)
									}
								}
							}
						}
					} else {
						if serviceName == "*" {
							for _serviceName, serviceData := range servicestatuses[hostname] {
								newMetric, err := HostServiceStatusToMetric(hostname, _serviceName, valueOf, serviceData)
								if err == nil {
									returnedMetrics = append(returnedMetrics, newMetric)
								}
							}
						} else {
							newMetric, err := HostServiceStatusToMetric(hostname, serviceName, valueOf, servicestatuses[hostname][serviceName])
							if err == nil {
								returnedMetrics = append(returnedMetrics, newMetric)
							}
						}
					}
				} else {
					// TODO: Make more efficient? It's just funny I'm doing the same thing
					// both times, but one's in an iteration and the other is a one-off. 
					if hostname == "*" {
						for _hostname, hostMetrics := range hoststatuses {
							newMetric, err := HostStatusToMetric(_hostname, valueOf, hostMetrics)
							if err == nil {
								returnedMetrics = append(returnedMetrics, newMetric)
							}
						}
					} else {
						if hostMetrics, ok := hoststatuses[hostname]; ok {
							newMetric, err := HostStatusToMetric(hostname, valueOf, hostMetrics)
							if err == nil {
								returnedMetrics = append(returnedMetrics, newMetric)
							}
						}
					}
				}
			} else {
				err = errors.New("Missing valueOf for host [" + hostname + "], service [" + serviceName + "]")
				break
			}
		} 
	}
	return returnedMetrics, err
}
