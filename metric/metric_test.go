package metric

import (
	"github.com/armon/go-metrics"
	prom "github.com/armon/go-metrics/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"testing"
)

func TestNewPrometheusSinkFrom(t *testing.T) {
	reg := prometheus.NewRegistry()

	sink, err := prom.NewPrometheusSinkFrom(prom.PrometheusOpts{
		Registerer: reg,
	})

	if err != nil {
		t.Fatalf("err = %v, want nil", err)
	}

	//sink, _ := prometheus.NewPrometheusPushSink(
	//	"127.0.0.1:9090", time.Second, b.Logger().Name())
	metrics.NewGlobal(metrics.DefaultConfig("default"), sink)
	metrics.EmitKey([]string{"questions", "meaning of life"}, 42)
	metrics.SetGauge([]string{"one", "two"}, 42)
	metrics.AddSample([]string{"method", "wow"}, 42)
	metrics.AddSample([]string{"method", "wow"}, 100)
	metrics.AddSample([]string{"method", "wow"}, 22)

	dtos, err := reg.Gather()
	if err != nil {
		t.Fatal(err)
		return
	}
	for i, dto := range dtos {
		t.Log(i, dto)
	}
	//check if register has a sink by unregistering it.
	ok := reg.Unregister(sink)
	if !ok {
		t.Fatalf("Unregister(sink) = false, want true")
	}
}

