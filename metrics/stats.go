package metrics

// Stats is a struct that contains all the gauges that are used to track the calculated stats of the application
type Stats struct {
	XPub           GaugeInterface
	Utxo           GaugeInterface
	TransactionIn  GaugeInterface
	TransactionOut GaugeInterface
	Paymail        GaugeInterface
	Destination    GaugeInterface
	AccessKey      GaugeInterface
}

func registerStats(collector Collector) Stats {
	return Stats{
		XPub:           collector.RegisterGauge(xpubGaugeName),
		Utxo:           collector.RegisterGauge(utxoGaugeName),
		TransactionIn:  collector.RegisterGauge(transactionInGaugeName),
		TransactionOut: collector.RegisterGauge(transactionOutGaugeName),
		Paymail:        collector.RegisterGauge(paymailGaugeName),
		Destination:    collector.RegisterGauge(destinationGaugeName),
		AccessKey:      collector.RegisterGauge(accessKeyGaugeName),
	}
}
