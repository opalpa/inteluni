package universes
import "oscarkilo.com/inteluni/substrates"

type ConwayUniverse struct {
  grid *substrates.Grid2d
}

func NewConwayUniverse(
  W, H int,
  initialComplexity int,  // suggestion: 10 to 30
  rng *substrates.SplitMix64,
) Universe {
  if initialComplexity < 0 || initialComplexity > 100 {
    panic("initialComplexity must be between 0 and 100")
  }
  g := substrates.NewGrid2d(W, H)
  u := &ConwayUniverse{
    grid: g,
  }
  u.seedInitialState(initialComplexity, rng)
  return u
}

func (u *ConwayUniverse) seedInitialState(
    complexityPercent int, rng *substrates.SplitMix64) {
  fillProb := float64(complexityPercent) / 100.0
  initialStateFunc := func(x, y, _ int) int {
    if rng.Float64() < fillProb {
      return 1 // Live cell
    }
    return 0 // Dead cell
  }
  u.grid.Map(initialStateFunc)
}

func (u *ConwayUniverse) Grid() *substrates.Grid2d {
  return u.grid
}

func (u *ConwayUniverse) Advance() {
  nextGrid := u.grid.Clone()
  width, height := u.grid.W(), u.grid.H()
  for x := 0; x < width; x++ {
    for y := 0; y < height; y++ {
      liveNeighbors := u.countLiveNeighbors(x, y)
      currentState := u.grid.XY(x, y)
      if currentState == 1 && (liveNeighbors < 2 || liveNeighbors > 3) {
        nextGrid.SetXY(x, y, 0) // Cell dies
      } else if currentState == 0 && liveNeighbors == 3 {
        nextGrid.SetXY(x, y, 1) // Cell becomes alive
      } else {
        nextGrid.SetXY(x, y, currentState) // State remains the same
      }
    }
  }
  u.grid = nextGrid
}

/*
func (u *ConwayUniverse) countLiveNeighbors(x, y int) int {
  width, height := u.grid.W(), u.grid.H()
  liveCount := 0
  for dx := -1; dx <= 1; dx++ {
    for dy := -1; dy <= 1; dy++ {
      if dx == 0 && dy == 0 {
        continue // Skip the cell itself
      }
      nx := (x + dx + width) % width
      ny := (y + dy + height) % height
      liveCount += u.grid.XY(nx, ny)
    }
  }
  return liveCount
}
*/
func (u *ConwayUniverse) countLiveNeighbors(x, y int) int {
  w, h := u.grid.W(), u.grid.H()
  xm1 := (x - 1 + w) % w
  xp1 := (x + 1) % w
  ym1 := (y - 1 + h) % h
  yp1 := (y + 1) % h
  return u.grid.XY(xm1, ym1) +
         u.grid.XY(x,   ym1) +
         u.grid.XY(xp1, ym1) +
         u.grid.XY(xm1, y)   +
         u.grid.XY(xp1, y)   +
         u.grid.XY(xm1, yp1) +
         u.grid.XY(x,   yp1) +
         u.grid.XY(xp1, yp1)
}

func (u *ConwayUniverse) MakeEvolver() substrates.Evolver {
  return func(
      src *substrates.Grid2d,
      _ *substrates.SplitMix64,
  ) *substrates.Grid2d {
    tempU := &ConwayUniverse{grid: src}
    tempU.Advance()
    return tempU.grid
  }
}

func (u *ConwayUniverse) Deterministic() bool {
  return true
}
