package shared

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"regexp"
	"strconv"
	"time"
)

type BlockDepthpair struct {
	block, depth interface{}
}

func monitor(minerNeighbourAddr string, miner MinerStruct, heartBeatInterval time.Duration) {
	for {
		allNeighbour.Lock()

		if time.Now().UnixNano()-allNeighbour.all[minerNeighbourAddr].RecentHeartbeat > int64(heartBeatInterval) {
			log.Printf("%s timed out, walalalalala\n", allNeighbour.all[minerNeighbourAddr].MinerAddr)
			delete(allNeighbour.all, minerNeighbourAddr)

			if len(allNeighbour.all) < int(miner.Settings.MinNumMinerConnections) {
				miner.NotEnoughNeighbourSig <- true
			}
			allNeighbour.Unlock()

			return
		}
		log.Printf("%s is alive\n", allNeighbour.all[minerNeighbourAddr].MinerAddr)
		allNeighbour.Unlock()
		time.Sleep(heartBeatInterval)
	}
}

func filter(m *MinerStruct, visited *[]*MinerStruct) bool {
	for _, s := range *visited {
		if s.MinerAddr == m.MinerAddr {
			return false
		}
	}
	return true
}

func computeNonceSecretHash(nonce string, secret string) string {
	h := md5.New()
	h.Write([]byte(nonce + secret))
	str := hex.EncodeToString(h.Sum(nil))
	return str
}

func doProofOfWork(m *MinerStruct, nonce string, numberOfZeroes int, newOPs []Operation, leadingBlock *Block, isDoingWorkForNoOp bool) *Block {
	i := int64(0)

	var zeroesBuffer bytes.Buffer
	for i := int64(0); i < int64(numberOfZeroes); i++ {
		zeroesBuffer.WriteString("0")
	}

	zeroes := zeroesBuffer.String()

	fmt.Println("+++++++++++++++++++++++++++++++++++++++++++++++Begin Proof of work+++++++++++++++++++++++++++++")
	for {
		select {
		case recievedBlock := <-m.MiningStopSig:
			fmt.Println("Received block from another miner")
			return recievedBlock
		case opFromMineNode := <-m.RecievedOpSig:
			fmt.Println("M-UPDATED OPERATION LIST FROM MINERS ")
			if isDoingWorkForNoOp {
				nonce = leadingBlock.CurrentHash + opFromMineNode.ShapeSvgString + opFromMineNode.Fill + opFromMineNode.Stroke + fmt.Sprint(opFromMineNode.AmountOfInk) + pubKeyToString(m.PairKey.PublicKey)
				newOPs = []Operation{opFromMineNode}
				isDoingWorkForNoOp = false
				fmt.Println(nonce)
				fmt.Println("M-Was calculating No-op, now calculating Operation ", opFromMineNode.Command)
			} else {
				m.OPBuffer = append(m.OPBuffer, opFromMineNode)
				fmt.Println("M-UPDATED OPERATION LIST FROM MINERS " + AllOperationsCommands(m.OPBuffer))
			}
		case opFromArtnode := <-m.RecievedArtNodeSig:
			if isDoingWorkForNoOp {
				isDoingWorkForNoOp = false
				nonce = leadingBlock.CurrentHash + opFromArtnode.ShapeSvgString + opFromArtnode.Fill + opFromArtnode.Stroke + fmt.Sprint(opFromArtnode.AmountOfInk) + pubKeyToString(m.PairKey.PublicKey)
				newOPs = []Operation{opFromArtnode}
				fmt.Println(nonce)
				fmt.Println("A-Was calculating No-op, now calculating Operation ", opFromArtnode.Command)
			} else {
				m.OPBuffer = append(m.OPBuffer, opFromArtnode)
			}
			visitedMiners := []*MinerStruct{m}
			fmt.Println("A-UPDATED OPERATION LIST FROM ART NODE" + AllOperationsCommands(m.OPBuffer))
			m.FloodOperation(&opFromArtnode, &visitedMiners)
		default:
			guessString := strconv.FormatInt(i, 10)

			hash := computeNonceSecretHash(nonce, guessString)
			if hash[32-numberOfZeroes:] == zeroes {
				log.Println("Found the hash, it is: ", hash)
				log.Println(" NOUNCE IS " + nonce)
				m.FoundHash = true
				return m.produceBlock(hash, newOPs, leadingBlock, guessString)
			}
			i++
		}
	}
}

