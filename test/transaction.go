package main

import (
	"math/big"
	"time"

	"github.com/axiomesh/axiom-kit/types"
)

const to = "0xc7F999b83Af6DF9e67d0a37Ee7e900bF38b3D013"

func MakeTransactions(count int) []*types.Transaction {
	txs := make([]*types.Transaction, count)
	time.Sleep(10 * time.Millisecond)
	s, err := types.GenerateSigner()
	if err != nil {
		panic(err)
	}
	for i := 0; i < count; i++ {
		tx, err := types.GenerateTransactionWithSigner(uint64(i), types.NewAddressByStr(to), big.NewInt(1), nil, s)
		if err != nil {
			panic(err)
		}
		txs[i] = tx
	}
	return txs
}
