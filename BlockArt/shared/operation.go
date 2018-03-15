package shared

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
)

type ShapeType int

const (
	// Path shape.
	PATH ShapeType = iota
	CIRCLE
	// Circle shape (extra credit).

)

type Operation struct {
	Command        string
	AmountOfInk    uint32
	Shapetype      ShapeType
	ShapeSvgString string
	Fill           string
	Stroke         string
	Issuer         *ecdsa.PrivateKey
	IssuerR        *big.Int
	IssuerS        *big.Int
	ValidFBlkNum   uint8
	Opid           uint32
	Draw           bool
}

func (o *Operation) CheckInk() bool {
	return false
}

func (o *Operation) CheckIntersection() bool {
	return false
}

func (o *Operation) CheckDuplicateSignature() bool {
	return false
}

func (o *Operation) CheckDeletedShapeExist() bool {
	return false
}

func (o *Operation) Validate() bool {
	// Check that each operation has sufficient ink associated with the public key that generated the operation.
	// Check that each operation does not violate the shape intersection policy described above.
	// Check that the operation with an identical signature has not been previously added to the blockchain (prevents operation replay attacks).
	// Check that an operation that deletes a shape refers to a shape that exists and which has not been previously deleted.
	return false
}

func (o *Operation) CheckIssuerSig() bool {
	if (o.Issuer == nil) || ((o.IssuerR == nil) || (o.IssuerS == nil)) {
		fmt.Println("------------------------------------------------They are all empty")
		return false
	}
	return ecdsa.Verify(&o.Issuer.PublicKey, []byte(o.Command), o.IssuerR, o.IssuerS)
}
