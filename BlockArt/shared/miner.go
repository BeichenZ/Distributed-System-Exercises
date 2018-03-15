package shared

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/rpc"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type FullSvgInfo struct {
	Path   string
	Fill   string
	Stroke string
}

type AllNeighbour struct {
	sync.RWMutex
	all map[string]*MinerStruct
}

type AllArtNodes struct {
	sync.RWMutex
	all map[string]int
}

type GlobalBlockCreationCounter struct {
	sync.RWMutex
	counter uint8
}

type LongestBlockStruct struct {
	Longest []InfoBlock
}

// type InfoOperation str

type InfoBlock struct {
	ListOperations []Operation
	PreviousHash   string
	CurrentHash    string
}

type SyncAddBlock struct {
	sync.RWMutex
}

var (
	allNeighbour       = AllNeighbour{all: make(map[string]*MinerStruct)}
	blockCounter       = GlobalBlockCreationCounter{counter: 0}
	allArtNodes        = AllArtNodes{all: make(map[string]int)}
	syncingAddingBlock = SyncAddBlock{}
)

type Miner interface {
	Register(address string, publicKey ecdsa.PublicKey) (*MinerNetSettings, error)

	GetNodes(publicKey ecdsa.PublicKey) ([]string, error)

	HeartBeat(publicKey ecdsa.PublicKey, interval uint32) error

	Mine(newOperation Operation) (string, error)

	CheckforNeighbours() bool

	Flood(visited *[]MinerStruct) error

	// RPC methods of Miner
	StopMining(miner MinerStruct, r *MinerStruct) error

	GetBlkChildren(bh string) ([]string, error)

	setUpConnWithArtNode(aip string) error
}

//Struct for descripting Geometry
type Point struct {
	X, Y float64
}
type LineSectVector struct {
	Start, End Point
}

//one move , represent like : m 100 100
type SingleMov struct {
	Cmd    rune
	X      float64
	Y      float64
	ValCnt int
}
type CircleMov struct {
	Cx float64
	Cy float64
	R  float64
}

// One operation contains multiple movs
type SingleOp struct {
	IsClosedShape bool
	MovList       []SingleMov
	InkCost       int
}

type BlockPayloadStruct struct {
	CurrentHash       string
	PreviousHash      string
	R                 big.Int
	S                 big.Int
	CurrentOPs        []Operation
	Children          []BlockPayloadStruct
	DistanceToGenesis int
	Nonce             int32
	SolverPublicKey   string //Make this field a string so no more seg fault
}

type MinerStruct struct {
	ServerAddr            string
	MinerAddr             string
	PairKey               ecdsa.PrivateKey
	Threshold             int
	ArtNodes              []string
	BlockChain            *Block
	ServerConnection      *rpc.Client
	MinerConnection       *rpc.Client
	Settings              MinerNetSettings
	MiningStopSig         chan *Block
	NotEnoughNeighbourSig chan bool
	FoundHash             bool
	RecentHeartbeat       int64
	ListOfOps_str         []string
	RecievedArtNodeSig    chan Operation
	RecievedOpSig         chan Operation
	OPBuffer              []Operation
	MinerInk              uint32
}

type MinerHeartbeatPayload struct {
	client    rpc.Client
	MinerAddr string
}
type MinerInfo struct {
	Address net.Addr
	Key     ecdsa.PublicKey
}

type MinerSettings struct {
	// Hash of the very first (empty) block in the chain.
	GenesisBlockHash string `json:"genesis-block-hash"`

	// The minimum number of ink miners that an ink miner should be
	// connected to.
	MinNumMinerConnections uint8 `json:"min-num-miner-connections"`

	// Mining ink reward per op and no-op blocks (>= 1)
	InkPerOpBlock   uint32 `json:"ink-per-op-block"`
	InkPerNoOpBlock uint32 `json:"ink-per-no-op-block"`

	// Number of milliseconds between heartbeat messages to the server.
	HeartBeat uint32 `json:"heartbeat"`

	// Proof of work difficulty: number of zeroes in prefix (>=0)
	PoWDifficultyOpBlock   uint8 `json:"pow-difficulty-op-block"`
	PoWDifficultyNoOpBlock uint8 `json:"pow-difficulty-no-op-block"`
}
type CanvasSettings struct {
	// Canvas dimensions
	CanvasXMax uint32 `json:"canvas-x-max"`
	CanvasYMax uint32 `json:"canvas-y-max"`
}

