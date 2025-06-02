package universes
import "testing"
import "oscarkilo.com/inteluni/substrates"

func TestNewConwayUniverse_GridDimensions(t *testing.T) {
  rng := substrates.NewSplitMix64(42)
  W := 10
  H := 20
  u := NewConwayUniverse(W, H, 10, rng)
  grid := u.Grid()
  if grid.W() != W {
    t.Errorf("grid width: expected %d, got %d", W, grid.W())
  }
  if grid.H() != H {
    t.Errorf("grid height: expected %d, got %d", H, grid.H())
  }
}

func TestSeedInitialState_PopulationRate(t *testing.T) {
  rng := substrates.NewSplitMix64(123)
  W := 20
  H := 20
  complexity := 25
  u := NewConwayUniverse(W, H, complexity, rng)
  total := W * H
  live := 0
  for x := 0; x < W; x++ {
    for y := 0; y < H; y++ {
      live += u.Grid().XY(x, y)
    }
  }
  ratio := float64(live) / float64(total)
  expected := float64(complexity) / 100.0
  if ratio < expected-0.1 || ratio > expected+0.1 {
    t.Errorf("live cell ratio: expected around %.2f, got %.2f", expected, ratio)
  }
}

func TestAdvance_LonelyDies(t *testing.T) {
  g := substrates.NewGrid2d(3, 3)
  g.SetXY(1, 1, 1)
  u := &ConwayUniverse{grid: g}
  u.Advance()
  if u.Grid().XY(1, 1) != 0 {
    t.Errorf("lonely cell should die")
  }
}

func TestAdvance_BirthRule(t *testing.T) {
  g := substrates.NewGrid2d(3, 3)
  g.SetXY(0, 0, 1)
  g.SetXY(0, 1, 1)
  g.SetXY(1, 0, 1)
  u := &ConwayUniverse{grid: g}
  u.Advance()
  if u.Grid().XY(1, 1) != 1 {
    t.Errorf("dead cell with 3 neighbors should become alive")
  }
}

func TestAdvance_Survival(t *testing.T) {
  g := substrates.NewGrid2d(3, 3)
  g.SetXY(0, 0, 1)
  g.SetXY(0, 1, 1)
  g.SetXY(1, 0, 1)
  g.SetXY(1, 1, 1)
  u := &ConwayUniverse{grid: g}
  u.Advance()
  if u.Grid().XY(0, 0) != 1 {
    t.Errorf("live cell with 3 neighbors should survive")
  }
}

func TestMakeEvolver_ProducesNextState(t *testing.T) {
  g := substrates.NewGrid2d(3, 3)
  g.SetXY(0, 0, 1)
  g.SetXY(0, 1, 1)
  g.SetXY(1, 0, 1)
  u := &ConwayUniverse{grid: g}
  evolver := u.MakeEvolver()
  next := evolver(g, nil)
  if next.XY(1, 1) != 1 {
    t.Errorf("evolver should produce correct next state")
  }
}

func TestDeterministic_ReturnsTrue(t *testing.T) {
  rng := substrates.NewSplitMix64(1)
  u := NewConwayUniverse(5, 5, 20, rng)
  if !u.Deterministic() {
    t.Errorf("Deterministic should return true")
  }
}

func TestMakeEvolver_InputGridUnchanged(t *testing.T) {
  g := substrates.NewGrid2d(3, 3)
  g.SetXY(0, 0, 1)
  g.SetXY(0, 1, 1)
  g.SetXY(1, 0, 1)
  snapshot := g.Clone()
  u := &ConwayUniverse{grid: g}
  evolver := u.MakeEvolver()
  result := evolver(g, nil)

  for x := 0; x < 3; x++ {
    for y := 0; y < 3; y++ {
      got := g.XY(x, y)
      want := snapshot.XY(x, y)
      if got != want {
        t.Errorf("grid changed at (%d,%d): expected %d, got %d", x, y, want, got)
      }
    }
  }

  if result == g {
    t.Errorf("evolver should return a new grid, but got same reference")
  }
}

