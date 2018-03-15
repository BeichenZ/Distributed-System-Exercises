/*

This package specifies the application's interface to the the BlockArt
library (blockartlib) to be used in project 1 of UBC CS 416 2017W2.

*/

package blockartlib

import (
	"crypto/ecdsa"
	"fmt"
	"math"
	"net"
	"net/rpc"
	"runtime"

	shared "../shared"
	cr "crypto/rand"

)

// Represents a type of shape in the BlockArt system.

var BlockTree *shared.Block
var BlockChain []shared.FullSvgInfo

// const (
// 	// Path shape.
// 	PATH shared.ShapeType = iota
// 	// CIRCLE
// 	// Circle shape (extra credit).
//
// )

// Settings for a canvas in BlockArt.
type CanvasSettings struct {
	// Canvas dimensions
	CanvasXMax uint32
	CanvasYMax uint32
}

type LongestBlockStruct struct {
	longest []shared.Block
}

// Settings for an instance of the BlockArt project/network.
type MinerNetSettings struct {
	// Hash of the very first (empty) block in the chain.
	GenesisBlockHash string

	// The minimum number of ink miners that an ink miner should be
	// connected to. If the ink miner dips below this number, then
	// they have to retrieve more nodes from the server using
	// GetNodes().
	MinNumMinerConnections uint8

	// Mining ink reward per op and no-op blocks (>= 1)
	InkPerOpBlock   uint32
	InkPerNoOpBlock uint32

	// Number of milliseconds between heartbeat messages to the server.
	HeartBeat uint32

	// Proof of work difficulty: number of zeroes in prefix (>=0)
	PoWDifficultyOpBlock   uint8
	PoWDifficultyNoOpBlock uint8

	// Canvas settings
	canvasSettings CanvasSettings
}

// Represents a canvas in the system.
type Canvas interface {
	// Adds a new shape to the canvas.
	// Can return the following errors:
	// - DisconnectedError
	// - InsufficientInkError
	// - InvalidShapeSvgStringError
	// - ShapeSvgStringTooLongError
	// - ShapeOverlapError
	// - OutOfBoundsError
	AddShape(validateNum uint8, shapeType shared.ShapeType, shapeSvgString string, fill string, stroke string) (shapeHash string, blockHash string, inkRemaining uint32, err error)

	// Returns the encoding of the shape as an svg string.
	// Can return the following errors:
	// - DisconnectedError
	// - InvalidShapeHashError
	GetSvgString(shapeHash string) (svgString string, err error)

	// Returns the amount of ink currently available.
	// Can return the following errors:
	// - DisconnectedError
	GetInk() (inkRemaining uint32, err error)

	// Removes a shape from the canvas.
	// Can return the following errors:
	// - DisconnectedError
	// - ShapeOwnerError
	DeleteShape(validateNum uint8, shapeHash string) (inkRemaining uint32, err error)

	// Retrieves hashes contained by a specific block.
	// Can return the following errors:
	// - DisconnectedError
	// - InvalidBlockHashError
	GetShapes(blockHash string) (shapeHashes []string, err error)

	// Returns the block hash of the genesis block.
	// Can return the following errors:
	// - DisconnectedError
	GetGenesisBlock() (blockHash string, err error)

	// Retrieves the children blocks of the block identified by blockHash.
	// Can return the following errors:
	// - DisconnectedError
	// - InvalidBlockHashError
	GetChildren(blockHash string) (blockHashes []string, err error)

	// Closes the canvas/connection to the BlockArt network.
	// - DisconnectedError
	CloseCanvas() (inkRemaining uint32, err error)

	//Testing functions, to be deleted
	/*
			IsSvgStringValid(svgStr string) (isValid bool,Op shared.SingleOp)
		  IsSvgOutofBounds(svgOP shared.SingleOp) bool
	*/
}

//For Bonus mark:
// func GetListOfOps() []shared.FullSvgInfo {
// 	var resultArr []shared.FullSvgInfo
// 	resultArr = append(resultArr, shared.FullSvgInfo{
// 		Path:   "M 10 10 h 10 v 10 h -10 v -10",
// 		Fill:   "red",
// 		Stroke: "black"}) //square
// 	resultArr = append(resultArr, shared.FullSvgInfo{
// 		Path:   "M 100 100 l 400 400",
// 		Fill:   "transparent",
// 		Stroke: "red"}) //Kinked line,
//
// 	return resultArr
// }

