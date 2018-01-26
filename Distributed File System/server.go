package main

import (
	//"crypto/md5"
	//"encoding/hex"

	// TODO
	"encoding/json"
	"fmt"
	//"log"
	//"math/rand"
	"net"
	"os"
	//"time"
)
type OneStringMsg struct {
	Msg string
}
//Main Method
func main() {
	if len(os.Args) != 2 {
		fmt.Println("Invalid Number of Command Line Argument")
		os.Exit(1)
	}

	tcpAddr_Server := os.Args[1]
	//Server's Overall Structure is referenced from https://coderwall.com/p/wohavg/creating-a-simple-tcp-server-in-go

	//Listen for incoming connections
	Iconn,err := net.Listen("tcp",tcpAddr_Server)
	Check_ServerError(err)
	defer Iconn.Close()
	fmt.Println("Listening on" + tcpAddr_Server)
	//main loop
	for{
		conn,err := Iconn.Accept()
		Check_ServerError(err)
		go handleRequest(conn)
	}

}
//Separate Thread to handle the request
func handleRequest(conn net.Conn){
	buf := make([]byte,1024)
	reqLen,err := conn.Read(buf)
	Check_NonFatalError(err)
	var receivedMsg OneStringMsg
	err = json.Unmarshal(buf[:reqLen],&receivedMsg)
	Check_NonFatalError(err)
	fmt.Println("message from client:",receivedMsg.Msg)
	conn.Write([]byte("Message Received."))
	conn.Close()
}

//Check for Server's error that leads to Server Shut-Down
func Check_ServerError(err error) {
	if err != nil {
		fmt.Println("Error Ocurred:", err)
		os.Exit(0)
	}
}
func Check_NonFatalError(err error){
	if err !=nil {
		fmt.Println("Non-Fatal Error:",err)	
	}
}
