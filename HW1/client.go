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
	"encoding/hex"

	// TODO
	"fmt"
	"os"
	"net"
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
	FortuneNonce  int64
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
	//Parse Command line Input
	if len(os.Args) != 4 {
		fmt.Println("Invalid Number of Command line Arguments")
		os.Exit(1)
	}
	udpAddr_Local,err := net.ResolveUDPAddr("udp",os.Args[1])
	CheckError(err)
	tcpAddr_Local,err := net.ResolveTCPAddr("tcp",os.Args[2])
	//CheckError(err)
	udpAddr_Aserver,err := net.ResolveUDPAddr("udp",os.Args[3])
	CheckError(err)
	
	fmt.Println("Local_UDP,Local_TCP,Target_UDP are",udpAddr_Local,tcpAddr_Local,udpAddr_Aserver)
	
	//establish UDP connection
	udp_Conn,err := net.DialUDP("udp",udpAddr_Local,udpAddr_Aserver)
	CheckError(err)
	defer udp_Conn.Close()
	
	//Write an aribitary message to UDP Server
	randomMsg:= make([]byte,1024)
	_,err = udp_Conn.Write(randomMsg)
	CheckError(err)
	
	//Receive Message From Server
	 fmt.Println("i am runnign")
     n, addr, err := udp_Conn.ReadFromUDP(randomMsg)
	 CheckError(err)
     fmt.Println("UDP Server : ", addr)
     fmt.Println("Received from UDP server : ", string(randomMsg[:n]))
	/*for {
		fmt.Println("i am runnign")
		rxMsg := make([]byte,100)
		n,err := udp_Conn.Read(rxMsg)
		CheckError(err)
		fmt.Println("Received From A server:",string(rxMsg[:n]))
		
	}*/
	
	fmt.Println("This is Done")


	//fmt.Println("UDP Server:",addr)
	//fmt.Println("Received From A server:",string(buffer[:n]))
	
		
	// Use json.Marshal json.Unmarshal for encoding/decoding to servers

}

func CheckError(err error){
	if err != nil {
		fmt.Println("Error Ocurred:",err)
		os.Exit(0)
	}
}


// Returns the MD5 hash as a hex string for the (nonce + secret) value.
func computeNonceSecretHash(nonce string, secret string) string {
	h := md5.New()
	h.Write([]byte(nonce + secret))
	str := hex.EncodeToString(h.Sum(nil))
	return str
}
