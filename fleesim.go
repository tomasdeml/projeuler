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

	"golang.org/x/sync/errgroup"
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

func NewSimulator(grid *FleeGrid, runs int, workers int) *Simulator {
	return &Simulator{
		grid:    grid,
		runs:    runs,
		workers: workers,
	}
}

func (s *Simulator) Execute(ctx context.Context) (*Result, error) {
	results := make(chan Result, s.workers)
	runs := s.runs / s.workers

	errgroup, ctx := errgroup.WithContext(ctx)
	for range s.workers {
		errgroup.Go(func() error {
			return s.execWorker(ctx, runs, results)
		})
	}

	err := errgroup.Wait()
	close(results)
	if err != nil {
		return nil, err
	}

	var emptyPosSum float64
	for res := range results {
		emptyPosSum += res.EmptyPositions
	}

	return &Result{EmptyPositions: emptyPosSum / float64(s.workers)}, nil
}

func (s *Simulator) execWorker(ctx context.Context, runs int, results chan<- Result) error {
	grid := s.grid
	for range runs {
		if ctx.Err() != nil {
			return nil
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
					return fmt.Errorf("incrementing the count at position [%d;%d]: %w", dstX, dstY, err)
				}
			}
		}
		grid = nextGrid
	}

	res := Result{}
	for x, y := range grid.Squares() {
		flees, err := grid.CountAt(x, y)
		if err != nil {
			return fmt.Errorf("getting a count at position [%d;%d]: %w", x, y, err)
		}
		if flees != 0 {
			continue
		}
		res.EmptyPositions++
	}

	results <- res
	return nil
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
