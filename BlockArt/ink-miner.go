package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/gob"
	"flag"
	"fmt"
	"net"
	//"net/http"

	"net/rpc"
	"os"

	shared "./shared"
	"encoding/hex"
	"crypto/x509"
	"reflect"
)

//var globalInkMinerPairKey ecdsa.PrivateKey
var thisInkMiner *shared.MinerStruct

func main() {
	// Register necessary struct for server communications
	gob.Register(&elliptic.CurveParams{})
	gob.Register(&net.TCPAddr{})

	// Construct minerAddr from flag provided in the terminal
	//minerPort := flag.String("p", "", "RPC server ip:port")
	//servAddr := flag.String("sa", "", "Server ip address")
	//pubKeyStr := flag.String("pubk", "", "a string")
	//privKeyStr := flag.String("privk", "", "a string")

	minerPort := os.Args[1]
	servAddr := os.Args[2]
	pubKeyStr := os.Args[3]
	privKeyStr := os.Args[4]

	flag.Parse()
	minerAddr := "127.0.0.1:" + minerPort

	// initialize miner given the server address and its own miner address
	inkMinerStruct := initializeMiner(servAddr, minerAddr, pubKeyStr,privKeyStr )
	//globalInkMinerPairKey = inkMinerStruct.PairKey
	fmt.Println("Miner Key: ", inkMinerStruct.PairKey.X)
	thisInkMiner = &inkMinerStruct
	// RPC - Register this miner to the server
	minerSettings, error := inkMinerStruct.Register(servAddr, inkMinerStruct.PairKey.PublicKey)
	if error != nil {
		fmt.Println(error.Error())
		os.Exit(0)
	}

	// setting returned from the server
	inkMinerStruct.Settings = minerSettings

	//start heartbeat to the server
	// heartBeatChannel := make(chan int)
	go inkMinerStruct.HeartBeat()
	// <-heartBeatChannel

	// Listen for Art noded that want to connect to it
	fmt.Println("Going to Listen to Art Nodes: ")
	listenArtConn, err := net.Listen("tcp", "127.0.0.1:") // listening on wtv port
	shared.CheckError(err)
	fmt.Println("Port Miner is lisening on ", listenArtConn.Addr())

	// check that the art node has the correct public/private key pair
	//initArt := new(shared.KeyCheck)
	initArt := &shared.KeyCheck{inkMinerStruct}

	rpc.Register(initArt)
	cs := &shared.CanvasSet{inkMinerStruct}
	rpc.Register(cs)
	anr := &shared.ArtNodeOpReg{&inkMinerStruct}
	go rpc.Register(anr)
	go rpc.Accept(listenArtConn)

	// While the heart is beating, keep fetching for neighbours

	// After going over the minimum neighbours value, start doing no-op

	OP := shared.Operation{ShapeSvgString: "no-op"}
	for {
		inkMinerStruct.CheckForNeighbour()
		inkMinerStruct.StartMining(OP)
	}
	return
}

func initializeMiner(servAddr string, minerAddr string, puks string, prks string) shared.MinerStruct {
	minerKey, _ := parseKeyPair(puks,prks)
	killSig := make(chan *shared.Block)
	NotEnoughNeighbourSig := make(chan bool)
	RecievedArtNodeSig := make(chan shared.Operation)
	RecievedOpSig := make(chan shared.Operation)

	return shared.MinerStruct{ServerAddr: servAddr,
		MinerAddr:             minerAddr,
		PairKey:               *minerKey,
		MiningStopSig:         killSig,
		NotEnoughNeighbourSig: NotEnoughNeighbourSig,
		FoundHash:             false,
		RecievedArtNodeSig:    RecievedArtNodeSig,
		RecievedOpSig:         RecievedOpSig,
	}
}

func parseKeyPair(pubKeyStr string, privKeyStr string) (*ecdsa.PrivateKey, error) {
	privKeyRestByte, _ := hex.DecodeString(privKeyStr)
	priv2, _ := x509.ParseECPrivateKey(privKeyRestByte)
	pubKeyRestByte, _ := hex.DecodeString(pubKeyStr)
	pub2, _ := x509.ParsePKIXPublicKey(pubKeyRestByte)
	if reflect.DeepEqual(priv2.Public(), pub2) {
		fmt.Println("parseKeyPair() keys are all good")
		return priv2, nil
	} else {
		fmt.Println("parseKeyPair() keys are not good")
		return nil, shared.InvalidKeyError(pubKeyStr)
	}
}
