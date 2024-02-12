package metrics

const domainPrefix = "bux_"

const (
	verifyMerkleRootsHistogramName = domainPrefix + "verify_merkle_roots_histogram"
	recordTransactionHistogramName = domainPrefix + "record_transaction_histogram"
	queryTransactionHistogramName  = domainPrefix + "query_transaction_histogram"
)

const (
	cronHistogramName          = domainPrefix + "cron_histogram"
	cronLastExecutionGaugeName = domainPrefix + "cron_last_execution_gauge"
)

const (
	statsGaugeName = domainPrefix + "stats_total"
)
