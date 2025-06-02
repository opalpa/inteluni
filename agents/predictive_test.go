package agents
import "math"
import "testing"
import "oscarkilo.com/inteluni/substrates"

func TestWorstCaseReducer(t *testing.T) {
  scoreByMove := map[substrates.Move][]float64{
    substrates.North: {0.2, 0.1},
    substrates.South: {0.6, 0.5},
    substrates.East:  {0.8, 0.4},
  }
  m, s := worstCaseReducer(scoreByMove)
  if m != substrates.South {
    t.Fatalf("expected South, got %v", m)
  }
  if math.Abs(s-0.5) > 1e-9 {
    t.Fatalf("expected score 0.5, got %v", s)
  }
}

func TestMeanCaseReducer(t *testing.T) {
  scoreByMove := map[substrates.Move][]float64{
    substrates.North: {0.2, 0.1},
    substrates.South: {0.9, 0.2},
    substrates.East:  {0.5, 0.4},
  }
  m, s := meanCaseReducer(scoreByMove)
  if m != substrates.South {
    t.Fatalf("expected South, got %v", m)
  }
  if math.Abs(s-0.55) > 1e-9 {
    t.Fatalf("expected score 0.55, got %v", s)
  }
}

func TestToroidal(t *testing.T) {
  pos := substrates.Pos{X: 0, Y: 0}
  got := toroidal(pos, substrates.West, 5, 5)
  want := substrates.Pos{X: 4, Y: 0}
  if got != want {
    t.Fatalf("expected %v, got %v", want, got)
  }
}

// TestStateKey verifies stability and sensitivity of the key builder.
func TestStateKey(t *testing.T) {
  g := substrates.NewGrid2d(2, 2)
  k1 := stateKey(g, substrates.Pos{X: 0, Y: 0}, 1)
  k2 := stateKey(g, substrates.Pos{X: 0, Y: 0}, 1)
  if k1 != k2 {
    t.Fatalf("identical states yielded different keys")
  }
  k3 := stateKey(g, substrates.Pos{X: 1, Y: 0}, 1)
  if k1 == k3 {
    t.Fatalf("different states yielded identical keys")
  }
}

func TestPredictiveDecide(t *testing.T) {
  g := substrates.NewGrid2d(3, 3)
  // place an obstacle directly north of the agent at (1,0)
  g.Map(func(x, y, _ int) int {
    if x == 1 && y == 0 {
      return 1
    }
    return 0
  })
  rng := substrates.NewSplitMix64(42)
  ag := &PredictiveAgent{
    baseAgent: baseAgent{
      id:        1,
      pos:       substrates.Pos{X: 1, Y: 1},
      foresight: 1,
      rng:       rng,
    },
  }
  evolver := func(src *substrates.Grid2d, rng *substrates.SplitMix64) *substrates.Grid2d {
    return src
  }
  move := ag.predictiveDecide(g, evolver, true)
  if move == substrates.North {
    t.Fatalf("agent chose to move into an obstacle")
  }
}

func asciiToGrid(spec []string) *substrates.Grid2d {
  h := len(spec)
  w := len(spec[0])
  g := substrates.NewGrid2d(w, h)
  g.Map(func(x, y, _ int) int {
    if spec[y][x] == '#' {
      return 1
    }
    return 0
  })
  return g
}

func alwaysSameEvolver(grid *substrates.Grid2d) substrates.Evolver {
  return func(src *substrates.Grid2d, _ *substrates.SplitMix64) *substrates.Grid2d {
    return grid
  }
}

func TestPredictiveSequenceEast(t *testing.T) {
  specs := [][]string{
    {
     "###",
     "##_",
     "###",
    },
  }
  evolver := alwaysSameEvolver(asciiToGrid(specs[0]))
  ag := &PredictiveAgent{
    baseAgent: baseAgent{
      id:        2,
      pos:       substrates.Pos{X: 1, Y: 1},
      foresight: 2,
      rng:       substrates.NewSplitMix64(0),
    },
  }
  originalSpecs := [][]string{
    {
      "___",
      "___",
      "___",
    },
  }
  original := asciiToGrid(originalSpecs[0])
  move := ag.predictiveDecide(original, evolver, true)
  if move != substrates.East {
    t.Fatalf("expected East, got %v", move)
  }
}

