package shared

import (
	"fmt"
	"math"
)

//Compare if all points of one shape is inside another shape
//Assumption: Input shape's edge are non-intersecting which checked by func IsTwoEdgeArrInterSect()
//Assumption: Two fill cannot be both transparent
func IsOneShapeCompleteInsideAnother(newvtxArr []Point, newedgeArr []LineSectVector, newsvgFill string, newsvgArea uint32, tarvtxArr []Point, taredgeArr []LineSectVector, tarFill string, tarArea uint32, canvasMaxX uint32) bool {
	//Step1:Determine Which shape is the supposely Outside / Bigger one
	//var bigvtxArr *[]Point
	var bigedgeArr *[]LineSectVector
	var smallvtxArr *[]Point
	//var smalledgeArr *[]LineSectVector

	if newsvgFill == "transparent" { //the transparent has to be completely inside to be considered Overlapping
		smallvtxArr = &newvtxArr
		//smalledgeArr = &newedgeArr
		//bigvtxArr = &tarvtxArr
		bigedgeArr = &taredgeArr
	} else if tarFill == "transparent" {
		smallvtxArr = &tarvtxArr
		//smalledgeArr = &taredgeArr
		//bigvtxArr = &newvtxArr
		bigedgeArr = &newedgeArr
	} else {
		if newsvgArea >= tarArea {
			smallvtxArr = &tarvtxArr
			//smalledgeArr = &taredgeArr
			//bigvtxArr = &newvtxArr
			bigedgeArr = &newedgeArr
		} else {
			smallvtxArr = &newvtxArr
			//smalledgeArr = &newedgeArr
			//bigvtxArr = &tarvtxArr
			bigedgeArr = &taredgeArr
		}

	}

	//Step-2: Use Ray Casting Algorithm to Determine if small one is completely inside Big one
	IntersectionCountArr := make([]int, len(*smallvtxArr))
	for indexS, vtxS := range *smallvtxArr {
		rightRay := LineSectVector{Start: vtxS, End: Point{float64(canvasMaxX), vtxS.Y}} //build Ray from Vertex to the very right of the canvas
		for _, edgeBig := range *bigedgeArr {
			if TwoLineSegmentIntersected(rightRay, edgeBig) {
				IntersectionCountArr[indexS]++
			}
		}
	}
	//Analyze the Counting results
	//Since prerequist: small shape is either in or out but not cross on the edge,
	//The array is either all odd or all even.->We only need to sample the first element
	//Ray Casting Algorithm: Even Number of Intersection: Outside, Odd Number of Intersection:Inside
	if math.Mod(float64(IntersectionCountArr[0]), 2) == 0 {
		return false
	} else {
		return true
	}
}

//Compare if one edge in an edge array with any other edge in another edge array
func IsTwoEdgeArrInterSect(newedgeArr []LineSectVector, taredgeArr []LineSectVector) bool {
	isCurrentPairInterSected := false
	for _, edgeN := range newedgeArr {
		for _, edgeT := range taredgeArr {
			if isCurrentPairInterSected = TwoLineSegmentIntersected(edgeN, edgeT); isCurrentPairInterSected {
				fmt.Println("Two Intersecting Edges are ", edgeN, edgeT)
				return true
			}
		}
	}
	return false
}

//Geometric Function Functions
func TwoLineSegmentIntersected(lineSeg1 LineSectVector, lineSeg2 LineSectVector) bool {
	//Compare two line segment and check if they intersect at the
	x1s := lineSeg1.Start.X
	y1s := lineSeg1.Start.Y
	x1e := lineSeg1.End.X
	y1e := lineSeg1.End.Y
	x2s := lineSeg2.Start.X
	y2s := lineSeg2.Start.Y
	x2e := lineSeg2.End.X
	y2e := lineSeg2.End.Y
	a1 := (y1e - y1s) / (x1e - x1s)
	a2 := (y2e - y2s) / (x2e - x2s)
	b1 := (y1s*x1e - y1e*x1s) / (x1e - x1s)
	b2 := (y2s*x2e - y2e*x2s) / (x2e - x2s)
	if (a1 - a2) == 0 {
		return false
	}
	x_sln := (b2 - b1) / (a1 - a2)
	flpc := float64(0.1) // floating point compensation value
	fmt.Println("Line Floating Point Comparison Compensation is", flpc)
	//Check if x_sln is at around any vertex
	//0.1 is used for dealing with floating point inaccuracy
	if (minf(x1s, x1e)+flpc < x_sln) && (x_sln < maxf(x1s, x1e)-flpc) && (minf(x2s, x2e)+flpc < x_sln) && (x_sln < maxf(x2s, x2e)-flpc) {
		return true
	} else {
		return false
	}

}
func minf(x, y float64) float64 {
	if x > y {
		return y
	} else {
		return x
	}
}
func maxf(x, y float64) float64 {
	if x > y {
		return x
	} else {
		return y
	}
}
