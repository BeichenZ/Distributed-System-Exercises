package shared

import (
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"log"
	"math"
	"net/rpc"
	"reflect"
	"time"
)

type Block_OP_Index struct {
	IndexB, IndexO int
}
type MinerRPCServer struct {
	Miner *MinerStruct
}

type BadBlockError string

func (e BadBlockError) Error() string {
	return fmt.Sprintf("BAD BLOCK")
}

func (l *ArtNodeOpReg) GiveMeBlockTree(reply *bool, recievedBlock *BlockPayloadStruct) error {

	thisMinerBlockChain := CopyBlockChainPayload(l.Miner.BlockChain)
	*recievedBlock = thisMinerBlockChain

	return nil
}

func (m *MinerRPCServer) SendChain(s string, blockChain *BlockPayloadStruct) error {
	log.Println(s + "============================================================")
	thisMinerBlockChain := CopyBlockChainPayload(m.Miner.BlockChain)
	*blockChain = thisMinerBlockChain
	return nil
}

func (m *MinerRPCServer) StopMining(block *Block, alive *bool) error {
	log.Println("stopped")
	if !m.Miner.FoundHash {
		log.Print("Someone send me a block, its hash is: ", block.CurrentHash)
		if block.Validate() {
			log.Println("Successfully Validated the block")
			go func() {
				m.Miner.MiningStopSig <- block
			}()
		} else {
			return BadBlockError("BAD BLOCK")
		}
	} else {
		log.Print("I have found the hash, but so did at least one other miner")
		log.Print("The hash is ", block.CurrentHash)
		syncingAddingBlock.Lock()
		AddNewBlock(m.Miner.BlockChain, block)
		syncingAddingBlock.Unlock()
	}
	log.Println("Sent channel info")
	return nil
}

func (m *MinerRPCServer) ReceivedOperation(operation *Operation, alive *bool) error {

	m.Miner.RecievedOpSig <- *operation

	return nil
}

func (m *MinerRPCServer) ReceiveMinerHeartBeat(minerNeighbourAddr string, alive *bool) error {
	allNeighbour.Lock()
	defer allNeighbour.Unlock()
	// log.Println("Heartbeat received from ", minerNeighbourAddr)
	// fmt.Println("_____________________________________________________________________________________")
	// fmt.Printf("%+v", allNeighbour.all)
	if _, ok := allNeighbour.all[minerNeighbourAddr]; ok {
		// log.Println("The miner is in the map")
		allNeighbour.all[minerNeighbourAddr].RecentHeartbeat = time.Now().UnixNano()
		// fmt.Println("Heartbeat RECEIVED ", allNeighbour.all[minerNeighbourAddr].RecentHeartbeat)
	} else {
		// log.Println("Nothing in the map")
	}
	return nil
}

func (m *MinerRPCServer) ActivateHeartBeat(SenderAddr string, alive *bool) error {
	go m.Miner.minerSendHeartBeat(SenderAddr)
	return nil
}

func (m *MinerRPCServer) ArtNodeRegister(ArtNodeAddr string, reply *bool) error {
	if _, exist := allArtNodes.all[ArtNodeAddr]; exist {
		fmt.Println("Artnode already here")
	} else {
		allArtNodes.Lock()
		allArtNodes.all[ArtNodeAddr] = 1
		allArtNodes.Unlock()
	}
	return nil
}

func (m *MinerRPCServer) MinerRegister(MinerNeighbourPayload *string, thisMinerChainLength *int) error {
	fmt.Println("------------------------------------I got here--------------------------------------")
	fmt.Println(MinerNeighbourPayload)
	// if len(allNeighbour.all) > int(m.Miner.Settings.MinNumMinerConnections) {
	// 	m.Miner.NotEnoughNeighbourSig <- false
	// }

	if _, ok := allNeighbour.all[(*MinerNeighbourPayload)]; ok {
		log.Println("The Miner is already here")
	} else {
		alive := false
		number := 0
		allNeighbour.all[(*MinerNeighbourPayload)] = &MinerStruct{
			MinerAddr: (*MinerNeighbourPayload),
			// MinerConnection: &MinerNeighbourPayload.client,
			RecentHeartbeat: time.Now().UnixNano(),
		}
		_, length := findDeepestBlocks(m.Miner.BlockChain, 0)
		//length := m.Miner.FindLongestChainLength()
		*thisMinerChainLength = length
		log.Println("Registration time is: ", time.Now().UnixNano())
		log.Println("Successfully recorded this neighbouring miner", (*MinerNeighbourPayload))
		client, error := rpc.Dial("tcp", *MinerNeighbourPayload)
		if error != nil {
			fmt.Println(error)
		}
		err := client.Call("MinerRPCServer.MinerRegister", m.Miner.MinerAddr, &number)
		if err != nil {
			fmt.Println(err)
		}
		go monitor(*MinerNeighbourPayload, *m.Miner, 1000*time.Millisecond)

		client.Call("MinerRPCServer.ActivateHeartBeat", m.Miner.MinerAddr, &alive)

	}
	return nil
}

