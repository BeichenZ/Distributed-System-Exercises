package shared

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strconv"
)

type BlockApi interface {
	GetStringBlock() string
}

// type UserSignatureSturct struct {
// 	r *big.Int
// 	s *big.Int
// }

type Block struct {
	CurrentHash  string
	PreviousHash string
	// UserSignature     UserSignatureSturct
	R                 *big.Int
	S                 *big.Int
	CurrentOPs         []Operation
	Children          []*Block
	DistanceToGenesis int
	Nonce             int32
	SolverPublicKey   *ecdsa.PublicKey
	OpSig 				string
}

// Return a string repersentation of PreviousHash, op, op-signature, pub-key,
func (b Block) GetString() string {
	return b.PreviousHash + AllOperationsCommands(b.CurrentOPs) + pubKeyToString(*b.SolverPublicKey)
}

func (b *Block) checkMD5() bool {
	if computeNonceSecretHash(b.GetString(), strconv.FormatInt(int64(b.Nonce), 10)) == b.CurrentHash {
		return true
	} else {
		fmt.Println("MD5 VALIDATION FAILED")
	}
	return false
}

func (b *Block) checkSolversSigForBlock() bool {

	if ecdsa.Verify(b.SolverPublicKey, []byte(b.CurrentHash), b.R, b.S) {
		return true
	} else {
		fmt.Println("SolversSig VALIDATION FAILED")
	}

	return false
}

func (b *Block) checkIssuerSigForOperation() bool {
	if len(b.CurrentOPs) == 0 {
		return true
	}

	for _, operation := range b.CurrentOPs {
		fmt.Println(operation.Command)
		if !operation.CheckIssuerSig() {
			return false
		}
	}

	return true
}

func (b *Block) Validate() bool {
	// Check that the nonce for the block is valid: PoW is correct and has the right difficulty.
	// Check that each operation in the block has a valid signature (this signature should be generated using the private key and the operation).
	// Check that the previous block hash points to a legal, previously generated, block.
	return b.checkMD5() && b.checkSolversSigForBlock() && b.checkIssuerSigForOperation()
}

func (b *Block) getStringFromBigInt() string {
	return ((b.R).String()) + ((b.S).String())
}
