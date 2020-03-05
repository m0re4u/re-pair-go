package systems

import (
	"math"
	"sync"

	"github.com/EngoEngine/engo"
)

const discreteStep = 8

// Point use the engo Point object
type Point struct {
	X, Y int
}

// Convert engo point to pathing point (more discrete)
func Convert(p engo.Point) Point {
	x := p.X
	y := p.Y
	return Point{X: int(x) / discreteStep, Y: int(y) / discreteStep}
}

// ConvertBack pathing point to engo point
func ConvertBack(p Point) engo.Point {
	x := p.X
	y := p.Y
	return engo.Point{X: (float32(x) * discreteStep) + 4, Y: (float32(y) * discreteStep) + 4}
}

// AStar interface for the A* implementation
type AStar interface {
	// Fill a given tile with a given weight this is used for making certain areas more complicated
	// to cross than others. For example you may have a higher weight for a wall or mountain.
	// This weight will be given back to you in the SetWeight function
	// Inbuilt A*'s use -1 to determine that it can not be passed at all.
	FillTile(p Point, weight int)

	// Resets the weight back to 0 for a given tile
	ClearTile(p Point)

	// Calculate the easiest path from ANY element in source to ANY element in target.
	// There is no hard rules about which element will become the start and end (unless your config
	// enforces it).
	// The start of the path is returned to you. If no path exists then the function will
	// return nil as the path.
	FindPath(config AStarConfig, source, target []Point) *PathPoint
}

// AStarConfig The user built configuration that determines how weights are calculated and
// also determines the stopping condition
type AStarConfig interface {
	// Determine if a valid end point has been reached. The end parameter
	// is the value passed in as source because the algorithm works backwards.
	IsEnd(p Point, end []Point, endMap map[Point]bool) bool

	// Calculate and set the weight for p.
	// fillWeight is the weight assigned to the tile when FillTile was called
	// or 0 if it was never called for that tile.
	// end is also provided so you can perform calculations such as distance remaining.
	SetWeight(p *PathPoint, fillWeight int, end []Point, endMap map[Point]bool) (allowed bool)

	// PostProcess the path after it has been calculated this might be useful if you want do do things
	// like reverse it or add constant moves at the start or end etc.
	PostProcess(p *PathPoint, rows, cols int, filledTiles map[Point]int) *PathPoint
}

type gridStruct struct {
	// A list of filled tiles and their weight
	tileLock    sync.Mutex
	filledTiles map[Point]int

	rows int
	cols int
}

// NewAStar new A*
func NewAStar(rows, cols int) AStar {
	return &gridStruct{
		rows: rows,
		cols: cols,

		filledTiles: make(map[Point]int),
	}
}

func (a *gridStruct) FillTile(p Point, weight int) {
	a.tileLock.Lock()
	defer a.tileLock.Unlock()

	a.filledTiles[p] = weight
}

func (a *gridStruct) ClearTile(p Point) {
	a.tileLock.Lock()
	defer a.tileLock.Unlock()

	delete(a.filledTiles, p)
}

func (a *gridStruct) FindPath(config AStarConfig, source, target []Point) *PathPoint {
	var openList = make(map[Point]*PathPoint)
	var closeList = make(map[Point]*PathPoint)

	sourceMap := make(map[Point]bool)
	for _, p := range source {
		sourceMap[p] = true
	}

	a.tileLock.Lock()
	for _, p := range target {
		fillWeight := a.filledTiles[p]
		pathPoint := &PathPoint{
			Point:        p,
			Parent:       nil,
			DistTraveled: 0,
			FillWeight:   fillWeight,
		}

		allowed := config.SetWeight(pathPoint, fillWeight, source, sourceMap)
		if allowed {
			openList[p] = pathPoint
		}
	}

	a.tileLock.Unlock()

	var current *PathPoint
	for {
		current = a.getMinWeight(openList)

		a.tileLock.Lock()
		if current == nil || config.IsEnd(current.Point, source, sourceMap) {
			a.tileLock.Unlock()
			break
		}
		a.tileLock.Unlock()

		delete(openList, current.Point)
		closeList[current.Point] = current

		surrounding := a.getSurrounding(current.Point)

		for _, p := range surrounding {
			_, ok := closeList[p]
			if ok {
				continue
			}

			a.tileLock.Lock()
			fillWeight := a.filledTiles[p]
			a.tileLock.Unlock()

			pathPoint := &PathPoint{
				Point:        p,
				Parent:       current,
				FillWeight:   current.FillWeight + fillWeight,
				DistTraveled: current.DistTraveled + 1,
			}

			a.tileLock.Lock()
			allowed := config.SetWeight(pathPoint, fillWeight, source, sourceMap)
			a.tileLock.Unlock()

			if !allowed {
				continue
			}

			existingPoint, ok := openList[p]
			if !ok {
				openList[p] = pathPoint
			} else {
				if pathPoint.Weight < existingPoint.Weight {
					existingPoint.Parent = pathPoint.Parent
				}
			}
		}
	}

	a.tileLock.Lock()
	current = config.PostProcess(current, a.rows, a.cols, a.filledTiles)
	a.tileLock.Unlock()

	return current
}

func (a *gridStruct) getMinWeight(openList map[Point]*PathPoint) *PathPoint {
	var min *PathPoint = nil
	var minWeight int = 0

	for _, p := range openList {
		if min == nil || p.Weight < minWeight {
			min = p
			minWeight = p.Weight
		}
	}
	return min
}

