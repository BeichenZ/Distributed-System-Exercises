/*

A trivial application to illustrate how the blockartlib library can be
used from an application in project 1 for UBC CS 416 2017W2.

Usage:
go run art-app.go
*/

package main

// Expects blockartlib.go to be in the ./blockartlib/ dir, relative to
// this art-app.go file
import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"./blockartlib"
	shared "./shared"
	// shared "./shared"
	//"encoding/gob"
	"bufio"
	"encoding/gob"
	"crypto/x509"
	"encoding/hex"

)

var globalCanvas blockartlib.Canvas

func GetListOfOps(w http.ResponseWriter, r *http.Request) {
	//longestChain := getLongestPath(m.BlockChain)
	//resultArr := make([]FullSvgInfo, 0)
	//for _, block := range longestChain {
	//	for _, op := range block.CurrentOPs {
	//		resultArr = append(resultArr, FullSvgInfo{
	//			Path:   op.ShapeSvgString,
	//			Fill:   op.Fill,
	//			Stroke: op.Stroke,
	//		})
	//	}
	//}
	//var resultArr []shared.FullSvgInfo
	//resultArr = append(resultArr, shared.FullSvgInfo{
	//	Path:   "M 10 10 h 10 v 10 h -10 v -10",
	//	Fill:   "red",
	//	Stroke: "black"}) //square
	//resultArr = append(resultArr, shared.FullSvgInfo{
	//	Path:   "M 100 100 l 400 400",
	//	Fill:   "transparent",
	//	Stroke: "red"}) //Kinked line,
	var response []shared.FullSvgInfo
	if len(blockartlib.BlockChain) == 0 {
		response = make([]shared.FullSvgInfo, 0)
	} else {
		response = blockartlib.BlockChain
	}
	s, err := json.Marshal(response)
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set(
		"Access-Control-Allow-Headers",
		"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization",
	)
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.WriteHeader(http.StatusOK)
	//Write json response back to response
	w.Write(s)
}