// Settings for an instance of the BlockArt project/network.
type MinerNetSettings struct {
	MinerSettings

	// Canvas settings
	CanvasSettings CanvasSettings `json:"canvas-settings"`
}

func copyBigInt(b *big.Int) int64 {
	return b.Int64()
}

func (m *MinerStruct) Register(address string, publicKey ecdsa.PublicKey) (MinerNetSettings, error) {
	// fmt.Println("public key", publicKey)
	///

	// RPC - Start rpc server on this ink miner
	minerServer := &MinerRPCServer{Miner: m}
	rpc.Register(minerServer)
	conn, error := net.Listen("tcp", m.MinerAddr)

	if error != nil {
		fmt.Println(error.Error())
		os.Exit(0)
	}

	go rpc.Accept(conn)

	client, error := rpc.Dial("tcp", address)
	minerSettings := &MinerNetSettings{}
	if error != nil {
		return *minerSettings, error
	}

	m.ServerConnection = client

	// RPC to server
	minerAddress, err := net.ResolveTCPAddr("tcp", m.MinerAddr)

	if err != nil {
		return *minerSettings, err
	}

	minerInfo := &MinerInfo{minerAddress, publicKey}
	err = client.Call("RServer.Register", minerInfo, minerSettings)

	if err != nil {
		return *minerSettings, err
	}
	// CurrentHash  string
	// PreviousHash string
	// // UserSignature     UserSignatureSturct
	// R                 *big.Int
	// S                 *big.Int
	// CurrentOP         Operation
	// Children          []*Block
	// DistanceToGenesis int
	// Nonce             int32
	// SolverPublicKey   *ecdsa.PublicKey
	genesisBlock := Block{
		CurrentHash:       minerSettings.GenesisBlockHash,
		PreviousHash:      "",
		R:                 &big.Int{},
		S:                 &big.Int{},
		CurrentOPs:        make([]Operation, 0),
		DistanceToGenesis: 0,
		Nonce:             int32(0),
		Children:          make([]*Block, 0),
		SolverPublicKey: &ecdsa.PublicKey{
			Curve: elliptic.P384(),
			X:     &big.Int{},
			Y:     &big.Int{},
		},
	}
	m.BlockChain = &genesisBlock
	return *minerSettings, err
}

func (m MinerStruct) HeartBeat() error {
	alive := false

	for {
		error := m.ServerConnection.Call("RServer.HeartBeat", m.PairKey.PublicKey, &alive)
		if error != nil {
			fmt.Println(error)
		}
		time.Sleep(time.Millisecond * time.Duration(800))
	}
}

func AllOperationsCommands(buffer []Operation) string {
	retstring := ""
	for _, op := range buffer {

		retstring += op.ShapeSvgString + op.Fill + op.Stroke + fmt.Sprint(op.AmountOfInk)
	}
	return retstring
}

func AddNewBlock(blk *Block, newBlock *Block) {
	if strings.Compare(blk.CurrentHash, newBlock.PreviousHash) == 0 {
		blk.Children = append(blk.Children, newBlock)
		return
	} else {
		for _, b := range blk.Children {
			AddNewBlock(b, newBlock)
		}
	}
}

