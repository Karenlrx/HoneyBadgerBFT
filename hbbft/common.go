package hbbft

import (
	"crypto/sha256"

	"github.com/axiomesh/axiom-kit/types"
)

type Config struct {
	// 节点数
	N int
	// 恶意节点数
	F int
	// 节点标识
	ID uint64
	// 所有节点的标识
	Nodes []uint64
	// 每个epoch提交的最大交易数量
	BatchSize int
}

// hb
type HBMessage struct {
	Epoch   uint64
	Payload interface{}
}

// rbc
type BroadcastMessage struct {
	Payload interface{}
}

// bba
type AgreementMessage struct {
	Epoch   uint64
	Message interface{}
}

func CalculateHash(data []byte) string {
	h := sha256.Sum256(data)
	return types.NewHash(h[:]).String()
}
