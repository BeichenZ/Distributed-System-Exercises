/*

This package specifies the application's interface to the distributed
file system (DFS) system to be used in assignment 2 of UBC CS 416
2017W2.

*/

package dfslib

import (
	"fmt"
	"os"
	"io/ioutil"
	"path/filepath"
	"errors"
	"encoding/json"
	"./shared"
	"net/rpc"
	"log"
	"net"
	"unicode"
	"time"
)
 

// A Chunk is the unit of reading/writing in DFS.
type Chunk [32]byte

// Represents a type of file access.
type FileMode int

const (
	// Read mode.
	READ FileMode = iota

	// Read/Write mode.
	WRITE

	// Disconnected read mode.
	DREAD
)

// =================================== Added Codes==================================
type DFSMetaData struct{
	ID int
}
//RPC:Corresponding Calls For app->server calls

type lib_RPCClient struct{
	client *rpc.Client
	ServerAddr string
}
func (t *lib_RPCClient) RegisterNewClient_Remote(localIP string,localPath string) (newClientMetaData DFSMetaData,err error) {
	args := &shared.RNCArgs{LocalIP:localIP,LocalPath:localPath}
	var replyMetaData DFSMetaData
	err = t.client.Call("DFSService.RegisterNewClient",args,&replyMetaData)
	
	return replyMetaData,err
}
func (t *lib_RPCClient)GlobalFileExists_Remote(fname string) (exists bool, err error){
	//To-Do: Return Disconnected Error
	args := &shared.OneStringMsg{Msg:fname}
	var replyData shared.ExistsMsg
	err = t.client.Call("DFSService.GlobalFileExists",args,&replyData)
	if err != nil {
		return false,DisconnectedError(t.ServerAddr) //The connection was timed out
	}
	return replyData.Exists,nil
}	
// DFS:Represent One instance of Client.
type DFSObj struct{
	localIP string
	serverAddr string
	localPath string
	id int
	connected bool
	tcpConn net.Conn
	rpcClient *lib_RPCClient
}
func (dfsObj *DFSObj) LocalFileExists(fname string) (exists bool,err error){
	destFilePath := filepath.Join(dfsObj.localPath,fname)
	if IsAGoodFileName(fname){
		if _,err = os.Stat(destFilePath);err == nil {
			return true,nil		
		}else {
			return false,nil		
		}
	} else {
		return false,BadFilenameError(fname)
	}
}
func IsAGoodFileName(fname string) bool {
	if len(fname)>16 || len(fname)<1 {
		return false
	}
	for _,r := range fname {
		if (!unicode.IsLower(r)) && (!unicode.IsDigit(r)){
			return false		
		}
	}
	return true
}
func (dfsObj *DFSObj) GlobalFileExists(fname string) (exists bool, err error){
	if !IsAGoodFileName(fname) {
		return false,BadFilenameError(fname)	
	}else {
		return dfsObj.rpcClient.GlobalFileExists_Remote(fname)
	}
}
func (dfsObj *DFSObj) Open(fname string,mode FileMode)(f DFSFile,err error){
	return nil,nil
}
func (dfsObj *DFSObj) UMountDFS() error{
	return DisconnectedError("Not Implemented")
}
////////////////////////////////////////////////////////////////////////////////////////////
// <ERROR DEFINITIONS>

// These type definitions allow the application to explicitly check
// for the kind of error that occurred. Each API call below lists the
// errors that it is allowed to raise.
//
// Also see:
// https://blog.golang.org/error-handling-and-go
// https://blog.golang.org/errors-are-values

// Contains serverAddr
type DisconnectedError string

func (e DisconnectedError) Error() string {
	return fmt.Sprintf("DFS: Not connnected to server [%s]", string(e))
}

// Contains chunkNum that is unavailable
type ChunkUnavailableError uint8

func (e ChunkUnavailableError) Error() string {
	return fmt.Sprintf("DFS: Latest verson of chunk [%s] unavailable", string(e))
}

// Contains filename
type OpenWriteConflictError string

func (e OpenWriteConflictError) Error() string {
	return fmt.Sprintf("DFS: Filename [%s] is opened for writing by another client", string(e))
}