func (m *MinerStruct) StartMining(initialOP Operation) (string, error) {
	// currentBlock := m.BlockChain[len(m.BlockChain)-1]
	// listOfOperation := currentBlock.GetStringOperations()

	for {
		select {
		case <-m.NotEnoughNeighbourSig:
			fmt.Println("not enough neighbour, stop minging here")
			// delete(m.LeafNodesMap, leadingBlock.CurrentHash)
			// m.LeafNodesMap[recievedBlock.CurrentHash] = recievedBlock
			return "", nil
		default:
			fmt.Println("I'm starting to mine")
			leadingBlock, _ := findDeepestBlocks(m.BlockChain, 0)
			var difficulutyLevel int
			//
			// fmt.Println("Logging out leading block here")
			// fmt.Println(leadingBlock)

			var nonce string
			isCalculatingNoOp := true
			listOfOpeartion := make([]Operation, 0)
			if len(m.OPBuffer) == 0 {
				//	Mine for no-op
				fmt.Println("Start doing no-op")

				initialOP = Operation{ShapeSvgString: "no-op", AmountOfInk: 0}
				difficulutyLevel = int(m.Settings.PoWDifficultyNoOpBlock)

				nonce = leadingBlock.CurrentHash + initialOP.ShapeSvgString + initialOP.Fill + initialOP.Stroke + fmt.Sprint(initialOP.AmountOfInk) + pubKeyToString(m.PairKey.PublicKey)
				fmt.Println(nonce)

				// Sign the Operation
				r, s, err := ecdsa.Sign(rand.Reader, &m.PairKey, []byte(initialOP.Command))
				if err != nil {
					fmt.Println(err)
				}

				initialOP.Issuer = &m.PairKey
				initialOP.IssuerR = r
				initialOP.IssuerS = s

				listOfOpeartion = append(listOfOpeartion, initialOP)
			} else {
				difficulutyLevel = int(m.Settings.PoWDifficultyOpBlock)
				nonce = leadingBlock.CurrentHash + AllOperationsCommands(m.OPBuffer) + pubKeyToString(m.PairKey.PublicKey)
				fmt.Println(nonce)
				listOfOpeartion = m.OPBuffer
				m.OPBuffer = make([]Operation, 0)
				isCalculatingNoOp = false
			}
			newBlock := doProofOfWork(m, nonce, difficulutyLevel, listOfOpeartion, leadingBlock, isCalculatingNoOp)
			blockCounter.Lock()
			blockCounter.counter++
			blockCounter.Unlock()

			syncingAddingBlock.Lock()
			AddNewBlock(m.BlockChain, newBlock)
			syncingAddingBlock.Unlock()
			// newThing := getLongestPath(m.BlockChain)
			// resultArr := make([]FullSvgInfo, 0)
			// for _, block := range newThing {
			// 	for _, op := range block.CurrentOPs {
			// 		resultArr = append(resultArr, FullSvgInfo{
			// 			Path:   op.ShapeSvgString,
			// 			Fill:   op.Fill,
			// 			Stroke: op.Stroke,
			// 		})
			// 	}
			// }
			// newStrings = append(newStrings, Block{CurrentHash: "ha"})
			// newStrings = append(newStrings, Block{CurrentHash: "wa"})
			// LongestBlockPayload := LongestBlockStruct{Longest: newThing}
			// newerBlock := Block{CurrentHash: "haha"}
			// fmt.Println("Logging out my artnode here")
			// go func() {
			// 	for k, _ := range allArtNodes.all {
			// 		fmt.Println(k)
			// 		reply := false
			// 		client, err := rpc.Dial("tcp", k)
			// 		if err != nil {
			// 			fmt.Println("Cannot connect to artnode: ", err)
			// 		}
			// 		fmt.Println("calling my art node")
			// 		client.Call("CanvasObject.ReceiveLongestChainFromMiner", resultArr, &reply)
			// 	}
			// }()
			// printBlock(m.BlockChain)

			fmt.Println("===================================LONGESTCHAIN===========================================")
			printBlockChain(getLongestPath(m.BlockChain))
			// TODO::
			// Add current blocks' operation to this miners ListOfOps_str
			// TODO maybe validate block here
			fmt.Println("\n")
		}
	}

	// newOperationsList := append(currentBlock.OPS, newOperation)
	//
	// newBlock := Block{newHash, currentBlock.CurrentHash, newOperationsList}
	//
	// m.BlockChain = append(m.BlockChain, newBlock)

	// update all its neighbours

	// return "", nil
}

// Bare minimum flooding protocol, Miner will disseminate notification through the network
func (m MinerStruct) Flood(newBlock *Block, visited *[]*MinerStruct) {
	// TODO construct a list of MinerStruct excluding the senders to avoid infinite loop
	// TODO what happense if node A calls flood, and before it can reach node B, node B calls flood?
	validNeighbours := make([]*MinerStruct, 0)
	fmt.Println("Flooding is called.......................................................")
	for _, v := range allNeighbour.all {
		if filter(v, visited) {
			validNeighbours = append(validNeighbours, v)
		}
	}
	fmt.Println("valid nei", len(validNeighbours))
	if len(validNeighbours) == 0 {

		return
	}

	for _, v := range validNeighbours {
		*visited = append(*visited, v)
	}
	for _, n := range validNeighbours {
		client, error := rpc.Dial("tcp", n.MinerAddr)
		if error != nil {
			fmt.Println(error)
			return
		}

		alive := false
		fmt.Println("visiting miner: ", n.MinerAddr)
		// passingBlock := copyBlock(newBlock)
		err := client.Call("MinerRPCServer.StopMining", newBlock, &alive)
		if err != nil {
			fmt.Println(err)
		}
		n.Flood(newBlock, visited)
	}
	return
}