func pubKeyToString(key ecdsa.PublicKey) string {
	return string(elliptic.Marshal(key.Curve, key.X, key.Y))
}

func ParseBlockChain(thisBlock BlockPayloadStruct) *Block {
	x, y := elliptic.Unmarshal(elliptic.P384(), []byte(thisBlock.SolverPublicKey))
	if thisBlock.PreviousHash == "" {
		x = &big.Int{}
		y = &big.Int{}
	}
	// fmt.Println(thisBlock.SolverPublicKey)
	// fmt.Println(x)
	// fmt.Println(y)
	producedBlock := &Block{
		CurrentHash:       thisBlock.CurrentHash,
		PreviousHash:      thisBlock.PreviousHash,
		R:                 &thisBlock.R,
		S:                 &thisBlock.S,
		CurrentOPs:        thisBlock.CurrentOPs,
		DistanceToGenesis: thisBlock.DistanceToGenesis,
		Nonce:             thisBlock.Nonce,
		SolverPublicKey: &ecdsa.PublicKey{
			Curve: elliptic.P384(),
			X:     x,
			Y:     y,
		},
	}
	var producedBlockChilden []*Block
	for _, child := range thisBlock.Children {
		producedBlockChilden = append(producedBlockChilden, ParseBlockChain(child))
	}
	producedBlock.Children = producedBlockChilden
	// fmt.Println("finshed copying the chain, the current hash is: ", producedBlock.CurrentHash)
	return producedBlock
}

func CopyBlockChainPayload(thisBlock *Block) BlockPayloadStruct {
	producedBlockPayload := BlockPayloadStruct{
		CurrentHash:       thisBlock.CurrentHash,
		PreviousHash:      thisBlock.PreviousHash,
		R:                 *thisBlock.R,
		S:                 *thisBlock.S,
		CurrentOPs:        thisBlock.CurrentOPs,
		DistanceToGenesis: thisBlock.DistanceToGenesis,
		Nonce:             thisBlock.Nonce,
		SolverPublicKey:   pubKeyToString(*thisBlock.SolverPublicKey),
	}
	var producedBlockChilden []BlockPayloadStruct
	for _, child := range thisBlock.Children {
		producedBlockChilden = append(producedBlockChilden, CopyBlockChainPayload(child))
	}
	producedBlockPayload.Children = producedBlockChilden
	// fmt.Println("finshed copying the chain, the current hash is: ", producedBlock.CurrentHash)
	return producedBlockPayload
}

func findDeepestBlocks(b *Block, depth int) (*Block, int) {
	if len(b.Children) == 0 {
		return b, depth
	} else {
		childrenDepth := make([]BlockDepthpair, 0)
		// Find depth of all children
		for _, child := range b.Children {
			block, depth := findDeepestBlocks(child, depth+1)
			childrenDepth = append(childrenDepth, BlockDepthpair{block, depth})
		}
		localMax := 0
		localMaxBlock := &Block{}

		// Pick the deepest blck from children
		for _, tuple := range childrenDepth {
			if tuple.depth.(int) > localMax {
				localMax = tuple.depth.(int)
				localMaxBlock = tuple.block.(*Block)
			}
		}
		return localMaxBlock, localMax
	}
}

func GetLongestPathForArtNode(b *Block) []InfoBlock {
	if b == nil {
		return nil
	}

	if len(b.Children) == 0 {
		tmpInfoBlock := InfoBlock{ListOperations: b.CurrentOPs}
		return []InfoBlock{tmpInfoBlock}
	} else {
		longestBlockChain := make([]InfoBlock, 0)

		deepestBlock, _ := findDeepestBlocks(b, 0)
		tmpInfoBlock := InfoBlock{ListOperations: deepestBlock.CurrentOPs, CurrentHash: deepestBlock.CurrentHash, PreviousHash: deepestBlock.PreviousHash}
		longestBlockChain = append(longestBlockChain, tmpInfoBlock)
		nthBlock := deepestBlock

		for nthBlock.PreviousHash != "" {

			foundBlock := findBlockUsingHash(nthBlock.PreviousHash, b)
			tmpInfoBlock := InfoBlock{ListOperations: foundBlock.CurrentOPs, PreviousHash: foundBlock.PreviousHash, CurrentHash: foundBlock.CurrentHash}

			longestBlockChain = append(longestBlockChain, tmpInfoBlock)
			nthBlock = foundBlock
		}

		return longestBlockChain
	}
}