// Contains file mode that is bad.
type BadFileModeError FileMode

func (e BadFileModeError) Error() string {
	return fmt.Sprintf("DFS: Cannot perform this operation in current file mode [%s]", string(e))
}

// Contains filename.
type WriteModeTimeoutError string

func (e WriteModeTimeoutError) Error() string {
	return fmt.Sprintf("DFS: Write access to filename [%s] has timed out; reopen the file", string(e))
}

// Contains filename
type BadFilenameError string

func (e BadFilenameError) Error() string {
	return fmt.Sprintf("DFS: Filename [%s] includes illegal characters or has the wrong length", string(e))
}

// Contains filename
type FileUnavailableError string

func (e FileUnavailableError) Error() string {
	return fmt.Sprintf("DFS: Filename [%s] is unavailable", string(e))
}

// Contains local path
type LocalPathError string

func (e LocalPathError) Error() string {
	return fmt.Sprintf("DFS: Cannot access local path [%s]", string(e))
}

// Contains filename
type FileDoesNotExistError string

func (e FileDoesNotExistError) Error() string {
	return fmt.Sprintf("DFS: Cannot open file [%s] in D mode as it does not exist locally", string(e))
}

// </ERROR DEFINITIONS>
////////////////////////////////////////////////////////////////////////////////////////////

// Represents a file in the DFS system.
type DFSFile interface {
	// Reads chunk number chunkNum into storage pointed to by
	// chunk. Returns a non-nil error if the read was unsuccessful.
	//
	// Can return the following errors:
	// - DisconnectedError (in READ,WRITE modes)
	// - ChunkUnavailableError (in READ,WRITE modes)
	Read(chunkNum uint8, chunk *Chunk) (err error)

	// Writes chunk number chunkNum from storage pointed to by
	// chunk. Returns a non-nil error if the write was unsuccessful.
	//
	// Can return the following errors:
	// - BadFileModeError (in READ,DREAD modes)
	// - DisconnectedError (in WRITE mode)
	// - WriteModeTimeoutError (in WRITE mode)
	Write(chunkNum uint8, chunk *Chunk) (err error)

	// Closes the file/cleans up. Can return the following errors:
	// - DisconnectedError
	Close() (err error)
}

// Represents a connection to the DFS system.
type DFS interface {
	// Check if a file with filename fname exists locally (i.e.,
	// available for DREAD reads).
	//
	// Can return the following errors:
	// - BadFilenameError (if filename contains non alpha-numeric chars or is not 1-16 chars long)
	LocalFileExists(fname string) (exists bool, err error)

	// Check if a file with filename fname exists globally.
	//
	// Can return the following errors:
	// - BadFilenameError (if filename contains non alpha-numeric chars or is not 1-16 chars long)
	// - DisconnectedError
	GlobalFileExists(fname string) (exists bool, err error)

	// Opens a filename with name fname using mode. Creates the file
	// in READ/WRITE modes if it does not exist. Returns a handle to
	// the file through which other operations on this file can be
	// made.
	//
	// Can return the following errors:
	// - OpenWriteConflictError (in WRITE mode)
	// - DisconnectedError (in READ,WRITE modes)
	// - FileUnavailableError (in READ,WRITE modes)
	// - FileDoesNotExistError (in DREAD mode)
	// - BadFilenameError (if filename contains non alpha-numeric chars or is not 1-16 chars long)
	Open(fname string, mode FileMode) (f DFSFile, err error)

	// Disconnects from the server. Can return the following errors:
	// - DisconnectedError
	UMountDFS() (err error)
}