func (m MinerStruct) FloodOperation(newOP *Operation, visited *[]*MinerStruct) {
	// TODO construct a list of MinerStruct excluding the senders to avoid infinite loop
	// TODO what happense if node A calls flood, and before it can reach node B, node B calls flood?
	validNeighbours := make([]*MinerStruct, 0)
	fmt.Println("Flooding is called.......................................................")
	for _, v := range allNeighbour.all {
		if filter(v, visited) {
			validNeighbours = append(validNeighbours, v)
		}
	}
	if len(validNeighbours) == 0 {

		return
	}

	for _, v := range validNeighbours {
		*visited = append(*visited, v)
	}
	for _, n := range validNeighbours {
		client, error := rpc.Dial("tcp", n.MinerAddr)
		if error != nil {
			fmt.Println(error)
			return
		}

		alive := false
		fmt.Println("visiting miner: ", n.MinerAddr)
		// passingBlock := copyBlock(newBlock)
		err := client.Call("MinerRPCServer.ReceivedOperation", newOP, &alive)
		if err != nil {
			fmt.Println(err)
		}
		n.FloodOperation(newOP, visited)
	}
	return
}

func (m *MinerStruct) produceBlock(currentHash string, newOPs []Operation, leadingBlock *Block, nonce string) *Block {
	// visitedMiners := make([]MinerStruct, 0)
	visitedMiners := []*MinerStruct{m}
	/// Find the leading block
	// CurrentOPs := []Operation{newOP}
	r, s, err := ecdsa.Sign(rand.Reader, &m.PairKey, []byte(currentHash))
	if err != nil {
		fmt.Println(err)
		os.Exit(500)
	}
	fmt.Println("Creating a new block with the new hash")
	fmt.Println(currentHash)
	sss, err := strconv.Atoi(nonce)

	if err != nil {
		fmt.Println(err)
	}
	newDistance := getBlockDistanceFromGensis(m.BlockChain, leadingBlock.CurrentHash)
	producedBlock := &Block{CurrentHash: currentHash,
		PreviousHash:      leadingBlock.CurrentHash,
		CurrentOPs:        newOPs,
		R:                 r,
		S:                 s,
		Children:          make([]*Block, 0),
		SolverPublicKey:   &m.PairKey.PublicKey,
		DistanceToGenesis: newDistance + 1,
		Nonce:             int32(sss)}

	fmt.Println("ITS NONCE IS IS " + nonce)

	m.Flood(producedBlock, &visitedMiners)

	fmt.Println("Need to let the other miners about this block")
	m.FoundHash = false

	fmt.Println("I have found the hash, this is my public key")
	if len(newOPs) == 1 && newOPs[0].Command == "no-op" {
		m.MinerInk += m.Settings.InkPerNoOpBlock
	} else {
		m.MinerInk += m.Settings.InkPerOpBlock
	}
	fmt.Println("Logging out how much ink the miner has")
	fmt.Println(m.MinerInk)
	return producedBlock
}

func (m *MinerStruct) minerSendHeartBeat(minerNeighbourAddr string) error {
	alive := false
	fmt.Println(minerNeighbourAddr)
	// fmt.Println("MAKING RPC CALL TO NEIGHBOUR ", minerNeighbourAddr)
	client, _ := rpc.Dial("tcp", minerNeighbourAddr)
	for {
		fmt.Println("sending heartbeat")
		// fmt.Println(minerToMinerConnection)
		err := client.Call("MinerRPCServer.ReceiveMinerHeartBeat", m.MinerAddr, &alive)
		if err == nil {
			log.Println(err)
		} else {
			return err
		}
		time.Sleep(time.Millisecond * time.Duration(400))
	}
}