func FilterBlockChain(prevChain []InfoBlock) []FullSvgInfo {

	filteredChain := make([]FullSvgInfo, 0)

	for _, b := range prevChain {

		for _, operation := range b.ListOperations {
			if operation.ShapeSvgString == "no-op" {
				continue
			}
			tmpInfoBlock := FullSvgInfo{}

			tmpInfoBlock.Fill = operation.Fill
			tmpInfoBlock.Path = operation.ShapeSvgString
			tmpInfoBlock.Stroke = operation.Stroke
			filteredChain = append(filteredChain, tmpInfoBlock)
		}
	}

	return filteredChain

}

func getLongestPath(b *Block) []Block {
	if b == nil {
		return nil
	}

	if len(b.Children) == 0 {
		return []Block{*b}
	} else {
		longestBlockChain := make([]Block, 0)
		deepestBlock, _ := findDeepestBlocks(b, 0)

		longestBlockChain = append(longestBlockChain, *deepestBlock)
		nthBlock := deepestBlock

		for nthBlock.PreviousHash != "" {

			foundBlock := findBlockUsingHash(nthBlock.PreviousHash, b)

			longestBlockChain = append(longestBlockChain, *foundBlock)
			nthBlock = foundBlock
		}

		return longestBlockChain
	}
}

func PrintBlock(m *Block) {

	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	if m.PreviousHash == "" {
		strconv.Itoa(len(m.Children))
		fmt.Println("Genesis BLOCK:    " + m.CurrentHash[0:5] + " has this " + strconv.Itoa(len(m.Children)) + " children")
	} else {
		fmt.Println(m.CurrentHash[0:5] + " has this " + strconv.Itoa(len(m.Children)) + " children and its parent is " + m.PreviousHash[0:5])
	}
	for _, c := range m.Children {
		PrintBlock(c)
	}
}

func PrintBlockChainForArtNode(chain []InfoBlock) {

	fmt.Printf("DEEPEST NODE ")
	for _, b := range chain {
		fmt.Printf(b.CurrentHash[0:5] + " -> ")
	}
	fmt.Printf(" GENESIS")

	fmt.Println()
}

func printBlockChain(chain []Block) {

	fmt.Printf("DEEPEST NODE ")
	for _, b := range chain {
		fmt.Printf(b.CurrentHash[0:5] + " -> ")
	}
	fmt.Printf(" GENESIS")

	fmt.Println()
}

func getBlockDistanceFromGensis(blk *Block, blkHash string) int {

	if blk == nil {
		return -1
	}

	if blk.CurrentHash == blkHash {
		return 0
	}

	if len(blk.Children) == 0 {
		return -1
	}

	distArray := make([]int, len(blk.Children))
	for i, subB := range blk.Children {
		distArray[i] = getBlockDistanceFromGensis(subB, blkHash)
	}

	return 1 + maxArray(distArray)
}

func maxArray(array []int) int {

	if len(array) == 0 {
		return 0
	}

	maxNum := array[0]

	for _, num := range array {

		if num > maxNum {
			maxNum = num
		}
	}
	return maxNum
}

func findBlockUsingHash(hash string, blk *Block) *Block {

	if blk.CurrentHash == hash {
		return blk
	}

	for _, subBlock := range blk.Children {
		result := findBlockUsingHash(hash, subBlock)
		if result != nil {
			return result
		}
	}

	return nil
}

