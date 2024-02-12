/*
Package metrics provides a way to track metrics in the application. Functionality is strictly tailored to the needs of the package and is not meant to be a general purpose metrics library.
*/
package metrics

import "time"

// Metrics is a struct that contains all the metrics that are used to track in the package
type Metrics struct {
	collector         Collector
	Stats             Stats
	verifyMerkleRoots HistogramVecInterface
	recordTransaction HistogramVecInterface
	queryTransaction  HistogramVecInterface
}

// NewMetrics is a constructor for the Metrics struct
func NewMetrics(collector Collector) *Metrics {
	return &Metrics{
		collector:         collector,
		Stats:             registerStats(collector),
		verifyMerkleRoots: collector.RegisterHistogramVec(verifyMerkleRootsHistogramName, "classification"),
		recordTransaction: collector.RegisterHistogramVec(recordTransactionHistogramName, "classification", "strategy"),
		queryTransaction:  collector.RegisterHistogramVec(queryTransactionHistogramName, "classification"),
	}
}

// EndWithClassification is a function returned by Track* methods that should be called when the tracked operation is finished
type EndWithClassification func(success bool)

// TrackVerifyMerkleRoots is used to track the time it takes to verify merkle roots
func (m *Metrics) TrackVerifyMerkleRoots() EndWithClassification {
	start := time.Now()
	return func(success bool) {
		m.verifyMerkleRoots.WithLabelValues(classify(success)).Observe(time.Since(start).Seconds())
	}
}

// TrackRecordTransaction is used to track the time it takes to record a transaction
func (m *Metrics) TrackRecordTransaction(strategyName string) EndWithClassification {
	start := time.Now()
	return func(success bool) {
		m.verifyMerkleRoots.WithLabelValues(classify(success), strategyName).Observe(time.Since(start).Seconds())
	}
}

// TrackQueryTransaction is used to track the time it takes to query a transaction
func (m *Metrics) TrackQueryTransaction() EndWithClassification {
	start := time.Now()
	return func(success bool) {
		m.verifyMerkleRoots.WithLabelValues(classify(success)).Observe(time.Since(start).Seconds())
	}
}

func classify(success bool) string {
	if success {
		return "success"
	}
	return "failure"
}
