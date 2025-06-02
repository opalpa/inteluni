package metrics
import "testing"
import "oscarkilo.com/inteluni/substrates"
import "oscarkilo.com/inteluni/universes"

func TestTauLMonotoneNoise(t *testing.T) {
  rng := substrates.NewSplitMix64(1)
  W, H := 16, 16
  noise := []float64{0.1, 0.4, 0.7}
  tau := make([]float64, len(noise))
  for i, n := range noise {
    u := universes.NewNoisyUniverse(W, H, n, 20, rng.NewFromSelf())
    tau[i] = TauL(u, rng.NewFromSelf())
  }
  if !(tau[0] > tau[1] && tau[1] > tau[2]) {
    t.Fatalf("τ_L not monotone w.r.t noise: %v", tau)
  }
}
