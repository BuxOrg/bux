package metrics

// Stats is a struct that contains all the gauges that are used to track the calculated stats of the application
type Stats struct {
	XPub GaugeInterface
}

func registerStats(collector Collector) Stats {
	return Stats{
		XPub: collector.RegisterGauge(xpubGaugeName),
	}
}
