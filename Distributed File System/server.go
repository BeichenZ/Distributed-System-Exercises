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
	"net/rpc"
	//"./dfslib"
	"./dfslib/shared"
	"errors"
	//"time"
)
//Implement shared.DFSService Interface
type DFSServiceObj int
//Register a new client and return a
func (t *DFSServiceObj) RegisterNewClient(args *shared.RNCArgs,reply *shared.RNCReply) error{
	//Assumption by Assignment, Number of Clients is capped at 16. No client can be deleted
	availableIndex := -1
	for index,element := range clientList {
		if element.occupied == false {
			availableIndex = index
			break
		} 
	}
	if availableIndex == -1 {
		return errors.New("Application:Client Count is more than 16.")
	}
	clientList[availableIndex].localIP = args.LocalIP
	clientList[availableIndex].localPath = args.LocalPath
	clientList[availableIndex].occupied = true
	//Note use the (index+1) as the unique ID for each client
	clientList[availableIndex].ID = availableIndex+1
	(*reply).ID = availableIndex+1
	return nil
}

//Data Structure Type
type SingleClientInfo struct {
	occupied bool
	localIP string
	localPath string
	ID int
}
//Global Data Storage Shared by Multiple RPC calls and Main
var clientList [16]SingleClientInfo
//Main Method
func main() {
	//Define Used Data Structure
	
	if len(os.Args) != 2 {
		fmt.Println("Invalid Number of Command Line Argument")
		os.Exit(1)
	}

	tcpAddr_Server := os.Args[1]
	//Server's Overall Structure is referenced from https://coderwall.com/p/wohavg/creating-a-simple-tcp-server-in-go
	//RPC Reference:https://parthdesai.me/articles/2016/05/20/go-rpc-server/
	//Listen for incoming connections
	Iconn,err := net.Listen("tcp",tcpAddr_Server)
	Check_ServerError(err)
	defer Iconn.Close()
	fmt.Println("Listening on" + tcpAddr_Server)
	
	//Register DFSService RPC Server
	DFSService_Instance := new(DFSServiceObj)
	DFSService_rpcServer := rpc.NewServer()
	registerRPC_DFSService(DFSService_rpcServer,DFSService_Instance)
	DFSService_rpcServer.Accept(Iconn)
	
 

}
//Wrappers For Registering RPC Services
func registerRPC_Arith(server *rpc.Server,arith shared.Arith){
	server.RegisterName("Arith_Interface",arith)
}
func registerRPC_DFSService(server *rpc.Server,dfsService shared.DFSService){
	server.RegisterName("DFSService",dfsService)
}

//Separate Thread to handle the request
func handleRequest(conn net.Conn){
	buf := make([]byte,1024)
	reqLen,err := conn.Read(buf)
	Check_NonFatalError(err)
	var receivedMsg shared.OneStringMsg
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
