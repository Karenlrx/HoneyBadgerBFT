package hbbft

import (
	"sync"

	"github.com/axiomesh/axiom-kit/types"
)

type transactionBuffer struct {
	lock sync.Mutex
	data []*types.Transaction
}

func newTransactionBuffer() *transactionBuffer {
	return &transactionBuffer{
		data: make([]*types.Transaction, 0, 1024*1024),
	}
}

func (b *transactionBuffer) add(transaction *types.Transaction) {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.data = append(b.data, transaction)
}

func (b *transactionBuffer) delete(transactions []*types.Transaction) {
	b.lock.Lock()
	defer b.lock.Unlock()

	temp := make(map[string]*types.Transaction)
	for i := 0; i < len(b.data); i++ {
		temp[b.data[i].RbftGetTxHash()] = b.data[i]
	}
	for i := 0; i < len(transactions); i++ {
		delete(temp, transactions[i].RbftGetTxHash())
	}
	data := make([]*types.Transaction, len(temp))
	i := 0
	for _, tx := range temp {
		data[i] = tx
		i++
	}
	b.data = data
}

func (b *transactionBuffer) len() int {
	b.lock.Lock()
	defer b.lock.Unlock()
	return len(b.data)
}