// The constructor for a new Canvas object instance. Takes the miner's
// IP:port address string and a public-private key pair (ecdsa private
// key type contains the public key). Returns a Canvas instance that
// can be used for all future interactions with blockartlib.
//
// The returned Canvas instance is a singleton: an application is
// expected to interact with just one Canvas instance at a time.
//
// Can return the following errors:
// - DisconnectedError
func OpenCanvas(minerAddr string, privKey ecdsa.PrivateKey) (canvas Canvas, setting CanvasSettings, err error) {
	fmt.Print("OpenCanvas(): Going to connect to miner")

	// Connect to Miner
	art2MinerCon, err := rpc.Dial("tcp", minerAddr)
	CheckError(err)
	fmt.Println("Connected  to Miner")


	// Check if given key matches Miner key pair
	artNodeMsg := []byte("Art node wants to connect")
	r, s, err := ecdsa.Sign(cr.Reader, &privKey, artNodeMsg)
	CheckError(err)
	var ank = shared.ArtnodeVer{
		Msg: artNodeMsg,
		Ra:  r,
		Sa:  s,
	}


	// Art Node Sets up Listening Port for Miner
	artListenConn, err := net.Listen("tcp", "127.0.0.1:")
	CheckError(err)
	fmt.Println("Artnode going to be a listener")
	artNodeIp := artListenConn.Addr()
	fmt.Println("this is the artnode IP:")
	fmt.Println(artNodeIp.String())
	// mux := http.NewServeMux()
	// mux.HandleFunc("/getshapes", GetListOfOps)
	// // mux.HandleFunc("/addshape", inkMinerStruct.addshape)
	//
	// go http.ListenAndServe(artNodeIp.String(), mux)
	var thisCanvasObj CanvasObject
	thisCanvasObj.Ptr = new(CanvasObjectReal)
	thisCanvasObj.Ptr.ArtNodeipStr = artNodeIp.String()
	fmt.Println("Artnode going to be a listener 2")
	rpc.Register(&thisCanvasObj)
	go rpc.Accept(artListenConn)
	runtime.Gosched()
	fmt.Println("Artnode going to be a listener 3")

	// see if the Miner key matches the one you have
	var reply bool
	//Key := "test"
	fmt.Println("GOING TO CALL ARTNODEKEYCHECK")
	err = art2MinerCon.Call("KeyCheck.ArtNodeKeyCheck", ank, &reply)
	CheckError(err)
	if reply {
		fmt.Println("ArtNode has same key as miner")

		thisCanvasObj.Ptr.ArtNode.AmConn = art2MinerCon

		// Art node gets canvas settings from Miner node
		fmt.Println("ArtNode going to get settings from miner")
		// old
		initMs, err := thisCanvasObj.Ptr.ArtNode.GetCanvasSettings(artNodeIp.String()) // get the canvas settings, list of current operations(in initMS TODO), send listening ipAddress
		setting = CanvasSettings(initMs.Cs)
		thisCanvasObj.Ptr.ListOfOps_str = initMs.ListOfOps_str
		thisCanvasObj.Ptr.XYLimit = shared.Point{X: float64(setting.CanvasXMax), Y: float64(setting.CanvasYMax)}

		CheckError(err)

		return thisCanvasObj, setting, nil
	} else {
		fmt.Println("ArtNode does not have same key as miner")
		return nil, CanvasSettings{}, shared.DisconnectedError("")
	}

}

//Underlying struct holding the info for canvas
type CanvasObjectReal struct {
	ArtNode         shared.ArtNodeStruct
	ListOfOps_str   []string
	ListOfOps_ops   []shared.SingleOp
	LastPenPosition shared.Point
	XYLimit         shared.Point
	ArtNodeipStr    string

	// Canvas settings field?
}

type CanvasObject struct {
	Ptr *CanvasObjectReal
}