func (m *MinerStruct) CheckForNeighbour() {
	var listofNeighbourIP = make([]net.Addr, 0)
	for len(listofNeighbourIP) < int(m.Settings.MinNumMinerConnections) {
		error := m.ServerConnection.Call("RServer.GetNodes", m.PairKey.PublicKey, &listofNeighbourIP)
		if error != nil {
			fmt.Println(error)
		}
	}
	localMax := -1
	var neighbourWithLongestChain string
	blockChain := BlockPayloadStruct{}
	for _, netIP := range listofNeighbourIP {

		fmt.Println("neighbour ip address", netIP.String())
		client, error := rpc.Dial("tcp", netIP.String())
		fmt.Println(client)

		if error != nil {
			fmt.Println(" can't connect")
			fmt.Println(error)
			log.Fatal(error)
			os.Exit(0)
		}
		neighbourBlockChainLength := 0
		// payLoad := MinerHeartbeatPayload{MinerAddr: netIP.String(), client: *client}
		log.Println("NETIP IS ", netIP.String())
		client.Call("MinerRPCServer.MinerRegister", m.MinerAddr, &neighbourBlockChainLength)
		log.Println("the neighbour's blockchain length is: ", neighbourBlockChainLength)
		for {
			if _, exists := allNeighbour.all[netIP.String()]; exists {
				fmt.Printf("The neighbour %v has registered as client", netIP.String())
				break
			}
		}

		if neighbourBlockChainLength > localMax {
			localMax = neighbourBlockChainLength
			neighbourWithLongestChain = netIP.String()
		}
	}
	// TODO get the chain from the neighbour with the longest chain
	longClient, err := rpc.Dial("tcp", neighbourWithLongestChain)
	log.Println("Connected to the longest client")
	if err != nil {
		log.Println(err)
	}
	longClient.Call("MinerRPCServer.SendChain", "give me your chain", &blockChain)
	m.BlockChain = ParseBlockChain(blockChain)
	log.Println("received block chain+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++=")
	// printBlock(m.BlockChain)
}

func (m *MinerStruct) GetBlkChildren(curBlk *Block, bh string) ([]string, error) {
	//fmt.Println("miner.go: GetBlkChildren() prinint the miners genesisBlock Children ", m.BlockChain.Children)
	var bChildHash []string
	if curBlk.CurrentHash == bh {
		bChildHash = make([]string, len(curBlk.Children))
		for i, bc := range curBlk.Children {
			bChildHash[i] = bc.CurrentHash
		}
		return bChildHash, nil
	} else {
		for _, bcc := range curBlk.Children {
			m.GetBlkChildren(bcc, bh)
		}

	}

	return bChildHash, nil
}

func (m *MinerStruct) GetSVGShapeString(curBlk *Block, shapeHash string) string {

	for _, operation := range curBlk.CurrentOPs {

		operationHash := pubKeyToString(operation.Issuer.PublicKey)

		if strings.Compare(operationHash, shapeHash) == 0 {
			return operation.ShapeSvgString
		}
	}

	for _, block := range curBlk.Children {
		return m.GetSVGShapeString(block, shapeHash)

	}
	return ""
}

func (m *MinerStruct) GetOpToDelete(curBlk *Block, shapeHash string) Operation {
	var opToDelete Operation

	for _, operation := range curBlk.CurrentOPs {

		operationHash := pubKeyToString(operation.Issuer.PublicKey)

		if strings.Compare(operationHash, shapeHash) == 0 {
			opToDelete = operation
			return opToDelete
		}
	}

	for _, block := range curBlk.Children {
		return m.GetOpToDelete(block, shapeHash)

	}

	return opToDelete

}

func (m *MinerStruct) GetInkBalance() uint32 {
	return 0

}

func (m *MinerStruct) setUpConnWithArtNode(aip string) error {
	// add ip to map
	allArtNodes.Lock()
	defer allArtNodes.Unlock()
	if _, exist := allArtNodes.all[aip]; exist {
		fmt.Println("art node already registered")
	} else {
		allArtNodes.all[aip] = 1
	}
	fmt.Println("setUpConnWithArtNode() Going to Dial up my art node")
	// _, err := rpc.Dial("tcp", aip)
	// CheckError(err)
	return nil
}