//Parse the Svg string if it's parsable
func IsSvgStringParsable_Parse(svgStr string) (isValid bool, Op SingleOp) {
	//Legal Example: "m 20 0 L 19 21",all separated by space,always start at m/M
	strCnt := len(svgStr)
	var movList []SingleMov
	parsedOp := SingleOp{MovList: movList}
	if strCnt < 3 {
		return false, parsedOp
	}
	if svgStr[0] != 'M' {
		return false, parsedOp
	}
	regex_2 := regexp.MustCompile("([mMvVlLhHZz])[\\s]([-]*[0-9]+)[\\s]([-]*[0-9]+)")
	regex_1 := regexp.MustCompile("([mMvVlLhHZz])[\\s]([-]*[0-9]+)")
	var matches []string
	var oneMov SingleMov
	var thisRune rune
	fmt.Println("string size is", strCnt)
	for i := 0; i < strCnt; i = i {
		fmt.Println("Current I is", i)
		thisRune = rune(svgStr[i])
		switch thisRune {
		case 'm', 'M', 'L', 'l':
			arr := regex_2.FindStringIndex(svgStr[i:])
			if arr == nil {
				return false, parsedOp
			} else {
				//if legal, Parse it
				matches = regex_2.FindStringSubmatch(svgStr[i:])
				intVal1, _ := strconv.Atoi(matches[2])
				intVal2, _ := strconv.Atoi(matches[3])
				oneMov = SingleMov{Cmd: rune(thisRune), X: float64(intVal1), Y: float64(intVal2), ValCnt: 2}
				parsedOp.MovList = append(parsedOp.MovList, oneMov)
				//Update Index
				fmt.Println("ML update next index is", arr[0], arr[1])
				i = i + arr[1] + 1
			}
		case 'H', 'h':
			arr := regex_1.FindStringIndex(svgStr[i:])
			fmt.Println("VH update next index is", arr[0], arr[1])
			if arr == nil {
				return false, parsedOp
			} else {
				matches = regex_1.FindStringSubmatch(svgStr[i:])
				intVal1, _ := strconv.Atoi(matches[2])
				oneMov = SingleMov{Cmd: rune(thisRune), X: float64(intVal1), Y: 0, ValCnt: 1}
				parsedOp.MovList = append(parsedOp.MovList, oneMov)
				i = i + arr[1] + 1
			}
		case 'v', 'V':
			arr := regex_1.FindStringIndex(svgStr[i:])
			fmt.Println("VH update next index is", arr[0], arr[1])
			if arr == nil {
				return false, parsedOp
			} else {
				matches = regex_1.FindStringSubmatch(svgStr[i:])
				intVal1, _ := strconv.Atoi(matches[2])
				oneMov = SingleMov{Cmd: rune(thisRune), X: 0, Y: float64(intVal1), ValCnt: 1}
				parsedOp.MovList = append(parsedOp.MovList, oneMov)
				i = i + arr[1] + 1
			}

		case 'Z', 'z':
			oneMov := SingleMov{Cmd: rune(thisRune), ValCnt: 0}
			parsedOp.MovList = append(parsedOp.MovList, oneMov)
			i = i + 2
		default:
			return false, parsedOp
		}
	}
	return true, parsedOp //pass all tests
}