func (t CanvasObject) AddShape(validateNum uint8, shapeType shared.ShapeType, shapeSvgString string, fill string, stroke string) (shapeHash string, blockHash string, inkRemaining uint32, err error) {
	//Check for ShapeSvgStringTooLongError
	//var IsTransFill bool
	var isClosedCurve bool
	var isSvgValid bool
	//var svgOP shared.SingleOp
	var vtxArr []shared.Point
	var edgeArr []shared.LineSectVector
	var inkCost uint32

	if len(shapeSvgString) > 128 {
		return "", "", 0, shared.ShapeSvgStringTooLongError(shapeSvgString)
	}
	//Check for InValidSvg, OutofBound, SvgString too long errors
	switch shapeType {
	case shared.PATH:
		parsable, svgOP := shared.IsSvgStringParsable_Parse(shapeSvgString)
		if !parsable {
			return "", "", 0, shared.InvalidShapeSvgStringError(shapeSvgString)
		} else {
			isSvgValid, isClosedCurve, vtxArr, edgeArr = t.IsParsableSvgValid_GetVtxEdge(shapeSvgString, fill, stroke, svgOP)
			if !isSvgValid {
				return "", "", 0, shared.InvalidShapeSvgStringError(shapeSvgString + fill + stroke)
			}
		}
		if t.IsSvgOutofBounds(svgOP) {
			return "", "", 0, shared.OutOfBoundsError{}
		}
		inkCost = uint32(t.CalculateShapeArea(isClosedCurve, vtxArr, edgeArr, fill))
	case shared.CIRCLE:
		parsable, svgCirOP := shared.IsSvgStringParsable_Parse_Cir(shapeSvgString)
		if !parsable {
			return "", "", 0, shared.InvalidShapeSvgStringError(shapeSvgString)
		} else {
			isSvgValid = t.IsParsableSvgValid_Cir(shapeSvgString, fill, stroke, svgCirOP) // TODO
			if !isSvgValid {
				return "", "", 0, shared.InvalidShapeSvgStringError(shapeSvgString + fill + stroke)
			}
		}
		if t.IsSvgOutofBounds_Cir(svgCirOP) {
			return "", "", 0, shared.OutOfBoundsError{}
		}
		inkCost = t.CalculateShapeArea_Cir(svgCirOP, fill, stroke)

	default:
		return "", "", 0, err
	}

	//Create New OPERATION
	inkCost++ //For rounding up the cost
	fmt.Println("AddShape(),The command is ", shapeSvgString)
	newOP := shared.Operation{
		Command:        shapeSvgString,
		AmountOfInk:    inkCost,
		Shapetype:      shapeType,
		ShapeSvgString: shapeSvgString,
		Fill:           fill,
		Stroke:         stroke,
		ValidFBlkNum:   validateNum,
		Draw:           true,
	}
	addSuccess, err := t.Ptr.ArtNode.ArtnodeOp(newOP) // fn needs to return boolean
	if addSuccess {

		// ASK for the tree from Miner node
		blk, err := t.Ptr.ArtNode.GetBlockTreeFromMiner()
		BlockTree = blk
		infoBlocks := shared.GetLongestPathForArtNode(blk)
		shared.PrintBlockChainForArtNode(infoBlocks)
		BlockChain = shared.FilterBlockChain(infoBlocks)
		//TODO: Shape,Block,InkRemaining
		return "", "", 0, err
	} else {
		return "", "", 0, err
	}
}
func (t CanvasObject) GetSvgString(shapeHash string) (svgString string, err error) {
	svgString, err = t.Ptr.ArtNode.GetSvgStringUsingOperationSignature(shapeHash)
	return svgString, err

}
func (t CanvasObject) GetInk() (inkRemaining uint32, err error) {
	ink, err := t.Ptr.ArtNode.GetInkBalFromMiner()
	fmt.Println("GetInk() Ink of miner", ink)
	// get longest branch from miner compute ink based on how many signitures are from the miner
	return ink, err
}
func (t CanvasObject) DeleteShape(validateNum uint8, shapeHash string) (inkRemaining uint32, err error) {
	delOp, err := t.Ptr.ArtNode.GetOpWithHash(shapeHash) // get operation from the shapeHash
	CheckError(err)
	delOp.Draw = false
	delOp.ValidFBlkNum = validateNum
	// this node has to sign it --- make sure it is done in the miner
	_, err = t.Ptr.ArtNode.ArtnodeOp(delOp) // fn needs to return boolean
	if err != nil {
		return 0, err
	}
	inkRemaining, _ = t.GetInk()
	return inkRemaining, err
}
func (t CanvasObject) GetShapes(blockHash string) (shapeHashes []string, err error) {
	var s []string
	return s, nil
}
func (t CanvasObject) GetGenesisBlock() (blockHash string, err error) {
	gb, err := t.Ptr.ArtNode.GetGenesisBlockFromMiner()
	fmt.Println("GetGenesisBlock() Genesis blk hash", gb)
	return gb, err
}
func (t CanvasObject) GetChildren(blockHash string) (blockHashes []string, err error) {
	blockHashes, err = t.Ptr.ArtNode.GetChildrenFromMiner(blockHash)
	return blockHashes, err
}