func ArtNodeAddshape(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set(
		"Access-Control-Allow-Headers",
		"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization",
	)
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.WriteHeader(http.StatusOK)
	if r.Method == "POST" {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body", http.StatusInternalServerError)
		}
		globalCanvas.AddShape(2, shared.CIRCLE, "cx 10 cy 10 r 8", "transparent", "red")
		fmt.Println(string(body))
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

func main() {
	//minerAddrP := flag.String("ma", "MinerAddr Missing", "a string")
	minerAddrP := os.Args[1]
	minerPrivateKey := os.Args[2]
	//minerPrivateKey := flag.String("mp", "minerPrivateKey missing", "a string")
	//fmt.Println(*minerPrivateKey[0:1] == " ")
	gob.Register(&elliptic.CurveParams{})

	//theShit := "3081a4020101043065cb964b29bee4f4e35bceb26d6e1c48903470706d52c3b79ce9c20da9aa9a60cff4f0302f7acc9177636962b5b9f89ea00706052b81040022a164036200045acf02cb504f75dd04275d04c17033f3eb9f4af1c33b07adcf904bc80e3665a9258579f079ac432a75afce690c52a07b50cd35b1c4a882d0e2b683f97e6d50bb58b6d5c8a4b02d7e05905f4398c28675a6974664e7e84adc217249821419019a"

	//fmt.Println(strings.TrimSpace(*minerPrivateKey))
	//fmt.Println(theShit)

	privKey := convertKeyPair(minerPrivateKey) // TODO: use crypto/ecdsa to read pub/priv keys from a file argument.

	//minerAddr := "127.0.0.1:39865" // hardcoded for now
	// Remove later
	flag.Parse()
	fmt.Println("command line arguments ", minerAddrP, minerPrivateKey, minerAddrP)
	fmt.Println(privKey)
	// 	// Open a canvas.
	canvas, settings, err := blockartlib.OpenCanvas(minerAddrP, *privKey)
	globalCanvas = canvas
	pieceOfShit := canvas.(blockartlib.CanvasObject)
	fmt.Printf("%+v", pieceOfShit.Ptr)

	artAddres := pieceOfShit.Ptr.ArtNodeipStr
	// port := artNodeIp.String()[strings.Index(artNodeIp.String(), ":"):len(artNodeIp.String())]
	port := artAddres[strings.Index(artAddres, ":"):len(artAddres)]
	fmt.Println(port)
	mux := http.NewServeMux()
	mux.HandleFunc("/getshapes", GetListOfOps)
	mux.HandleFunc("/addshape", ArtNodeAddshape)

	go http.ListenAndServe(":5000", mux)

	if checkError(err) != nil {
		fmt.Println(err.Error())
		return
	}
	//For testing,Can be deleted
	//  isOpvalid,testOp := canvas.IsSvgStringValid("m 100 100 l 500 400 l 1000 2000")
	//	isOutofBound := canvas.IsSvgOutofBounds(testOp)
	//	fmt.Println("operation first second third",isOpvalid,string(testOp.MovList[0].Cmd),string(testOp.MovList[1].Cmd))
	//	fmt.Println("Operation is out of bound!:",isOutofBound)

	validateNum := 2
	fmt.Println("remove after", canvas, settings, validateNum)

	// Getter method checks
	fmt.Println("remove after", canvas, settings, validateNum)
	ink, err := canvas.GetInk()
	fmt.Println("art-app.main(): going to get ink from miner", ink, "   ", err)
	gb, err := canvas.GetGenesisBlock()
	fmt.Println("art-app.main(): going to get genesis block from miner", gb, "   ", err)

	// Add a line.

	for {
		buf := bufio.NewReader(os.Stdin)
		var svgString string
		var fill string
		var color string
		var path shared.ShapeType
		fmt.Println("> Press A : Add Shape")
		sentence, err := buf.ReadByte()
		if err != nil {
			fmt.Println(err)
		} else {
			command := string(sentence)
			fmt.Println(command == "A")
			if command == "A" {
				buf := bufio.NewReader(os.Stdin)
				fmt.Println("       > Enter the SVG string")

				sentence, _, _ := buf.ReadLine()

				if err != nil {
					fmt.Println(err)
				} else {

					svgString = string(sentence)
					buf := bufio.NewReader(os.Stdin)
					fmt.Println("       > Enter fill")

					sentence, _, _ := buf.ReadLine()

					if err != nil {
						fmt.Println(err)
					} else {
						fill = string(sentence)
						buf := bufio.NewReader(os.Stdin)
						fmt.Println("       > Enter Color")

						sentence, _, _ := buf.ReadLine()

						if err != nil {
							fmt.Println(err)
						} else {

							color = string(sentence)
							buf := bufio.NewReader(os.Stdin)
							fmt.Println("      > Enter P for Path or C for Circle")
							sentence, _ := buf.ReadByte()

							if err != nil {
								fmt.Println(err)
							} else {
								if string(sentence) == "P" {
									path = shared.PATH
								} else if string(sentence) == "C" {
									path = shared.CIRCLE
								} else {
									fmt.Println("Incorrect options")
								}
							}

							fmt.Println("DRAWING ======================")
							_, _, _, err = canvas.AddShape(2, path, svgString, fill, color)
							if err != nil {
								fmt.Println(err)
							}

						}
					}

				}

			}
		}

	}

	fmt.Println("ADDING SHAPES+++++")

	// _, _, _, err = canvas.AddShape(2, blockartlib.PATH, "M 0 0 l 10 10", "transparent", "red")
	// _, _, _, err = canvas.AddShape(2, blockartlib.PATH, "M 2 9 l 10 10", "transparent", "blue")
	// _, _, _, err = canvas.AddShape(2, blockartlib.PATH, "M 20 90 l 10 10", "transparent", "green")
	// _, _, _, err = canvas.AddShape(2, blockartlib.PATH, "M 21 98 l 10 10", "transparent", "black")

	if checkError(err) != nil {
		return
	}

	// Add another line.
	//	shapeHash2, blockHash2, ink2, err := canvas.AddShape(validateNum, blockartlib.PATH, "M 0 0 L 5 0", "transparent", "blue")
	if checkError(err) != nil {
		return
	}

	// Delete the first line.
	//	ink3, err := canvas.DeleteShape(validateNum, shapeHash)
	if checkError(err) != nil {
		return
	}

	// assert ink3 > ink2

	// Close the canvas.
	//	ink4, err := canvas.CloseCanvas()
	if checkError(err) != nil {
		return
	}
}

// If error is non-nil, print it out and return it.
func checkError(err error) error {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error ", err.Error())
		return err
	}
	return nil
}

// Helper functions

// gets the key pair given the public key of the miner --- change***
func convertKeyPair(key string) *ecdsa.PrivateKey {
	privateKeyBytesRestored, _ := hex.DecodeString(key)
	AmPrivateKeyPair, _ := x509.ParseECPrivateKey(privateKeyBytesRestored)
	return AmPrivateKeyPair
}