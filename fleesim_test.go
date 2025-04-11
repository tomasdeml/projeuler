package projeuler213_test

import (
	"context"
	"fmt"
	"projeuler213"
	"testing"

	"github.com/stretchr/testify/require"
)

const initCount = 1

func TestSimulator(t *testing.T) {
	t.Run("The number of empty positions should be in the expected range between [300;399]", func(t *testing.T) {
		g := projeuler213.NewGrid(30, initCount)
		sim := projeuler213.NewSimulator(g, 100_000, 8)

		res, err := sim.Execute(context.Background())
		require.NoError(t, err)
		require.GreaterOrEqual(t, res.EmptyPositions, float64(300))
		require.LessOrEqual(t, res.EmptyPositions, float64(399))

		t.Logf("%.6f", res.EmptyPositions)
	})
}

func TestGrid(t *testing.T) {
	const gridSize = 10

	t.Run("A new grid has all counts set to 1", func(t *testing.T) {
		g := projeuler213.NewGrid(gridSize, initCount)
		for x := range gridSize {
			for y := range gridSize {
				val, err := g.CountAt(x, y)
				require.NoError(t, err)
				require.Equal(t, projeuler213.FleeCount(1), val)
			}
		}
	})

	t.Run("A new empty grid has all counts set to 0", func(t *testing.T) {
		g := projeuler213.NewEmptyGrid(gridSize)
		for x := range gridSize {
			for y := range gridSize {
				val, err := g.CountAt(x, y)
				require.NoError(t, err)
				require.Equal(t, projeuler213.FleeCount(0), val)
			}
		}
	})

	t.Run("Squares enumerator returns [x;y] positions for each square", func(t *testing.T) {
		g := projeuler213.NewGrid(gridSize, initCount)

		squares := make(map[string]bool)
		for x, y := range g.Squares() {
			require.GreaterOrEqual(t, x, 0)
			require.Less(t, x, gridSize)

			require.GreaterOrEqual(t, y, 0)
			require.Less(t, y, gridSize)

			pos := fmt.Sprintf("%d;%d", x, y)
			require.NotContains(t, squares, pos)
			squares[pos] = true
		}

		require.Len(t, squares, gridSize*gridSize)
	})
}