func (a *gridStruct) getSurrounding(p Point) []Point {
	var surrounding []Point

	row, col := int(p.X), int(p.Y)

	if row > 0 {
		surrounding = append(surrounding, Point{row - 1, col})
	}
	if row < a.rows-1 {
		surrounding = append(surrounding, Point{row + 1, col})
	}

	if col > 0 {
		surrounding = append(surrounding, Point{row, col - 1})
	}
	if col < a.cols-1 {
		surrounding = append(surrounding, Point{row, col + 1})
	}

	return surrounding
}

// PathPoint A point along a path.
// FillWeight is the sum of all the fill weights so far and
// DistTraveled is the total distance traveled so far
//
// WeightData is an interface that can be set to anything that Config wants
// it will never be touched by the rest of the code but if you wish to
// have any custom data held per node you can use WeightData
type PathPoint struct {
	Point
	Parent *PathPoint

	Weight       int
	FillWeight   int
	DistTraveled int

	WeightData interface{}
}

// Dist Manhattan distance NOT euclidean distance because in our routing we cant go diagonally between the points.
func (p Point) Dist(other Point) int {
	return int(math.Abs(float64(p.X-other.X)) + math.Abs(float64(p.Y-other.Y)))
}

//######################################################################
//######################################################################

type pointToPoint struct {
	VoidPostProcess
}

// Basic point to point routing, only a single source
// is supported and it will panic if given multiple sources
//
// Weights are calulated by summing the tiles fill_weight, the total distance traveled
// and the current distance from the target
func NewPointToPoint() AStarConfig {
	p2p := &pointToPoint{}

	return p2p
}

func (p2p *pointToPoint) SetWeight(p *PathPoint, fill_weight int, end []Point, end_map map[Point]bool) bool {
	if len(end) != 1 {
		panic("Invalid end specified")
	}

	if fill_weight == -1 {
		return false
	}

	p.Weight = p.FillWeight + p.DistTraveled + p.Point.Dist(end[0])

	return true
}

func (p2p *pointToPoint) IsEnd(p Point, end []Point, end_map map[Point]bool) bool {
	if len(end) != 1 {
		panic("Invalid end specified")
	}
	return p == end[0]
}

//######################################################################
//######################################################################

type rowToRow struct {
	VoidPostProcess
}

// Based off the PointToPoint config except that it uses row based targeting.
// The column value is ignored when calculating the weight and when determining
// if we have reached the end.
//
// A single point should be given for the source which will determine the starting row.
// for the target you should provide every valid entry on the target row for the best results.
// you do not have to but the path may look a little strange sometimes.
func NewRowToRow() AStarConfig {
	r2r := &rowToRow{}
	return r2r
}

func (r2r *rowToRow) SetWeight(p *PathPoint, fill_weight int, end []Point, end_map map[Point]bool) bool {
	if len(end) != 1 {
		panic("Invalid end specified")
	}

	if fill_weight == -1 {
		return false
	}

	p.Weight = p.FillWeight + p.DistTraveled + int(math.Abs(float64(p.X-end[0].X)))

	return true
}

func (r2r *rowToRow) IsEnd(p Point, end []Point, end_map map[Point]bool) bool {
	if len(end) != 1 {
		panic("Invalid end specified")
	}
	return p.X == end[0].X
}

//######################################################################
//######################################################################

type listToPoint struct {
}

type listToPointForward struct {
	listToPoint
	VoidPostProcess
}

type listToPointReverse struct {
	listToPoint
	ReversePostProcess
}

// list to point routing, from a list of points to a single point.
// multiple targets are supported but is slower than the others.
//
// Weights are calulated by summing the tiles fill_weight, the total distance traveled
// and the current distance from the closeset target
//
// The reverse parameter determines if the final path is returned in reverse. This uses the
// ReversePostProcessing struct and can be useful if you want to for example find a route
// back to the main path, instead of from the path to a particular place.
func NewListToPoint(reverse bool) AStarConfig {

	if reverse {
		return &listToPointReverse{}
	} else {
		return &listToPointForward{}
	}
}

func (p2l *listToPoint) SetWeight(p *PathPoint, fill_weight int, end []Point, end_map map[Point]bool) bool {
	if fill_weight == -1 {
		return false
	}

	path_length := len(end)

	min_dist := -1
	for i, end_p := range end {
		dist := p.Point.Dist(end_p) + (path_length - i)
		if min_dist == -1 || dist < min_dist {
			min_dist = dist
		}
	}

	p.Weight = p.FillWeight + p.DistTraveled + min_dist

	return true
}

func (p2l *listToPoint) IsEnd(p Point, end []Point, end_map map[Point]bool) bool {
	return end_map[p]
}

//######################################################################
// POST PROCESSORS
//######################################################################

// A post processing struct that can be embedded into a
// config and have no postprocessing applied
type VoidPostProcess struct {
}

func (v *VoidPostProcess) PostProcess(p *PathPoint, rows, cols int, filledTiles map[Point]int) *PathPoint {
	return p
}

// A post processing struct that will reverse the path thats given to it
// listToPoint for example can only generate from path to target target not
// the other way around so you can use this struct to apply reversing to the final
// path
type ReversePostProcess struct {
}

func (v *ReversePostProcess) PostProcess(p *PathPoint, rows, cols int, filledTiles map[Point]int) *PathPoint {
	var path_prev *PathPoint = nil

	for p != nil {
		next := p.Parent
		p.Parent = path_prev

		path_prev = p
		p = next
	}

	return path_prev
}
