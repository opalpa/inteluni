package substrates

// SplitMix64: fast, simple, cloneable RNG
type SplitMix64 struct {
  state uint64
}

func NewSplitMix64(seed uint64) *SplitMix64 {
  return &SplitMix64{state: seed}
}

// Returns next uint64
func (r *SplitMix64) NextUint64() uint64 {
  r.state += 0x9E3779B97F4A7C15
  z := r.state
  z = (z ^ (z >> 30)) * 0xBF58476D1CE4E5B9
  z = (z ^ (z >> 27)) * 0x94D049BB133111EB
  return z ^ (z >> 31)
}

// Returns float64 in [0.0, 1.0)
func (r *SplitMix64) Float64() float64 {
  return float64(r.NextUint64()>>11) / (1 << 53)
}

// Intn returns a random int in [0, n). It panics if n <= 0.
func (r *SplitMix64) Intn(n int) int {
  if n <= 0 {
    panic("Intn: n must be positive")
  }
  // Convert to uint64 to handle large int values safely
  max := uint64(n)
  // Calculate the threshold to avoid modulo bias
  threshold := ^uint64(0) - (^uint64(0) % max)
  // Rejection sampling
  for {
    v := r.NextUint64()
    if v < threshold {
      return int(v % max)
    }
  }
}

func (r *SplitMix64) NewFromSelf() *SplitMix64 {
  return NewSplitMix64(r.NextUint64())
}

func (r *SplitMix64) Clone() *SplitMix64 {
  return &SplitMix64{state: r.state}
}
