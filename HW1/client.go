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
	"encoding/json"
	"fmt"
	"net"
	"os"
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

	fmt.Println("Local_UDP,Local_TCP,Target_UDP are", udpAddr_Local, tcpAddr_Local, udpAddr_Aserver)

	//establish UDP connection
	udp_Conn, err := net.DialUDP("udp", udpAddr_Local, udpAddr_Aserver)
	fmt.Println("Finish Dialing up")
	//udp_Conn,err := net.DialUDP("udp",os.Args[1],os.Args[3])
	CheckError(err)


	//Write an aribitary message to UDP Server
	randomMsg := make([]byte, 1024)
	_, err = udp_Conn.Write(randomMsg)
	CheckError(err)

	//Receive Message From Server
	fmt.Println("i am runnign")
	n, _, err := udp_Conn.ReadFromUDP(randomMsg)
	CheckError(err)

	//Decode JSON Object
	//Reference:https://gobyexample.com/json
	//Reference:invalid character '\x00' after top-level value
	err = json.Unmarshal(randomMsg[:n], &nMsg)
	CheckError(err)
	udp_Conn.Close()
	fmt.Println("Received From UDP Server : Display by default", randomMsg[:n])
	fmt.Println("Received From UDP Server : Display by default", nMsg.N)
	fmt.Println("Received From UDP Server : Display by default", nMsg.Nonce)

	//Calculate the Secret
	secret := computeNonce(nMsg.N, nMsg.Nonce) + nMsg.Nonce
	encoded_Secret,err := json.Marshal(secret)
	CheckError(err)

	udp_Conn, err = net.DialUDP("udp", udpAddr_Local, udpAddr_Aserver)
	CheckError(err)
	_,err = udp_Conn.Write(encoded_Secret)
	fmt.Println("i am waiting")

	randomMsg2 := make([]byte,1024)
	n, _, err = udp_Conn.ReadFromUDP(randomMsg2)
	CheckError(err)
	err = json.Unmarshal(randomMsg2[:n], &fiMsg)
	CheckError(err)
	fmt.Println("Received From UDP Server : Display by default", randomMsg2[:n])
	fmt.Println("Received From UDP Server : Display by default", fiMsg.FortuneServer)
	fmt.Println("Received From UDP Server : Display by default", fiMsg.FortuneNonce)




	//Encode Secret to Json and Send To UDP
	
	

}

func CheckError(err error) {
	if err != nil {
		fmt.Println("Error Ocurred:", err)
		os.Exit(0)
	}
}

//Cacluate the secret string
func computeNonce(N int64, Nonce string) string {
	//Insanity Check
	var md5CheckSum []byte
	md5CheckSum=computeNonceSecretHash("here-be-your-nonce","FVVTErKnJq")
	fmt.Println("Does this checksum satisfy requirment: ",Check_ifNZeros(N,md5CheckSum))
	return "FVVTErKnJq"

}

// Returns the MD5 hash as a hex string for the (nonce + secret) value.
func Check_ifNZeros(N int64,checksum []byte) bool{
	var Nb int64 = N/2
	fmt.Println("Nb is",Nb)
	var i int64
	if(N%2==0){
		for i=0;i<Nb;i++ {
			if(checksum[15-i] != 0){
				fmt.Println("ït's 1")
				return false
			}
		}
	} else{
		for i=0;i<Nb-1;i++{
			if(checksum[15-i] != 0){
				fmt.Println("it's 2")				
                                return false
                        }
		}
		if(N == 0){
			if(checksum[15]<<4 !=0){
			fmt.Println("it's 3")
			return false
			}
		} else if (checksum[15-Nb]<<4 != 0){
				fmt.Println("it's 4")
                                return false
                        }
		
	}
	return true
	
}
func computeNonceSecretHash(nonce string, secret string) []byte {
	h := md5.New()
	h.Write([]byte(nonce + secret))
	str := hex.EncodeToString(h.Sum(nil))
	fmt.Println("CheckSum is",str) 
	return h.Sum(nil)
}
