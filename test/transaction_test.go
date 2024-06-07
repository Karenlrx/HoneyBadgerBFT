package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"testing"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/klauspost/reedsolomon"
	"github.com/stretchr/testify/require"
)

func TestName(t *testing.T) {
	txs := MakeTransactions(10)
	data, err := types.MarshalTransactions(txs)
	require.Nil(t, err)

	txs2, err := types.UnmarshalTransactions(data)
	require.Nil(t, err)
	require.Equal(t, len(txs), len(txs2))
	for _, transaction := range txs2 {
		fmt.Printf("receive tx[nonce:%d]\n", transaction.RbftGetNonce())
	}

	h := sha256.Sum256(data)
	hash := types.NewHash(h[:]).String()
	println("original hash:", hash)
	outData := Encoder(data)
	txs3, err := types.UnmarshalTransactions(outData)
	require.Nil(t, err)
	require.Equal(t, len(txs), len(txs3))
	for _, transaction := range txs3 {
		fmt.Printf("after receive tx[nonce:%d]\n", transaction.RbftGetNonce())
	}

}

func Encoder(data []byte) []byte {
	// Create an encoder with 17 data and 3 parity slices.
	enc, _ := reedsolomon.New(2, 2)

	// Split the data into shards
	shards, _ := enc.Split(data)

	// Encode the parity set
	_ = enc.Encode(shards)

	// Verify the parity set
	ok, _ := enc.Verify(shards)
	if ok {
		fmt.Println("ok")
	}

	// Delete two shards
	shards[0], shards[3] = nil, nil

	// Reconstruct the shards
	_ = enc.Reconstruct(shards)

	// Verify the data set
	ok, _ = enc.Verify(shards)
	if ok {
		fmt.Println("Reconstruct ok")
	}
	// Output: ok
	// ok

	out := new(bytes.Buffer)
	err := enc.Join(out, shards, len(data))
	if err != nil {
		fmt.Println(err)
	}
	return out.Bytes()
}
