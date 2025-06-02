package universes
import "oscarkilo.com/inteluni/substrates"

type NoisyUniverse struct {
  grid       *substrates.Grid2d
  noise      float64 // 0.0 to 1.0
  complexity int     // 0 to 100
  rand       *substrates.SplitMix64
}

func NewNoisyUniverse(
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

  var g *substrates.Grid2d = substrates.NewGrid2d(W, H)
  u := &NoisyUniverse{
    grid:       g,
    noise:      noise,
    complexity: complexity,
    rand:       rng,
  }
  u.seedObstacles()
  return u
}

func (u *NoisyUniverse) seedObstacles() {
  fillProb := float64(u.complexity) / 100.0
  obstacleFunc := func(x, y, _ int) int {
    if u.rand.Float64() < fillProb {
      return 1 // Obstacle
    }
    return 0 // Empty
  }
  u.grid.Map(obstacleFunc)
}

func (u *NoisyUniverse) Grid() *substrates.Grid2d {
  return u.grid
}

func (u *NoisyUniverse) Advance() {
  u.injectNoise()
}

func (u *NoisyUniverse) injectNoise() {
  fillProb := float64(u.complexity) / 100.0
  noiseFunc := func(x, y, val int) int {
    if u.rand.Float64() < u.noise {
      if u.rand.Float64() < fillProb {
        return 1
      }
      return 0
    }
    return val // retain original value
  }
  u.grid.Map(noiseFunc)
}

func (u *NoisyUniverse) MakeEvolver() substrates.Evolver {
  return func(
      src *substrates.Grid2d,
      rng *substrates.SplitMix64,
  ) *substrates.Grid2d {
    clone := src.Clone()
    tempU := &NoisyUniverse{
      grid:       clone,
      noise:      u.noise,
      complexity: u.complexity,
      rand:       rng,
    }
    tempU.injectNoise()
    return clone
  }
}

func (u *NoisyUniverse) Deterministic() bool {
  return u.noise == 0.0
}
