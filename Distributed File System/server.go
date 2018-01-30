package main

import (
	//"crypto/md5"
	//"encoding/hex"

	// TODO
	"encoding/json"
	"fmt"
	"log"
	//"math/rand"
	"net"
	"os"
	"net/rpc"
	//"./dfslib"
	"./dfslib/shared"
	"errors"
	"strings"
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
			break //since no user can be deleted.The first empty one cannot be repetitive
		}else {
			if strings.Compare(element.localIP,args.LocalIP)==0 && strings.Compare(element.localPath,args.LocalPath)==0 {
				availableIndex = index
				break //if this user is previous registered			
			} 
		} 
		
		
	}
	if availableIndex == -1 {
		return errors.New("Application:Client Count is more than 16.")
	}
	//Populate Local Client Info
	clientList[availableIndex].localIP = args.LocalIP
	clientList[availableIndex].localPath = args.LocalPath
	clientList[availableIndex].occupied = true
	clientList[availableIndex].ID = availableIndex+1
	clientList[availableIndex].fileMap = make(map[string]SingleDFSFileInfo_ClientList)
	//Pass Back Remote Client Info
	(*reply).ID = availableIndex+1
	return nil
}

func (t *DFSServiceObj)GlobalFileExists(args *shared.OneStringMsg, reply *shared.ExistsMsg) error{
	//To-Do: Disconnected Error
	for _,element := range clientList {
		if _,exists := element.fileMap[args.Msg];exists{
			fmt.Println("Find File Name",args.Msg,"under ID:",element.ID)
			(*reply).Exists = true
			return nil
		}
	}
	(*reply).Exists = false
	return nil
}
//Functionality:Register New File. Update status of a file a specific client holds during Open & Write
func (t *DFSServiceObj)UpdateFileInfo(args *shared.GenericArgs, reply *shared.GenericReply) error{
	isNewFile := args.BoolOne
	fname := args.StringOne
	clientIndex := args.IntOne - 1
	//For Newly Created File, add it to the client directly
	//Note:ClientID = 1 + Its_Index_in_clientList[]
	if isNewFile {
		//Fill info for clientList
		client := clientList[clientIndex]
		client.fileMap[fname] = SingleDFSFileInfo_ClientList{} // by Default, version=0 is trivial version	
		//Add new entry to globalFileMap
		var tempArray [256]int
		globalFileMap[fname] = SingleDFSFileInfo_FileList{fname:fname,chunkCount:0,chunkVersionMap:make(map[int]ChunkVersionToHolderMap),topVersion:tempArray}
		log.Println("New File Entry Created with Name :",globalFileMap[fname].fname)
		(*reply).BoolOne = true
		return nil
	} else {
		return nil
	}
}

//Data Structure Type
type SingleClientInfo struct {
	occupied bool
	localIP string
	localPath string
	ID int
	fileMap map[string]SingleDFSFileInfo_ClientList
}
//Single File Info for Client List to use
type SingleDFSFileInfo_ClientList struct{
	chunkVersionArray [256]int //version number 0 is the default version
}
//For a single Chunk.key:version number,value:array of whether client at index contains such version 
type ChunkVersionToHolderMap map[int][16]bool
//Single File Info for File List to Use
type SingleDFSFileInfo_FileList struct {
	chunkCount int
	chunkVersionMap map[int]ChunkVersionToHolderMap//Key:Chunk No, Value:Version to Holder's List Map 
	topVersion [256]int //Top version number for 256 possible chunk
	fname string
}
//Global Data Storage Shared by Multiple RPC calls and Main
var clientList [16]SingleClientInfo
var globalFileMap map[string]SingleDFSFileInfo_FileList
//Main Method
func main() {
	//Define Used Data Structure
	
	if len(os.Args) != 2 {
		fmt.Println("Invalid Number of Command Line Argument")
		os.Exit(1)
	}
	tcpAddr_Server := os.Args[1]

	//Initialize Data Structures:
	globalFileMap = make(map[string]SingleDFSFileInfo_FileList)
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
