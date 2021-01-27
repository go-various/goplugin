package metric

import (
	"github.com/armon/go-metrics"
	prom "github.com/armon/go-metrics/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"os"
	"time"
)

var (
	// Create a metrics registry.
	Registry          *prometheus.Registry
	PluginCountMetric *prometheus.CounterVec
	PluginGaugeMetric *prometheus.GaugeVec
)

func init() {

	Registry = prometheus.NewRegistry()
	PluginCountMetric = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "various",
		Subsystem: "plugin",
		Name:      "request",
	}, []string{"backend", "namespace", "operation"})

	PluginGaugeMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "various",
		Subsystem: "plugin",
		Name:      "delay",
	}, []string{"backend", "namespace", "operation"})

	// Register standard server metrics and customized metrics to registry.
	Registry.MustRegister(
		PluginCountMetric,
		PluginGaugeMetric,
	)
	sink, _ := prom.NewPrometheusSinkFrom(prom.PrometheusOpts{
		Registerer: Registry,
	})
	c := &metrics.Config{
		ServiceName:          "various", // Use client provided service
		HostName:             "",
		EnableHostname:       false,            // Enable hostname prefix
		EnableRuntimeMetrics: true,             // Enable runtime profiling
		EnableTypePrefix:     false,            // Disable type prefix
		TimerGranularity:     time.Millisecond, // Timers are in milliseconds
		ProfileInterval:      time.Second,      // Poll runtime every second
		FilterDefault:        true,             // Don't filter metrics by default
	}
	// Try to get the hostname
	name, _ := os.Hostname()
	c.HostName = name

	metrics.NewGlobal(c, sink)
}
