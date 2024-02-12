package metrics

// Collector is an interface that is used to register metrics
type Collector interface {
	RegisterGauge(name string) GaugeInterface
	RegisterGaugeVec(name string, labels ...string) GaugeVecInterface
	RegisterHistogramVec(name string, labels ...string) HistogramVecInterface
}

type GaugeVecInterface interface {
	WithLabelValues(lvs ...string) GaugeInterface
}

// GaugeInterface is an interface that is used to track gauges of values
type GaugeInterface interface {
	Set(value float64)
}

// HistogramVecInterface is an interface that is used to register histograms with labels
type HistogramVecInterface interface {
	WithLabelValues(lvs ...string) HistogramInterface
}

// HistogramInterface is an interface that is used to track histograms of values
type HistogramInterface interface {
	Observe(value float64)
}
