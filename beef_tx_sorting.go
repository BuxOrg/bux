package bux

func kahnTopologicalSortTransaction(transactions []*Transaction) []*Transaction {
	// map to store transactions by their IDs for random access.
	transactionMap := make(map[string]*Transaction, len(transactions))

	inDegree := make(map[string]int)
	queue := make([]string, 0)
	result := make([]*Transaction, 0)

	for _, tx := range transactions {
		transactionMap[tx.ID] = tx
		inDegree[tx.ID] = 0
	}

	calculateDegrees(inDegree, transactions)
	fullfillZeroDegreeQueue(queue, inDegree)

	for len(queue) > 0 {
		tx := transactionMap[queue[0]]
		queue = queue[1:]
		result = append(result, tx)

		recalculateNeighbors(tx, inDegree, queue)
	}

	reverseInPlace(result)
	return result
}

func calculateDegrees(inDegree map[string]int, transactions []*Transaction) {
	for _, tx := range transactions {
		for _, input := range tx.draftTransaction.Configuration.Inputs {
			inDegree[input.UtxoPointer.TransactionID]++
		}
	}
}

func fullfillZeroDegreeQueue(queue []string, inDegree map[string]int) {
	for txID, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, txID)
		}
	}
}

func recalculateNeighbors(tx *Transaction, inDegree map[string]int, queue []string) {
	for _, input := range tx.draftTransaction.Configuration.Inputs {
		neighborID := input.UtxoPointer.TransactionID
		inDegree[neighborID]--

		if inDegree[neighborID] == 0 {
			queue = append(queue, neighborID)
		}
	}
}

func reverseInPlace(collection []*Transaction) {
	for i, j := 0, len(collection)-1; i < j; i, j = i+1, j-1 {
		collection[i], collection[j] = collection[j], collection[i]
	}
}
