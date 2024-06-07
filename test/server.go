package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/liu-jianhao/HoneyBadgerBFT/hbbft"
)

type Server struct {
	id             uint64
	hb             *hbbft.HoneyBadger
	lock           sync.RWMutex
	transactionMap map[string]*types.Transaction
	totalCommit    int
	start          time.Time
}

func newServer(id uint64, nodes []uint64) *Server {
	hb, err := hbbft.NewHoneyBadger(hbbft.Config{
		N:         len(nodes),
		F:         len(nodes) / 4,
		ID:        id,
		Nodes:     nodes,
		BatchSize: batchSize,
	})
	if err != nil {
		panic(err)
	}
	return &Server{
		id:             id,
		hb:             hb,
		transactionMap: make(map[string]*types.Transaction),
		start:          time.Now(),
	}
}

func (s *Server) addTransactions(txx ...*types.Transaction) {
	for _, tx := range txx {
		s.lock.Lock()
		s.transactionMap[tx.RbftGetTxHash()] = tx
		s.lock.Unlock()

		s.hb.AddTransaction(tx)
	}
}

func (s *Server) addTransactionLoop() {
	timer := time.NewTicker(1000 * time.Millisecond)
	for {
		<-timer.C
		s.addTransactions(MakeTransactions(1024)...)
	}
}

func (s *Server) commitLoop() {
	timer := time.NewTicker(time.Second * 10)
	n := 0
	for {
		select {
		case <-timer.C:
			out := s.hb.Outputs()
			epochList := make([]uint64, 0)
			for e, txx := range out {
				for _, tx := range txx {
					//fmt.Printf("node%d reveive commit in epoch %d, tx[account: %s, nonce:%d]\n", s.id, e, tx.RbftGetFrom(), tx.RbftGetNonce())
					hash := tx.RbftGetTxHash()
					s.lock.Lock()
					n++
					delete(s.transactionMap, hash)
					s.lock.Unlock()
				}
				epochList = append(epochList, e)
			}

			var minEpoch uint64
			var maxEpoch uint64
			for i, e := range epochList {
				if i == 0 {
					minEpoch = e
					maxEpoch = e
				}
				if e <= minEpoch {
					minEpoch = e
				}
				if e > maxEpoch {
					maxEpoch = e
				}
			}
			epochInterval := maxEpoch - minEpoch

			s.totalCommit += n
			delta := time.Since(s.start)
			if s.id == 1 {
				fmt.Println("")
				fmt.Println("*************************************************")
				// fmt.Printf("server %d\n", s.id)
				fmt.Printf("commited %d transactions over %v\n", s.totalCommit, delta)
				fmt.Printf("throughput %d TX/s\n", s.totalCommit/int(delta.Seconds()))
				fmt.Printf("epoch %d - %d\n", minEpoch, maxEpoch)
				// fmt.Printf("epochInterval %d\n", epochInterval)
				fmt.Printf("latency %d ms\n", 5000/epochInterval)
				fmt.Println("*************************************************")
				fmt.Println("")
			}
			n = 0
		}
	}
}

func makeNetwork(n int) []*Server {
	nodes := make([]*Server, n)
	for i := 0; i < n; i++ {
		ids := make([]uint64, n)
		for i := 0; i < n; i++ {
			ids[i] = uint64(i)
		}
		nodes[i] = newServer(uint64(i), ids)
	}
	return nodes
}