func (t CanvasObject) CloseCanvas() (inkRemaining uint32, err error) {
	inkRemaining, _ = t.GetInk()
	err = t.Ptr.ArtNode.AmConn.Close()
	return inkRemaining, err

}

//varify If string is closed, return vtxArray and EdgeArray
func (t CanvasObject) IsParsableSvgValid_GetVtxEdge(svgStr string, fill string, stroke string, Op shared.SingleOp) (isValid bool, isClosed bool, vtxArray []shared.Point, edgeArray []shared.LineSectVector) {
	var vtxArr []shared.Point
	var edgeArr []shared.LineSectVector
	var isthisClosed bool
	// For Non-Transparent Fill, Must be closed
	if isthisClosed, vtxArr, edgeArr := shared.IsClosedShapeAndGetVtx(Op); !isthisClosed && fill != "transparent" {
		fmt.Println("Non-closed curve shape", svgStr, "but with fill:", fill)
		return false, isthisClosed, vtxArr, edgeArr
	}
	//No Fully Transparent Shape
	if (fill == "transparent" && stroke == "transparent") || (fill == "none" && stroke == "none") {
		return false, isthisClosed, vtxArr, edgeArr
	}
	// For Non-Transparent Fill,Must Not Be Self-Intersecting
	if isSelfInterSected := t.IsSelfIntersect(vtxArr, edgeArr); fill != "transparent" && isSelfInterSected {
		fmt.Println("Self intersected shape", svgStr, "but with fill:", fill)
		return false, isthisClosed, vtxArr, edgeArr
	}
	// Pass all tests:
	return true, isthisClosed, vtxArr, edgeArr
}
func (t CanvasObject) IsSvgOutofBounds(svgOP shared.SingleOp) bool {
	xVal := t.Ptr.LastPenPosition.X
	yVal := t.Ptr.LastPenPosition.Y
	for _, element := range svgOP.MovList {
		switch element.Cmd {
		case 'M', 'L', 'H', 'V':
			if element.X > t.Ptr.XYLimit.X || element.X < 0 || element.Y > t.Ptr.XYLimit.Y || element.Y < 0 {
				return true
			} else {
				xVal = xVal + element.X
				yVal = yVal + element.Y
			}
		case 'm', 'l', 'v', 'h':
			if element.X+xVal > t.Ptr.XYLimit.X || element.X+xVal < 0 || element.Y+yVal > t.Ptr.XYLimit.Y || element.Y+yVal < 0 {
				return true
			} else {
				xVal = xVal + element.X
				yVal = yVal + element.Y
			}
		default:
		}
	}
	return false
}

/*
func (t CanvasObject) ParseOpsStrings() {
	opsArrSize := len(t.ptr.ListOfOps_ops)
	for i, element := range t.ptr.ListOfOps_str {
		if valid, oneOp := t.IsSvgStringValid(element); valid {
			if i <= (opsArrSize - 1) {
				t.ptr.ListOfOps_ops[i] = oneOp
			} else {
				t.ptr.ListOfOps_ops = append(t.ptr.ListOfOps_ops, oneOp)
			}
		}
	}
}
*/

func (t CanvasObject) CalculateShapeArea(IsClosed bool, vtxArr []shared.Point, edgeArr []shared.LineSectVector, fill string) float64 {
	//Given the parsed results from IsClosedShapeAndGetVtx
	var area float64
	area = float64(0)
	if !IsClosed || fill == "transparent" {
		//Non-Closed Shape's Area is the summation of the
		for _, element := range edgeArr {
			area += Distance_TwoPoint(element.Start, element.End)
		}
	} else { //non-self-intersecting closed polygon,Checked by other functions inside AddShape
		area = Area_SingleClosedPolygon(vtxArr)
	}
	return area
}

