// The projeuler213 package implements the following simulation:
// A 30x30 grid of squares contains 900 fleas, initially one flea per square.
// When a bell is rung, each flea jumps to an adjacent square at random (usually 4 possibilities, except for fleas on the edge of the grid or at the corners).
// What is the expected number of unoccupied squares after 50 rings of the bell? Give your answer rounded to six decimal places.
package projeuler213

import (
	"context"
	"fmt"
	"iter"
	"math/rand"
	"sync"
)

type FleeCount int

type FleeGrid struct {
	size int
	grid [][]FleeCount
}

func NewGrid(size int, initCount FleeCount) *FleeGrid {
	grid := make([][]FleeCount, size)
	for x := range size {
		grid[x] = make([]FleeCount, size)
		for y := range size {
			grid[x][y] = initCount
		}
	}
	return &FleeGrid{size: size, grid: grid}
}

func NewEmptyGrid(size int) *FleeGrid {
	grid := make([][]FleeCount, size)
	for x := range size {
		grid[x] = make([]FleeCount, size)
	}
	return &FleeGrid{size: size, grid: grid}
}

func (g *FleeGrid) IncCountAt(x, y int) error {
	if !g.validPos(x, y) {
		return fmt.Errorf("position [%d;%d] is out of bounds", x, y)
	}
	g.grid[x][y]++
	return nil
}

func (g *FleeGrid) CountAt(x, y int) (FleeCount, error) {
	if !g.validPos(x, y) {
		return 0, fmt.Errorf("position [%d;%d] is out of bounds", x, y)
	}
	cnt := g.grid[x][y]
	if cnt < 0 {
		return 0, fmt.Errorf("the count at [%d;%d] has been corrupted: %d is < 0", x, y, cnt)
	}

	return cnt, nil
}

func (g *FleeGrid) Squares() iter.Seq2[int, int] {
	return func(yield func(int, int) bool) {
		for x := range g.size {
			for y := range g.size {
				if !yield(x, y) {
					return
				}
			}
		}
	}
}

func (g *FleeGrid) validPos(x, y int) bool {
	return (x >= 0 && x < g.size) && (y >= 0 && y < g.size)
}

type Simulator struct {
	grid    *FleeGrid
	runs    int
	workers int
}

type Result struct {
	EmptyPositions float64
}

type workerResult struct {
	Result Result
	Err    error
}

func NewSimulator(grid *FleeGrid, runs int, workers int) *Simulator {
	return &Simulator{
		grid:    grid,
		runs:    runs,
		workers: workers,
	}
}

func (s *Simulator) Execute(ctx context.Context) (*Result, error) {
	results := make(chan workerResult, s.workers)
	runs := s.runs / s.workers

	var wg sync.WaitGroup
	for range s.workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.execWorker(ctx, runs, results)
		}()
	}

	wg.Wait()
	close(results)

	var emptyPosSum float64
	for range s.workers {
		res := <-results
		if res.Err != nil {
			return nil, res.Err
		}
		emptyPosSum += res.Result.EmptyPositions
	}

	return &Result{EmptyPositions: emptyPosSum / float64(s.workers)}, nil
}

func (s *Simulator) execWorker(ctx context.Context, runs int, results chan<- workerResult) {
	grid := s.grid
	for range runs {
		if ctx.Err() != nil {
			return
		}
		nextGrid := NewEmptyGrid(s.grid.size)
		for x, y := range grid.Squares() {
			flees, _ := grid.CountAt(x, y)
			if flees == 0 {
				continue
			}
			for range flees {
				dstX, dstY := x, y
				if flipCoin() {
					dstX = jumpFrom(x, grid.size)
				} else {
					dstY = jumpFrom(y, grid.size)
				}
				if err := nextGrid.IncCountAt(dstX, dstY); err != nil {
					results <- workerResult{Err: fmt.Errorf("incrementing a count at position [%d;%d]: %w", dstX, dstY, err)}
					return
				}
			}
		}
		grid = nextGrid
	}

	res := workerResult{Result: Result{}}
	for x, y := range grid.Squares() {
		flees, err := grid.CountAt(x, y)
		if err != nil {
			results <- workerResult{Err: fmt.Errorf("getting a count at position [%d;%d]: %w", x, y, err)}
			return
		}
		if flees != 0 {
			continue
		}
		res.Result.EmptyPositions++
	}

	results <- res
}

func jumpFrom(pos int, gridSize int) int {
	switch pos {
	case 0:
		return pos + 1
	case gridSize - 1:
		return pos - 1
	default:
		if flipCoin() {
			return pos + 1
		}
		return pos - 1
	}
}

func flipCoin() bool {
	return rand.Intn(2) != 0
}