// The constructor for a new DFS object instance. Takes the server's
// IP:port address string as parameter, the localIP to use to
// establish the connection to the server, and a localPath path on the
// local filesystem where the client has allocated storage (and
// possibly existing state) for this DFS.
//
// The returned dfs instance is singleton: an application is expected
// to interact with just one dfs at a time.
//
// This call should succeed regardless of whether the server is
// reachable. Otherwise, applications cannot access (local) files
// while disconnected.
//
// Can return the following errors:
// - LocalPathError
// - Networking errors related to localIP or serverAddr
func MountDFS(serverAddr string, localIP string, localPath string) (dfs DFS, err error) {
	// Initialize
	thisDFS := DFSObj {localIP:localIP,serverAddr:serverAddr,localPath:localPath}
	
	//Check Local Path and Write Permission
	if isvalid:= IsValidLocalPath(localPath);!isvalid{
		return nil,LocalPathError(localPath)
	}
	//Check if it's Returning client
	dfsMetaFile_Addr := filepath.Join(localPath,"dfsMeta.json")
	metaFileExist,dfsMetaFile,err := Read_DFSMetaData(dfsMetaFile_Addr)
	CheckNonFatalError(err)

	//Connect to Server
	//tcpServer_Addr,err := net.ResolveTCPAddr("tcp",serverAddr)
	//if err !=nil {
	//	CheckNonFatalError(err)
	//	return nil,err
	//}
	//tcpLocal_Addr,err := net.ResolveTCPAddr("tcp",localIP)
	CheckNonFatalError(err)
	tcpConn, err := net.DialTimeout("tcp",serverAddr,time.Duration(2*time.Second))
	CheckNonFatalError(err)
	//in case it disconnects
	if err != nil {
		thisDFS.connected = false
		thisDFS.id = -1 //invalid ID
		thisDFS.tcpConn = nil
		//if it starts from nothing and disconnected, no dfsMeta.json file should be created
		//Based on pizza 241,no error related to network should return
		return &thisDFS,nil
		
	}
	//Connection is successful
	thisDFS.tcpConn = tcpConn
	thisDFS.connected = true
	//Build RPC Connection
	rpcClient := &lib_RPCClient{client:rpc.NewClient(tcpConn),ServerAddr:serverAddr}
	thisDFS.rpcClient = rpcClient	
	if !metaFileExist {
		newMetaData,err := rpcClient.RegisterNewClient_Remote(localIP,localPath)
		CheckNonFatalError(err)
		Write_DFSMetaData(dfsMetaFile_Addr,newMetaData)
		log.Println("New Client MetaData File Created with ID",newMetaData.ID)
		thisDFS.id = newMetaData.ID			
	}else {
		thisDFS.id = dfsMetaFile.ID
		log.Println("Detect Existing DFS MetaData at",localPath,"with ID:",dfsMetaFile.ID)
	}
	return &thisDFS, nil
}

//Check if a string form a valid path
//Note&Assumption: if a path is valid but leading to a restricted access area 
// given current user's permission, the function will return false and the user 
// would not be able to make changes to the path specified.

func IsValidLocalPath (localPath string) bool{
	if _,err := os.Stat(localPath); err == nil {
		return true
	}else {
		return false
	}
}
//Return False if No MetaDatafile is presented
//else, read and return the meta data
func Read_DFSMetaData (dfsFileAddr string) (fileExists bool,dfsMetaData DFSMetaData,err error) {
	var metaData DFSMetaData
	if _,err := os.Stat(dfsFileAddr);err == nil {
		raw,err := ioutil.ReadFile(dfsFileAddr)
		CheckFatalError(err)
		json.Unmarshal(raw,&metaData)
		return true,metaData,nil
	}else{
		return false,metaData,errors.New("No DFSMetaFile Present at: "+dfsFileAddr)
	}
}
//Write Meta Data to MetaDataFile
//Assume MetaData File already exists
func Write_DFSMetaData(dfsFileAddr string, dfsMetaData DFSMetaData){
	encodedObj,err := json.Marshal(dfsMetaData)
	CheckFatalError(err)
	err = ioutil.WriteFile(dfsFileAddr,encodedObj,0644)//writefile will overwrite!
	CheckFatalError(err)
	f,err := os.Open(dfsFileAddr)
	CheckFatalError(err)
	err = f.Sync()
	CheckFatalError(err)
	err = f.Close()
	CheckFatalError(err)	
}

func CheckNonFatalError (err error) {
	if err != nil {
		fmt.Println("NonFatal Error/Warning(Expectede) Ocurred:", err)
	}
}
func CheckFatalError (err error){
	if err != nil {
		fmt.Println("Fatal Error Occured:",err)
		os.Exit(1)
	}
}

