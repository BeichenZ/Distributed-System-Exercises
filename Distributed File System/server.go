package main

import (
	//"crypto/md5"
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

//Main Method
func main() {
	if len(os.Args) != 2 {
		fmt.Println("Invalid Number of Command Line Argument")
		os.Exit(1)
	}

	tcpAddr_Server,err := net.ResolveTCPAddr("tcp",os.Args[1])
	Check_ServerError(err)
	//Server's Overall Structure is referenced from https://coderwall.com/p/wohavg/creating-a-simple-tcp-server-in-go

	//Listen for incoming connections
	Iconn,err := net.Listen("tcp",tcpAddr_Server)
	Check_ServerError(err)

}

//Check for Server's error that leads to Server Shut-Down
func Check_ServerError(err error) {
	if err != nil {
		fmt.Println("Error Ocurred:", err)
		os.Exit(0)
	}
}
