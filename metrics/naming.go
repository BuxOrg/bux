package metrics

const domainPrefix = "bux_"

const (
	verifyMerkleRootsHistogramName = domainPrefix + "verify_merkle_roots_histogram"
	recordTransactionHistogramName = domainPrefix + "record_transaction_histogram"
)

const xpubGaugeName = domainPrefix + "xpub_gauge"
