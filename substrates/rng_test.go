package substrates
import "testing"

func TestSplitMix64Clone(t *testing.T) {
  rng := NewSplitMix64(12345)
  clone := rng.Clone()
  const nTests = 10
  for i := 0; i < nTests; i++ {
    r1 := rng.Intn(100)
    r2 := clone.Intn(100)
    if r1 != r2 {
      t.Errorf("mismatch at %d: got %d, want %d", i, r2, r1)
    }
  }

  rng.Intn(100)
  r1 := rng.Intn(100)
  r2 := clone.Intn(100)
  if r1 == r2 {
    t.Errorf("clone not independent: both produced %d", r1)
  }

  rng2 := NewSplitMix64(12345)
  for i := 0; i < nTests + 1; i++ {
    _ = rng2.Intn(100)  // burn to catch up to clone
  }

  for i := 0; i < nTests; i++ {
    r1 := rng2.Intn(100)
    r2 := clone.Intn(100)
    if r1 != r2 {
      t.Errorf("clone diverged at %d: got %d, want %d", i, r2, r1)
    }
  }
}
