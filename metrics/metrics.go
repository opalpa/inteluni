package metrics
import "bytes"
import "compress/gzip"  // DEFLATE RFC 1951
import "encoding/binary"
import "math"
import "oscarkilo.com/inteluni/substrates"
import "oscarkilo.com/inteluni/universes"
import "sort"

// ---------- Kolmogorov‑proxy K ----------
// K = (compressed / raw) on the entire episode buffer.
func KolmogorovProxy(grids []*substrates.Grid2d) float64 {
  if len(grids) == 0 {
    panic("KolmogorovProxy: no grids provided")
  }
  var raw bytes.Buffer
  var err error
  err = binary.Write(&raw, binary.LittleEndian, int32(grids[0].W()))
  if err != nil {
    panic("KolmogorovProxy: writing grid width failed")
  }
  err = binary.Write(&raw, binary.LittleEndian, int32(grids[0].H()))
  if err != nil {
    panic("KolmogorovProxy: writing grid height failed")
  }
  for _, g := range grids {
    for y := 0; y < g.H(); y++ {
      for x := 0; x < g.W(); x++ {
        if err := raw.WriteByte(byte(g.XY(x, y))); err != nil {
          panic("KolmogorovProxy: writing grid failed")
        }
      }
    }
  }
  var cmp bytes.Buffer
  zw := gzip.NewWriter(&cmp)
  _, err = zw.Write(raw.Bytes())
  if err != nil {
    panic("KolmogorovProxy: compression failed")
  }
  err = zw.Close()
  if err != nil {
    panic("KolmogorovProxy: compression failed")
  }
  return float64(cmp.Len())/float64(raw.Len())
}

// ---------- Lyapunov horizon τ_L ----------
func TauL(
  universe universes.Universe,
  rng *substrates.SplitMix64,
) float64 {
  const tauruns = 11
  manytau := make([]float64, tauruns)
  for i := 0; i < tauruns; i++ {
    tau := tauLOne(universe, rng)
    manytau[i] = tau
  }
  // return the median
  sort.Slice(manytau, func(i, j int) bool {
    return manytau[i] < manytau[j]
  })
  medianLocation := len(manytau) / 2
  median := manytau[medianLocation]
  return median
}

// Perturb one bit, measure divergence d(t), fit log(d) = λt + const.
// If fitted Lyapunov exponent λ ≤ 0 we assign the horizon τ_L = steps.
// It retains ordering: all non‑divergent runs share
// the same “very long but finite” τ L.
// Log‑scales and heat‑maps behave (no masked cells).
func tauLOne(
  universe universes.Universe,
  rng *substrates.SplitMix64,
) float64 {
  base := universe.Grid().Clone()
  pert := base.Clone()
  // flip one random cell
  x := rng.Intn(base.W())
  y := rng.Intn(base.H())
  val := base.XY(x, y)
  pert.SetXY(x, y, 1-val)

  steps := 50
  evolver := universe.MakeEvolver()
  // rng1, rng2 := rng.Clone(), rng.Clone()  // same noise pattern
  rng1, rng2 := rng.NewFromSelf(), rng.NewFromSelf()
  d := make([]float64, steps)
  for t := 0; t < steps; t++ {
    d[t] = hamming(base, pert)
    if t == 0 && d[t] != 1 {
      panic("TauL: initial perturbation should be 1")
    }
    base = evolver(base, rng1)
    pert = evolver(pert, rng2)
  }

  cutoff := 0
  maxDiff := float64(base.W() * base.H())
  for cutoff < steps && d[cutoff] < 0.05*maxDiff {
    cutoff++
  }
  if cutoff < 3 {
    cutoff = 3
  }
  xs := make([]float64, cutoff)
  ys := make([]float64, cutoff)
  for i := 0; i < cutoff; i++ {
    xs[i], ys[i] = float64(i), math.Log(d[i]+1e-9)
  }
  λ := olsSlope(xs, ys)
  if λ <= 0 {
    // return math.Inf(1)
    return float64(steps) // facilitate log scale
  }
  return 1.0 / λ
}

func hamming(a, b *substrates.Grid2d) float64 {
  diff := 0
  for y := 0; y < a.H(); y++ {
    for x := 0; x < a.W(); x++ {
      if a.XY(x, y) != b.XY(x, y) {
        diff++
      }
    }
  }
  return float64(diff)
}

// olsSlope returns the ordinary‑least‑squares slope of y on x.
// If xs and ys are empty or var(xs)==0, it returns 0.
//   slope = Cov(xs,ys) / Var(xs)
func olsSlope(xs, ys []float64) float64 {
  n := len(xs)
  if n == 0 || n != len(ys) {
    return 0.0
  }
  var sumX, sumY, sumXX, sumXY float64
  for i := 0; i < n; i++ {
    x, y := xs[i], ys[i]
    if math.IsNaN(x) || math.IsNaN(y) {
      panic("olsSlope: invalid input NaN")
    }
    if math.IsInf(x, 0) || math.IsInf(y, 0) {
      panic("olsSlope: invalid input Inf")
    }
    sumX += x
    sumY += y
    sumXX += x * x
    sumXY += x * y
  }
  // variance and covariance (dividing by n later cancels)
  cov := sumXY - (sumX*sumY)/float64(n)
  varX := sumXX - (sumX*sumX)/float64(n)
  if varX == 0 {
    return 0.0
  }
  return cov / varX
}