type KeyCheck struct {
	Miner MinerStruct
}

func (l *KeyCheck) ArtNodeKeyCheck(anm ArtnodeVer, validKey *bool) error {
	fmt.Println("ArtNodeKeyCheck() Going to check key")
	if !ecdsa.Verify(&l.Miner.PairKey.PublicKey, anm.Msg, anm.Ra, anm.Sa) {
		fmt.Printf("Error key aint same")
		*validKey = false
		return nil
	}
	fmt.Printf("key is same message is ", anm.Msg)
	*validKey = true

	return nil
}

type CanvasSet struct {
	Miner MinerStruct
}
type ArtNodeOpReg struct {
	Miner *MinerStruct
}

func (l *ArtNodeOpReg) DoArtNodeOp(op *Operation, reply *int) error {
	//Sign the operation
	r, s, err := ecdsa.Sign(rand.Reader, &l.Miner.PairKey, []byte((*op).ShapeSvgString))

	if err != nil {
		fmt.Println(err)
	}

	op.Issuer = &l.Miner.PairKey
	op.IssuerR = r
	op.IssuerS = s
	fmt.Println(pubKeyToString(op.Issuer.PublicKey))
	//reply decode: 0->Success,1->insufficientInk, 2->OverlappedShape
	// Check InsufficientInkError
	fmt.Println("STARTING MINING+++++++++++++++++")
	if l.Miner.MinerInk < (*op).AmountOfInk {
		fmt.Println("The current ink we have is:", l.Miner.MinerInk)
		fmt.Println("Insufficient Ink Detected for shape:", (*op).Command, "With requested ink:", (*op).AmountOfInk)
		*reply = 1
		return nil
	}
	// Check ShapeOverlapError

	if isOverLap := IsShapeOverLapWithOthers(op, l); isOverLap {
		fmt.Println("OverLapped Shape for shape string:", (*op).Command)
		*reply = 2
		return nil
	}
	// switch op.Shapetype {
	// case PATH:
	// 	/*
	// 		if isOverLap := IsShapeOverLapWithOthers(op,l); isOverLap {
	// 			fmt.Println("OverLapped Shape for shape string:", (*op).Command)
	// 			*reply = 2
	// 			return nil
	// 		}
	// 	*/
	// case CIRCLE:
	// 	if IsShapeOverLapWithOthers_Cir(op, l) {
	// 		*reply = 2
	// 		return nil
	// 	}
	// default:
	// 	*reply = 4
	// 	return nil
	//
	// }

	fmt.Println(op.Command)

	go func() {
		l.Miner.RecievedArtNodeSig <- *op
	}()

	blockCounter.Lock()
	origCounter := blockCounter.counter
	blockCounter.Unlock()

	// TODO  What if multiple art node tries to make operation

	validNum := op.ValidFBlkNum
	// ** Maybe set timeout if it takes too long
	// while l.Miner.chain is not validateNum more than before don't return anything, only return when its not
	waitStartTime := time.Now()
	then := time.Now()
	for {
		//fmt.Println("DoArtNodeOp: IN loop")
		//Set-up overall timeout
		then = time.Now()
		if then.Sub(waitStartTime).Seconds() > 50 {
			*reply = 3
			return DisconnectedError("Wait For ValidateNum Take Too Long")
		}
		blockCounter.Lock()
		if (blockCounter.counter - origCounter) == validNum {
			*reply = 0
			blockCounter.Unlock()
			break
		}
		blockCounter.Unlock()
	}
	fmt.Println("DoArtNodeOp() validateNum condition satisfied")
	return nil
}