func TestPredictiveSequenceNorth(t *testing.T) {
  specs := [][]string{
    {
     "#_#",
     "###",
     "###",
    },
  }
  evolver := alwaysSameEvolver(asciiToGrid(specs[0]))
  ag := &PredictiveAgent{
    baseAgent: baseAgent{
      id:        2,
      pos:       substrates.Pos{X: 1, Y: 1},
      foresight: 2,
      rng:       substrates.NewSplitMix64(0),
    },
  }
  originalSpecs := [][]string{
    {
      "___",
      "___",
      "___",
    },
  }
  original := asciiToGrid(originalSpecs[0])
  move := ag.predictiveDecide(original, evolver, true)
  if move != substrates.North {
    t.Fatalf("expected North, got %v", move)
  }
}

func randomEvolver(grids []*substrates.Grid2d) substrates.Evolver {
  return func(src *substrates.Grid2d, rng *substrates.SplitMix64) *substrates.Grid2d {
    g := grids[rng.Intn(len(grids))]
    return g
  }
}

func TestPredictiveSequenceRandom(t *testing.T) {
  specs := [][]string{
    {
      "_##",
      "##_",
      "##_",
    },
    {
      "###",
      "#_#",
      "##_",
    },
  }
  var grids []*substrates.Grid2d
  for _, spec := range specs {
    grids = append(grids, asciiToGrid(spec))
  }
  evolver := randomEvolver(grids)
  ag := &PredictiveAgent{
    baseAgent: baseAgent{
      id:        3,
      pos:       substrates.Pos{X: 2, Y: 1},
      foresight: 3,
      rng:       substrates.NewSplitMix64(1),
    },
  }
  move := ag.predictiveDecide(grids[0], evolver, false)
  if move != substrates.South {
    t.Fatalf("expected South, got %v", move)
  }
}

func TestRolloutBudgetDeterministic(t *testing.T) {
  for _, depth := range []int{0, 1, 5, 10, 100} {
    for _, horizon := range []int{1, 5, 10, 100} {
      if depth > horizon {
        continue // skip cases where depth > horizon
      }
      got := rolloutBudget(depth, horizon, true)
      if got != 1 {
        t.Fatalf("deterministic case: expected 1, got %d", got)
      }
    }
  }
}

func TestRolloutBudgetDepth(t *testing.T) {
  horizon := 10
  last := 0
  for depth := 0; depth <= horizon; depth++ {
    got := rolloutBudgetMM(depth, horizon, false, 1, 8)
    if got < last {
      t.Fatalf("non-monotonic rollout: depth %d gave %d < previous %d",
        depth, got, last)
    }
    if got < 1 || got > 8 {
      t.Fatalf("invalid rollout count at depth %d: %d", depth, got)
    }
    last = got
  }
}

func TestRolloutBudgetSpecific(t *testing.T) {
  table := []struct{
    depth        int
    horizon      int
    deterministic bool
    expected     int
  }{
    {0, 10, false, 1},
    {5, 10, false, 5},
    {10, 10, false, 8},
    {7, 10, false, 6},
    {2, 10, false, 2},
    {10, 10, true, 1},
    {0, 1, false, 1},  // edge case
    {1, 1, false, 8},  // edge case
  }

  for _, row := range table {
    got := rolloutBudgetMM(row.depth, row.horizon, row.deterministic, 1, 8)
    if got != row.expected {
      t.Fatalf("depth=%d horizon=%d det=%v: expected %d, got %d",
        row.depth, row.horizon, row.deterministic, row.expected, got)
    }
  }

  defer func() {
    if r := recover(); r == nil {
      t.Fatalf("expected panic on invalid horizon=0, got none")
    }
  }()
  _ = rolloutBudgetMM(1, 0, false, 1, 8)
}