func makeTestUniverse(w, h int) *ConwayUniverse {
  g := substrates.NewGrid2d(w, h)
  return &ConwayUniverse{grid: g}
}

func TestCountLiveNeighbors_Empty(t *testing.T) {
  u := makeTestUniverse(3, 3)
  count := u.countLiveNeighbors(1, 1)
  if count != 0 {
    t.Errorf("expected 0 live neighbors, got %d", count)
  }
}

func TestCountLiveNeighbors_Full(t *testing.T) {
  u := makeTestUniverse(3, 3)
  for y := 0; y < 3; y++ {
    for x := 0; x < 3; x++ {
      u.grid.SetXY(x, y, 1)
    }
  }
  count := u.countLiveNeighbors(1, 1)
  if count != 8 {
    t.Errorf("expected 8 live neighbors, got %d", count)
  }
}

func TestCountLiveNeighbors_CenterCell(t *testing.T) {
  u := makeTestUniverse(3, 3)
  // Set only the diagonals around center to alive
  u.grid.SetXY(0, 0, 1)
  u.grid.SetXY(2, 0, 1)
  u.grid.SetXY(0, 2, 1)
  u.grid.SetXY(2, 2, 1)

  count := u.countLiveNeighbors(1, 1)
  if count != 4 {
    t.Errorf("expected 4 live neighbors, got %d", count)
  }
}

func TestCountLiveNeighbors_EdgeWrap(t *testing.T) {
  u := makeTestUniverse(3, 3)
  u.grid.SetXY(2, 2, 1) // bottom-right corner
  count := u.countLiveNeighbors(0, 0)
  if count != 1 {
    t.Errorf("expected 1 live neighbor with wraparound, got %d", count)
  }
}

func TestCountLiveNeighbors_WrapFull(t *testing.T) {
  u := makeTestUniverse(3, 3)
  u.grid.SetXY(2, 2, 1)
  u.grid.SetXY(2, 0, 1)
  u.grid.SetXY(2, 1, 1)

  count := u.countLiveNeighbors(0, 0) // should wrap from right and bottom
  if count != 3 {
    t.Errorf("expected 3 live neighbors with wraparound, got %d", count)
  }
}

func TestCountLiveNeighbors_SelfIsIgnored(t *testing.T) {
  u := makeTestUniverse(3, 3)
  u.grid.SetXY(1, 1, 1) // center cell is alive
  count := u.countLiveNeighbors(1, 1)
  if count != 0 {
    t.Errorf("expected 0 live neighbors (self ignored), got %d", count)
  }
}

func TestCountLiveNeighbors_LargePattern(t *testing.T) {
  u := makeTestUniverse(5, 5)

  // Layout (1 = alive, 0 = dead):
  // 0 1 0 0 1
  // 1 1 1 0 0
  // 0 1 0 1 0
  // 0 0 0 0 0
  // 1 0 1 0 1

  // Top row
  u.grid.SetXY(1, 0, 1)
  u.grid.SetXY(4, 0, 1)

  // Row 1
  u.grid.SetXY(0, 1, 1)
  u.grid.SetXY(1, 1, 1)
  u.grid.SetXY(2, 1, 1)

  // Row 2
  u.grid.SetXY(1, 2, 1)
  u.grid.SetXY(3, 2, 1)

  // Row 4 (bottom)
  u.grid.SetXY(0, 4, 1)
  u.grid.SetXY(2, 4, 1)
  u.grid.SetXY(4, 4, 1)

  cases := []struct {
    x, y   int
    expect int
  }{
    {1, 0, 5},
    {4, 0, 3},
    {0, 1, 4},
    {1, 1, 4},
    {2, 1, 4},
    {1, 2, 3},
    {3, 2, 1},
    {0, 4, 3},
    {2, 4, 1},
    {4, 4, 2},
  }

  for _, c := range cases {
    got := u.countLiveNeighbors(c.x, c.y)
    if got != c.expect {
      t.Errorf("countLiveNeighbors(%d,%d): expected %d, got %d", c.x, c.y, c.expect, got)
    }
  }
}