//Input should be already checked as Closed shape
func (t CanvasObject) IsSelfIntersect(vtxArr []shared.Point, edgeArr []shared.LineSectVector) (IsSelfIntersect bool) {
	//O(n^2),Iteratively check any two line segment in the shape
	edgeCount := len(edgeArr)
	//Check for intersection in the middle of the line
	for i, _ := range edgeArr {
		for j := i + 1; j <= edgeCount; j++ {
			//For the i at last index, i will compare with the first element
			if j == edgeCount {
				if shared.TwoLineSegmentIntersected(edgeArr[i], edgeArr[0]) {
					return true
				}
			} else {
				if shared.TwoLineSegmentIntersected(edgeArr[i], edgeArr[j]) {
					return true
				}
			}
		}
	}
	//Check For one vertax more than three line type of intersection
	vtxCount := len(vtxArr)
	vtxEdgeCountArr := make([]int, vtxCount)
	for i, _ := range vtxArr {
		for j := i + 1; j < vtxCount; j++ {
			if vtxArr[i] == vtxArr[j] {
				vtxEdgeCountArr[i] += 1
				if vtxEdgeCountArr[i] > 1 {
					return true
				}
			}
		}
	}
	return false
}

func Distance_TwoPoint(x, y shared.Point) float64 {
	return math.Sqrt(math.Pow(x.X-y.X, 2) + math.Pow(x.Y-y.Y, 2))
}

//Only Handle Non-Intersecting Polygon. Self-Intersected shape should be sub divided into before calling
func Area_SingleClosedPolygon(vtxArr []shared.Point) float64 {
	//Algorithm Reference:https://www.mathopenref.com/coordpolygonarea.html
	vtxCount := len(vtxArr)
	if vtxCount <= 1 {
		return 0
	}
	firstVtx := vtxArr[0]
	var nextVtx shared.Point
	area := float64(0)
	for index, element := range vtxArr {
		if index >= vtxCount-1 {
			nextVtx = firstVtx
		} else {
			nextVtx = vtxArr[index+1]
		}
		area += 0.5 * (element.X*nextVtx.Y - element.Y*nextVtx.X)
	}
	return math.Abs(area)
}

func (t CanvasObject) ReceiveLongestChainFromMiner(chain []shared.FullSvgInfo, ack *bool) error {
	fmt.Println("receiving blockchain from miner")
	fmt.Println(chain)
	return nil
}

// Additional Helper
func CheckError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
	}
}

// **** Circle Functions
func (t CanvasObject) IsParsableSvgValid_Cir(svgStr string, fill string, stroke string, Op shared.CircleMov) bool {
	//var vtxArr []shared.Point
	//var edgeArr []shared.LineSectVector
	//var isthisClosed bool
	// For Non-Transparent Fill, Must be closed
	// if isthisClosed, vtxArr, edgeArr := shared.IsClosedShapeAndGetVtx(Op); !isClosed && fill != "transparent" {
	// 	fmt.Println("Non-closed curve shape", svgStr, "but with fill:", fill)
	// 	return false, isthisClosed, vtxArr, edgeArr
	// }
	//No Fully Transparent Shape
	if (fill == "transparent" && stroke == "transparent") || (fill == "none" && stroke == "none") {
		return false //, isthisClosed, vtxArr, edgeArr
	}
	// For Non-Transparent Fill,Must Not Be Self-Intersecting
	// if isSelfInterSected := t.IsSelfIntersect(vtxArr, edgeArr); fill != "transparent" && isSelfInterSected {
	// 	fmt.Println("Self intersected shape", svgStr, "but with fill:", fill)
	// 	return false, isthisClosed, vtxArr, edgeArr
	// }
	// Pass all tests:
	return true //, isthisClosed, vtxArr, edgeArr
}
func (t CanvasObject) IsSvgOutofBounds_Cir(OpCir shared.CircleMov) bool {
	return (OpCir.Cx+OpCir.R > t.Ptr.XYLimit.X) || (OpCir.Cx-OpCir.R > t.Ptr.XYLimit.X) || (OpCir.Cy+OpCir.R > t.Ptr.XYLimit.Y) || (OpCir.Cy-OpCir.R > t.Ptr.XYLimit.Y)

}
func (t CanvasObject) CalculateShapeArea_Cir(svgCirOp shared.CircleMov, fill string, stroke string) uint32 {
	if (fill == "none") || (fill == "transparent") {
		return uint32(2 * svgCirOp.R * math.Pi)
	} else {
		return uint32(math.Pi * math.Pow(svgCirOp.R, 2))
	}

}
