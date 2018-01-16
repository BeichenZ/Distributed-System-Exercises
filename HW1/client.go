/*
Implements the solution to assignment 1 for UBC CS 416 2017 W2.

Usage:
$ go run client.go [local UDP ip:port] [local TCP ip:port] [aserver UDP ip:port]

Example:
$ go run client.go 127.0.0.1:2020 127.0.0.1:3030 127.0.0.1:7070

*/

package main

import (
	"crypto/md5"
	//"encoding/hex"

	// TODO
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"time"
)

/////////// Msgs used by both auth and fortune servers:

// An error message from the server.
type ErrMessage struct {
	Error string
}

/////////// Auth server msgs:

// Message containing a nonce from auth-server.
type NonceMessage struct {
	Nonce string
	N     int64 // PoW difficulty: number of zeroes expected at end of md5(nonce+secret)
}

// Message containing an the secret value from client to auth-server.
type SecretMessage struct {
	Secret string
}

// Message with details for contacting the fortune-server.
type FortuneInfoMessage struct {
	FortuneServer string // TCP ip:port for contacting the fserver

	FortuneNonce int64
}

/////////// Fortune server msgs:

// Message requesting a fortune from the fortune-server.
type FortuneReqMessage struct {
	FortuneNonce int64
}

// Response from the fortune-server containing the fortune.
type FortuneMessage struct {
	Fortune string
	Rank    int64 // Rank of this client solution
}

// Main workhorse method.
func main() {

	//Please Note: As a beginner of Golang, i read lots of reference
	//and record them in the comment section for future reference
	//Please Kindly Ignore

	//Parse Command line Input
	if len(os.Args) != 4 {
		fmt.Println("Invalid Number of Command line Arguments")
		os.Exit(1)
	}
	udpAddr_Local, err := net.ResolveUDPAddr("udp", os.Args[1])
	CheckError(err)
	tcpAddr_Local, err := net.ResolveTCPAddr("tcp", os.Args[2])
	//CheckError(err)
	udpAddr_Aserver, err := net.ResolveUDPAddr("udp", os.Args[3])
	CheckError(err)

	//Create Useful Data Structure
	var nMsg NonceMessage
	var fiMsg FortuneInfoMessage
	var powMsg SecretMessage
	var frMsg FortuneReqMessage
	var ftMsg FortuneMessage

	//establish UDP connection
	udp_Conn, err := net.DialUDP("udp", udpAddr_Local, udpAddr_Aserver)
	CheckError(err)
	defer udp_Conn.Close()
	CheckError(err)

	//Write an aribitary message to UDP Server
	randomMsg := make([]byte, 1024)
	_, err = udp_Conn.Write(randomMsg)
	CheckError(err)

	//Receive Message From Server
	n, _, err := udp_Conn.ReadFromUDP(randomMsg)
	CheckError(err)

	//Decode JSON Object
	err = json.Unmarshal(randomMsg[:n], &nMsg)
	CheckError(err)

	//Calculate the Secret
	secret := computeNonce(nMsg.N, nMsg.Nonce)
	powMsg.Secret = secret
	encoded_Secret, err := json.Marshal(powMsg)
	CheckError(err)

	_, err = udp_Conn.Write(encoded_Secret)

	n, _, err = udp_Conn.ReadFromUDP(randomMsg)
	CheckError(err)
	err = json.Unmarshal(randomMsg[:n], &fiMsg)
	CheckError(err)

	//Send to F-Server
	tcpAddr_Fserver, err := net.ResolveTCPAddr("tcp", fiMsg.FortuneServer)
	CheckError(err)
	tcp_Conn, err := net.DialTCP("tcp", tcpAddr_Local, tcpAddr_Fserver)
	CheckError(err)
	defer tcp_Conn.Close()

	frMsg.FortuneNonce = fiMsg.FortuneNonce
	encoded_FortuneNonce, err := json.Marshal(frMsg)
	CheckError(err)
	_, err = tcp_Conn.Write(encoded_FortuneNonce)
	CheckError(err)

	n, err = tcp_Conn.Read(randomMsg)
	CheckError(err)
	err = json.Unmarshal(randomMsg[:n], &ftMsg)
	CheckError(err)
	log.Println(ftMsg.Fortune)

}

func CheckError(err error) {
	if err != nil {
		fmt.Println("Error Ocurred:", err)
		os.Exit(0)
	}
}

//Cacluate the secret string
func computeNonce(N int64, Nonce string) string {
	var secretTemp []byte
	rand.Seed(int64(time.Now().Nanosecond()))

	//Trial Generate Random String:
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	secretBArray := make([]rune, 16)
	for {
		for i := range secretBArray {
			secretBArray[i] = letters[rand.Intn(62)]
		}

		secretTemp = computeNonceSecretHash(Nonce, string(secretBArray))
		if Check_ifNZeros(N, secretTemp) {
			return string(secretBArray)
		}
	}
}
func Check_ifNZeros(N int64, checksum []byte) bool {
	var Nb int64 = N / 2
	var i int64
	//str := hex.EncodeToString(checksum)
	if N%2 == 0 {
		for i = 0; i < Nb; i++ {
			if checksum[15-i] != 0 {
				return false
			}
		}
	} else {
		for i = 0; i < Nb; i++ {
			if checksum[15-i] != 0 {
				return false
			}
		}
		if N == 0 {
			if checksum[15]<<4 != 0 {
				return false
			}
		} else if checksum[15-Nb]<<4 != 0 {
			return false
		}

	}
	return true

}
func computeNonceSecretHash(nonce string, secret string) []byte {
	h := md5.New()
	h.Write([]byte(nonce + secret))
	return h.Sum(nil)
}
