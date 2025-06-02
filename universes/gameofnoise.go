package universes
import "oscarkilo.com/inteluni/substrates"

type GameOfNoiseUniverse struct {
  grid          *substrates.Grid2d
  noise         float64 // 0.0 to 1.0
  complexity    int     // 0 to 100
  rand          *substrates.SplitMix64
}

func NewGameOfNoiseUniverse(
    W, H int,
    noise float64,
    complexity int,
    rng *substrates.SplitMix64,
) Universe {
  if noise < 0.0 || noise > 1.0 {
    panic("noise must be between 0.0 and 1.0")
  }
  if complexity < 0 || complexity > 100 {
    panic("complexity must be between 0 and 100")
  }
  g := substrates.NewGrid2d(W, H)
  u := &GameOfNoiseUniverse{
    grid:       g,
    noise:      noise,
    complexity: complexity,
    rand:       rng,
  }
  u.seedInitialState()
  return u
}

func (u *GameOfNoiseUniverse) seedInitialState() {
  fillProb := float64(u.complexity) / 100.0
  initialFunc := func(x, y, _ int) int {
    if u.rand.Float64() < fillProb {
      return 1 // Live cell or obstacle
    }
    return 0 // Empty or dead
  }
  u.grid.Map(initialFunc)
}

func (u *GameOfNoiseUniverse) Grid() *substrates.Grid2d {
  return u.grid
}

func (u *GameOfNoiseUniverse) Advance() {
  // conway deterministic rules first
  conway := &ConwayUniverse{grid: u.grid}
  conway.Advance()
  u.grid = conway.grid
  // noise second
  noisy := &NoisyUniverse{
    grid:       u.grid,
    noise:      u.noise,
    complexity: u.complexity,
    rand:       u.rand,
  }
  noisy.Advance()
  u.grid = noisy.grid
}

func (u *GameOfNoiseUniverse) MakeEvolver() substrates.Evolver {
  return func(
    src *substrates.Grid2d,
    rng *substrates.SplitMix64,
  ) *substrates.Grid2d {
    clone := src.Clone()
    conway := &ConwayUniverse{grid: clone}
    conway.Advance()
    noisy := &NoisyUniverse{
      grid:       conway.grid,
      noise:      u.noise,
      complexity: u.complexity,
      rand:       rng,
    }
    noisy.Advance()
    return noisy.grid
  }
}

func (u *GameOfNoiseUniverse) Deterministic() bool {
  return u.noise == 0.0
}
