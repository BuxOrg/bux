package metrics

import "github.com/prometheus/client_golang/prometheus"

// Stats is a struct that contains all the gauges that are used to track the calculated stats of the application
type Stats struct {
	XPub        prometheus.Gauge
	Utxo        prometheus.Gauge
	Paymail     prometheus.Gauge
	Destination prometheus.Gauge
	AccessKey   prometheus.Gauge
}

func registerStats(collector Collector) Stats {
	return Stats{
		XPub:        collector.RegisterGauge(xpubGaugeName),
		Utxo:        collector.RegisterGauge(utxoGaugeName),
		Paymail:     collector.RegisterGauge(paymailGaugeName),
		Destination: collector.RegisterGauge(destinationGaugeName),
		AccessKey:   collector.RegisterGauge(accessKeyGaugeName),
	}
}
