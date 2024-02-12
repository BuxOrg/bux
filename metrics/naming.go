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
	xpubGaugeName           = domainPrefix + "xpub_gauge"
	utxoGaugeName           = domainPrefix + "utxo_gauge"
	transactionInGaugeName  = domainPrefix + "transaction_in_gauge"
	transactionOutGaugeName = domainPrefix + "transaction_out_gauge"
	paymailGaugeName        = domainPrefix + "paymail_gauge"
	destinationGaugeName    = domainPrefix + "destination_gauge"
	accessKeyGaugeName      = domainPrefix + "access_key_gauge"
)