func IsShapeOverLapWithOthers(op *Operation, l *ArtNodeOpReg) bool {
	//For operation from same miner , do not check
	//For operation from different miner, check for overlapping

	svgString := (*op).ShapeSvgString
	fmt.Println("svgString:", svgString)
	newsvgFill := (*op).Fill
	fmt.Println("newSvgFill:", newsvgFill)
	//svgStroke := (*op).Stroke
	newsvgArea := (*op).AmountOfInk
	svgPrivateKey_ptr := (*op).Issuer
	_, newSvgOp := IsSvgStringParsable_Parse(svgString)
	_, newvtxArr, newedgeArr := IsClosedShapeAndGetVtx(newSvgOp)
	var tarSvgOp SingleOp
	//var tarSvgOp_IsClosed bool
	var tarvtxArr []Point
	var taredgeArr []LineSectVector
	isTwoOverLap := false
	//Below Two varialbe is to prevent the siutation of add->delete->false conflict situation
	var Block_OP_IndexArr []Block_OP_Index //Record all the conflicting shape and check if they have been deleted
	//Store number of such pair happens, No one should appear even numbers

	minerCopy := *(l.Miner)
	canvasMaxX := minerCopy.Settings.CanvasSettings.CanvasXMax
	longestChainArr_Invt := getLongestPath(minerCopy.BlockChain)
	if len(longestChainArr_Invt) < 2 {
		return false
	}

	//Loop through Each Block.
BlockLoop:
	for indexB, tarBlock := range longestChainArr_Invt {
		if len(tarBlock.CurrentOPs) == 0 { //Check Genesis Block
			fmt.Println("Meet Genesis Block")
			continue BlockLoop
		} else if tarBlock.CurrentOPs[0].ShapeSvgString == "no-op" {
			fmt.Println("Meet Noop Block")
			continue BlockLoop
		}
	OperationLoop:
		for indexO, tarOp := range tarBlock.CurrentOPs {
			//HandleDelete Operation.Not counted
			if tarOp.Draw == false {
				deletedIndex := Block_OP_Index{indexB, indexO}
				for index, element := range Block_OP_IndexArr {
					if element == deletedIndex {
						if index == len(Block_OP_IndexArr)-1 {
							Block_OP_IndexArr = Block_OP_IndexArr[:index]
							break
						} else {
							Block_OP_IndexArr = append(Block_OP_IndexArr[:index], Block_OP_IndexArr[index+1:]...)
							break
						}
					}
				}
				continue OperationLoop //since it's a delete action, no need to check any further
			}

			//Handle Two ops are from same public key.Dont check if it's same
			if tarOp.Issuer != nil && svgPrivateKey_ptr != nil {
				if reflect.DeepEqual(*(tarOp.Issuer), *svgPrivateKey_ptr) {
					fmt.Println("TwoOperationFromSame Miner, Dont check")
					continue OperationLoop
				}
			}
			//TODO:Handle differently for Circle and Straignt-Line Shape
			//Initialize Data of Target Shape
			isTwoOverLap = false //Initialize to be false for each check
			tarArea := tarOp.AmountOfInk
			tarFill := tarOp.Fill
			_, tarSvgOp = IsSvgStringParsable_Parse(tarOp.ShapeSvgString)
			_, tarvtxArr, taredgeArr = IsClosedShapeAndGetVtx(tarSvgOp)

			//Handle TWO Different Cases.Assumption all string at this stage are all valid
			//Case1: Both are line only shape(transparent fill).
			if newsvgFill == "transparent" && tarOp.Fill == "transparent" {
				isTwoOverLap = IsTwoEdgeArrInterSect(newedgeArr, taredgeArr) // Line Shape only need to check intersect
			} else {
				//Case2: One of them is filled. Check Edge Intersect first, if not then if all point in side
				if isTwoOverLap = IsTwoEdgeArrInterSect(newedgeArr, taredgeArr); !isTwoOverLap {
					isTwoOverLap = IsOneShapeCompleteInsideAnother(newvtxArr, newedgeArr, newsvgFill, newsvgArea, tarvtxArr, taredgeArr, tarFill, tarArea, canvasMaxX)
				}
			}
			//Register the results
			if isTwoOverLap {
				Block_OP_IndexArr = append(Block_OP_IndexArr, Block_OP_Index{indexB, indexO})
			}
		}
	}

	//Check if the any of the conflicting shape has actually be deleted
	if len(Block_OP_IndexArr) == 0 {
		//fmt.Println("One of OverLapped String Pair is", svgString, "and", longestChainArr_Invt[Block_OP_IndexArr[0].IndexB].CurrentOPs[Block_OP_IndexArr[0].IndexO].ShapeSvgString)
		return false
	} else {
		//Only preint out the very first conflicted array
		return true
	}
}
func IsShapeOverLapWithOthers_Cir(cop *Operation, l *ArtNodeOpReg) bool {
	minerCopy := *(l.Miner)
	longestChainArr := getLongestPath(minerCopy.BlockChain)
	var isOverLapped bool
	_, curCircleParsed := IsSvgStringParsable_Parse_Cir(cop.ShapeSvgString)
	for _, b := range longestChainArr {
		for _, ob := range b.CurrentOPs {
			if ob.ShapeSvgString != "no-op" {
				switch ob.Shapetype {
				case PATH:
					_, tarSvgOp := IsSvgStringParsable_Parse(ob.ShapeSvgString)
					tartSvgOp_IsClosed, tarvtxArr, taredgeArr := IsClosedShapeAndGetVtx(tarSvgOp)
					fmt.Println(tartSvgOp_IsClosed)
					// basic intersection test
					// check the intersections of edges
					isOverLapped = IsPerimeterOverlapPoints_Cir(curCircleParsed, tarvtxArr) // overlap with points
					if isOverLapped {
						return isOverLapped
					}
					// overlap check with edges
					isOverLapped = IsPerimeterOverlapEdges_Cir(curCircleParsed, taredgeArr)
					if isOverLapped {
						return isOverLapped
					}
					// Circle and Path are transparent -- no intersection
					if ((ob.Fill == "transparent") || (ob.Fill == "none")) && ((cop.Fill == "transparent") || (cop.Fill == "none")) {
						return false
					}

				case CIRCLE:
					_, tarCir := IsSvgStringParsable_Parse_Cir(ob.ShapeSvgString)
					dc := relCentreDist_Cir(curCircleParsed, tarCir)
					if dc == 0 {
						return true
					} // circles are coincident
					if dc > 0 { // circles don't intersect fantastico
						return false
					}
					if dc < 0 { // circles do intersect
						if ((ob.Fill == "transparent") || (ob.Fill == "none")) && ((cop.Fill == "transparent") || (cop.Fill == "none")) {
							return false
						}
						// both filled
						if (ob.Fill != "transparent") || (ob.Fill != "none") && (cop.Fill != "transparent") || (cop.Fill != "none") {
							return true
						}
						// find the smaller circle
						if curCircleParsed.R < tarCir.R {
							return (((cop.Fill == "transparent") || (cop.Fill == "none")) && ((ob.Fill != "transparent") || (ob.Fill != "none")))

						} else {
							return (((ob.Fill == "transparent") || (ob.Fill == "none")) && ((cop.Fill != "transparent") || (cop.Fill != "none")))
						}
					}

					return false
				}
			}

		}
	}
	return isOverLapped
}

