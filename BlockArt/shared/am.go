// Artnode - Miner API

package shared

import (
	"crypto/ecdsa"
	"fmt"
	"net/rpc"
	//"time"
	"log"
	"math/big"

)

type ArtNodeStruct struct {
	ArtNodeId int
	PairKey   ecdsa.PrivateKey
	AmConn    *rpc.Client
}

type ArtnodeVer struct {
	Msg []byte
	Ra *big.Int
	Sa *big.Int
}


type InitialCanvasSetting struct {
	Cs            CanvasSettings
	ListOfOps_str []string
}
type AddShapeReply struct {
	Success bool
	Err     error
}

func (a *ArtNodeStruct) GetCanvasSettings(anIP string) (InitialCanvasSetting, error) {
	initCS := &InitialCanvasSetting{}
	fmt.Println("GetCanvasSettings() got here")
	err := a.AmConn.Call("CanvasSet.GetCanvasSettingsFromMiner", anIP, initCS)
	CheckError(err)
	return *initCS, err
}

// Artnode issues an operation
// returns a number which indicates which indicates the status of the operation
// for now boolean
func (a *ArtNodeStruct) ArtnodeOp(op Operation) (validOp bool, err error) {
	var reply int
	locaerr := a.AmConn.Call("ArtNodeOpReg.DoArtNodeOp", op, &reply)
	CheckError(locaerr)
	//Parse the return int to error
	if locaerr != nil {
		return false, DisconnectedError("AddShape() Disconnected") //Any Undefined error will be disconnected error
	}
	switch reply {
	case 0:
		return true, nil
	case 1:
		return false, InsufficientInkError(uint32(op.AmountOfInk))
	case 2:
		return false, ShapeOverlapError(op.ShapeSvgString)
	case 3:
		return false, TimedOutTooLongError("DoArtNodeOp")
	default:
		return false, nil
	}
}

func (a *ArtNodeStruct) GetBlockTreeFromMiner() (*Block, error) {

	receivedBlockChain := BlockPayloadStruct{}

	reply := false
	err := a.AmConn.Call("ArtNodeOpReg.GiveMeBlockTree", reply, &receivedBlockChain)
	if err != nil {
		log.Println(err)
		return nil, err
	}



	thisNewBlock := ParseBlockChain(receivedBlockChain)

	return thisNewBlock, err
}

func (a *ArtNodeStruct) GetInkBalFromMiner() (uint32, error) {
	var i uint32
	err := a.AmConn.Call("ArtNodeOpReg.ArtnodeInkRequest", "ink pls", &i)
	fmt.Println("GetInkBalFromMiner() ", i)
	return i, err
}
func (a *ArtNodeStruct) GetGenesisBlockFromMiner() (string, error) {
	var gb string
	err := a.AmConn.Call("ArtNodeOpReg.ArtnodeGenBlkRequest", "Genisis blk", &gb)
	return gb, err
}
func (a *ArtNodeStruct) GetChildrenFromMiner(bHash string) ([]string, error) {
	var mch []string

	err := a.AmConn.Call("ArtNodeOpReg.ArtnodeBlkChildRequest", bHash, &mch)
	return mch, err
}
func (a *ArtNodeStruct) GetSvgStringUsingOperationSignature(shapeHash string) (string, error) {
	var svgstring string

	err := a.AmConn.Call("ArtNodeOpReg.ArtnodeSvgStringRequest", shapeHash, &svgstring)
	return svgstring, err
}
func (a *ArtNodeStruct) GetOpWithHash(shapeHash string) (Operation, error) {

	var delOp Operation
	err := a.AmConn.Call("ArtNodeOpReg.ArtnodeGetOpWithHashRequest", shapeHash, &delOp)
	if len(delOp.Command) == 0 {
		err = ShapeOwnerError("")
	}
	if err != nil {
		err = DisconnectedError("")
	}
	return delOp, err
}