func IsClosedShapeAndGetVtx(op SingleOp) (IsClosed bool, vtxArray []Point, edgeArray []LineSectVector) {
	var vtxArr []Point
	var edgeArr []LineSectVector
	var curVtx Point
	var preVtx Point
	var nextVtx Point
	var originalStart Point
	var lastSubPathStart Point
	//traverse all operation, identify list of edge and points
	//TODO : Corner Case when an open shape has the same ending point
	movCount := len(op.MovList)
	if movCount < 1 {
		return false, vtxArr, edgeArr
	} //Panic Check
	originalStart.X = op.MovList[0].X // Assume the first mov is always 'M' which is validated by IsValidSvgString
	originalStart.Y = op.MovList[0].Y
	for _, element := range op.MovList {
		switch element.Cmd {
		case 'M', 'L':
			preVtx = curVtx
			curVtx.X = element.X
			curVtx.Y = element.Y
			if element.Cmd != 'M' {
				edgeArr = append(edgeArr, LineSectVector{preVtx, curVtx}) //add new line segment
			} else {
				lastSubPathStart = curVtx // prepare for potential Z/z command
			}
			vtxArr = append(vtxArr, curVtx) // add new vertex
		case 'V':
			preVtx = curVtx
			curVtx.Y = element.Y
			if element.Cmd != 'M' {
				edgeArr = append(edgeArr, LineSectVector{preVtx, curVtx}) //add new line segment
			} else {
				lastSubPathStart = curVtx // prepare for potential Z/z command
			}
			vtxArr = append(vtxArr, curVtx) // add new vertex
		case 'H':
			preVtx = curVtx
			curVtx.X = element.X
			if element.Cmd != 'M' {
				edgeArr = append(edgeArr, LineSectVector{preVtx, curVtx}) //add new line segment
			} else {
				lastSubPathStart = curVtx // prepare for potential Z/z command
			}
			vtxArr = append(vtxArr, curVtx) // add new vertex
		case 'm', 'v', 'h', 'l':
			preVtx = curVtx
			curVtx.X += element.X
			curVtx.Y += element.Y
			if element.Cmd != 'm' {
				edgeArr = append(edgeArr, LineSectVector{preVtx, curVtx})
			} else {
				lastSubPathStart = curVtx
			}
			vtxArr = append(vtxArr, curVtx)
		case 'Z', 'z':
			preVtx = curVtx
			curVtx = lastSubPathStart
			edgeArr = append(edgeArr, LineSectVector{preVtx, curVtx})
		}
	}
	//List through the edge array and identify if everything is connected.Reuse variables
	if len(edgeArr) < 1 {
		return false, vtxArr, edgeArr
	}
	preVtx = edgeArr[0].Start
	for _, element := range edgeArr { // Check For discontinuity
		curVtx = element.Start
		nextVtx = element.End
		if curVtx != preVtx {
			return false, vtxArr, edgeArr
		} else {
			//vtxArr = append(vtxArr, curVtx)
			//vtxArr = append(vtxArr, nextVtx)
			preVtx = nextVtx
		}
	}
	//If entire edge is continous, Check if it returns to the same points
	if nextVtx != originalStart {
		return false, vtxArr, edgeArr
	} else {
		uniqueVtxCount := len(vtxArr) - 1
		return true, vtxArr[:uniqueVtxCount], edgeArr // the last "nextVtx" will be an overlapping of the staring point
	}
}

func IsSvgStringParsable_Parse_Cir(svgStr string) (isValid bool, Op CircleMov) {
	// Valid svgString for circle: "cx 100 cy r 6"
	regex_1 := regexp.MustCompile("cx[\\s]([-]*[0-9]+)")
	regex_2 := regexp.MustCompile("cy[\\s]([-]*[0-9]+)")
	regex_3 := regexp.MustCompile("r[\\s]([-]*[0-9]+)")

	var mov CircleMov
	var matches []string
	var thisRune rune
	strCnt := len(svgStr)

	if strCnt < 3 {
		return false, mov
	}

	for i := 0; i < strCnt; i = i {
		thisRune = rune(svgStr[i])
		fmt.Println(thisRune)
		switch thisRune {
		case 'r':
			//fmt.Println("r expression ", regex_3)
			arr := regex_3.FindStringIndex(svgStr[i:])
			if arr == nil {
				return false, mov
			} else {
				matches = regex_3.FindStringSubmatch(svgStr[i:])
				rVal, _ := strconv.Atoi(matches[1])
				mov.R = float64(rVal)
				fmt.Println(mov)
			}
			i = strCnt
		case 'c':
			if rune(svgStr[i+1]) == 'x' {
				//fmt.Println("cx expression ", regex_1)
				arr := regex_1.FindStringIndex(svgStr[i:])
				if arr == nil {
					return
				} else {
					matches = regex_1.FindStringSubmatch(svgStr[i:])
					xVal, _ := strconv.Atoi(matches[1])
					mov.Cx = float64(xVal)
					i = i + arr[1] + 1
				}
			} else {
				//fmt.Println("cy expression ", regex_2)
				arr := regex_2.FindStringIndex(svgStr[i:])
				if arr == nil {
					return
				} else {
					matches = regex_2.FindStringSubmatch(svgStr[i:])
					yVal, _ := strconv.Atoi(matches[1])
					mov.Cy = float64(yVal)
					i = i + arr[1] + 1
				}
			}
		default:
			return false, mov
			fmt.Println("IsSvgStringParsable_Parse_Cir() Invalid parse")
		}

	}
	if (mov.Cx >= 0) && (mov.Cy >= 0) && (mov.R > 0) {
		return true, mov
	}
	fmt.Println("IsSvgStringParsable_Parse_cir() circle feils are neg")
	return false, mov

}
