package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
)

// Usage:
//
// $ go run keyCreation.go -nk <number of key pairs to create>

// create given number of keys and store them in a file as strings
func convKeyPairToString(kp *ecdsa.PrivateKey) (string, string) {
	privateKeyBytes, _ := x509.MarshalECPrivateKey(kp)
	privKeyString := hex.EncodeToString(privateKeyBytes)
	publicKeyBytes, _ := x509.MarshalPKIXPublicKey(kp.Public())
	pubKeyString := hex.EncodeToString(publicKeyBytes)

	return privKeyString, pubKeyString
}

func main() {
	numOfKeys := flag.Int("nk", 27, "an int")
	flag.Parse()
	if *numOfKeys != 0 {
		for i := 0; i < *numOfKeys; i++ {
			// Generate Keypair
			keyPair, _ := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)

			strPriv, strPub := convKeyPairToString(keyPair)
			fmt.Println("main ", strPriv, "\n", strPub)

			// Store string representation of key pair in file
			f, err := os.OpenFile("privateKey.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			CheckError(err)
			defer f.Close()
			_, err = f.WriteString("\n")
			_, err = f.WriteString("\n")
			_, err = f.WriteString(strPriv)
			CheckError(err)
			f.Sync()

			f, err = os.OpenFile("publicKey.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			CheckError(err)
			defer f.Close()
			_, err = f.WriteString("\n")
			_, err = f.WriteString("\n")
			_, err = f.WriteString(strPub)
			CheckError(err)
			f.Sync()

		}
	}
	return

}

func CheckError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
	}
}

