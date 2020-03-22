package main

import "fmt"

func (cli *CLI) reindexURPO(nodeID string) {
	bc := NewBlockchain(nodeID)
	URPOSet := URPOSet{bc}
	URPOSet.Reindex()

	count := URPOSet.CountTransactions()
	fmt.Printf("Done! There are %d transactions in the URPO set.\n", count)
}