func IsPerimeterOverlapPoints_Cir(c CircleMov, tarvtxArr []Point) bool {
	for _, p := range tarvtxArr {
		if math.Pow((p.X-c.Cx), 2)-math.Pow((p.X-c.Cx), 2) == math.Pow(c.R, 2) {
			return true
		}

	}
	return false

}
func IsPerimeterOverlapEdges_Cir(c CircleMov, taredgeArr []LineSectVector) bool {
	for _, l := range taredgeArr {
		x1 := l.Start.X
		y1 := l.Start.Y
		x2 := l.End.X
		y2 := l.End.Y
		d := (math.Abs((x2-x1)*c.Cx + (y1-y2)*c.Cy + (x1-x2)*c.Cy + (y2-y1)*c.Cx)) / math.Sqrt(math.Pow(x2-x1, 2)+math.Pow(y1-y2, 2))
		if d < c.R {
			return true
		}
	}
	return false

}
func IsPerimeterOverlap_Cir(c CircleMov, ct CircleMov) bool {
	d := math.Abs(math.Pow(c.Cx-ct.Cx, 2) + math.Pow(c.Cy-ct.Cy, 2))
	if d == 0 {
		return true
	}
	return false

}
func relCentreDist_Cir(c CircleMov, ct CircleMov) float64 {
	d := math.Abs(math.Pow(c.Cx-ct.Cx, 2) + math.Pow(c.Cy-ct.Cy, 2))
	return d

}
func (l *ArtNodeOpReg) ArtnodeInkRequest(s string, remainInk *uint32) error {
	*remainInk = l.Miner.GetInkBalance()
	return nil
}

func (l *ArtNodeOpReg) ArtnodeGenBlkRequest(s string, genBlkHash *string) error {
	*genBlkHash = l.Miner.BlockChain.CurrentHash
	fmt.Println("ArtnodeGenBlkRequest() ", genBlkHash)
	return nil
}

func (l *ArtNodeOpReg) ArtnodeBlkChildRequest(bHash string, blkCh *[]string) (err error) {
	*blkCh, err = l.Miner.GetBlkChildren(l.Miner.BlockChain, bHash)
	return err
}

func (l *ArtNodeOpReg) ArtnodeSvgStringRequest(shapeHash string, svgString *string) (err error) {
	*svgString = l.Miner.GetSVGShapeString(l.Miner.BlockChain, shapeHash)
	return err
}
func (l *ArtNodeOpReg) ArtnodeGetOpWithHashRequest(shapeHash string, opToDel *Operation) error {
	*opToDel = l.Miner.GetOpToDelete(l.Miner.BlockChain, shapeHash)
	return nil
}

func (l *CanvasSet) GetCanvasSettingsFromMiner(s string, ics *InitialCanvasSetting) error {
	fmt.Println("request for CanvasSettings")
	ics.Cs = CanvasSettings(l.Miner.Settings.CanvasSettings)
	ics.ListOfOps_str = l.Miner.ListOfOps_str
	l.Miner.setUpConnWithArtNode(s)
	fmt.Println("GetCanvasSettingsFromMiner() ", *ics)
	return nil
}

func CheckError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
	}
}
