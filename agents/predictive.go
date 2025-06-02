package agents
// Predictive planning with discounted‑survival reward.
import "strings"
import "fmt"
import "oscarkilo.com/inteluni/substrates"
import "math"

const (
  aliveReward = 1.0     // per safe step
  deathPenalty = 0.0    // reward for death, straight survival time
  gamma = 0.9           // discount rate of future rewards

// At each simulated step the agent receives
//     r_t = 1            if still alive
//           0            if it collides
//
// and maximises the discounted return
//     G = Σ_{t=0}^{H-1} γ^t r_t                       (Bellman, 1957)
//
// This is the standard formulation for episodic survival tasks in
// reinforcement learning (Sutton & Barto, 2018, ch. 3).  It reduces to
// “expected time‑to‑collision” when γ = 1, and becomes risk‑averse as
// γ → 0.
//
// Multiple roll‑outs per move generate an outcome distribution; we
// collapse that distribution with a *survivalReducer*:
//   • worst‑case  - robust control (Howard & Matheson, 1972)
//   • mean‑case   - expected utility
// Possible future reducers:
//   • CVaR or other risk function  (Tamar et al., 2015).
//
// References
// ----------
// • Bellman, R. *Dynamic Programming.* Princeton Univ. Press, 1957.
// • Sutton, R. S. & Barto, A. G. *Reinforcement Learning: An
//   Introduction* (2nd ed.). MIT Press, 2018.
// • Howard, R. A. & Matheson, J. E. “Risk‑Sensitive Markov Decision
//   Processes.” *Management Science*, 1972.
// • Tamar, A. et al. “Policy Gradient for Coherent Risk Measures.”
//   *Proc. ICML 2015*.

  memoize = true
)

type survivalReducer func(
    scoreByMove map[substrates.Move][]float64,
) (substrates.Move, float64)

func worstCaseReducer(
    scoreByMove map[substrates.Move][]float64,
) (substrates.Move, float64) {
  // find the move with highest low score
  bestMove := substrates.Stay
  bestScore := -1.0
  for move, scores := range scoreByMove {
    if len(scores) == 0 {
      panic("no scores for move")
    }
    worstScore := scores[0]
    for _, score := range scores[1:] {
      if score < worstScore {
        worstScore = score
      }
    }
    if worstScore > bestScore {
      bestMove = move
      bestScore = worstScore
    }
  }
  if bestScore < 0 {
    panic("no valid moves found")
  }
  return bestMove, bestScore
}

func meanCaseReducer(
    scoreByMove map[substrates.Move][]float64,
) (substrates.Move, float64) {
  // find the move with highest mean score
  bestMove := substrates.Stay
  bestScore := -1.0
  for move, scores := range scoreByMove {
    if len(scores) == 0 {
      panic("no scores for move")
    }
    sum := 0.0
    for _, score := range scores {
      sum += score
    }
    meanScore := sum / float64(len(scores))
    if meanScore > bestScore {
      bestMove = move
      bestScore = meanScore
    }
  }
  if bestScore < 0 {
    panic("no valid moves found")
  }
  return bestMove, bestScore
}

func toroidal(
    pos substrates.Pos,
    m substrates.Move,
    W, H int,
) substrates.Pos {
  x := (pos.X + m.DX() + W) % W
  y := (pos.Y + m.DY() + H) % H
  return substrates.Pos{X: x, Y: y}
}

var possibleMoves = []substrates.Move{
    substrates.North,
    substrates.South,
    substrates.East,
    substrates.West,
    substrates.Stay,
}

type memoEntry struct {
  move  substrates.Move
  score float64
}

func (a *PredictiveAgent) predictiveDecide(
    g *substrates.Grid2d,
    evolve substrates.Evolver,
    deterministic bool,
) substrates.Move {
  memo := make(map[string]memoEntry)
  reducer := worstCaseReducer
  // reducer := meanCaseReducer
  if a.foresight <= 0 {
    panic("foresight must be positive")
  }
  depthLeft := a.foresight
  maxDepth := a.foresight
  moveToScore := make(map[substrates.Move][]float64)
  runs := rolloutBudget(depthLeft, maxDepth, deterministic)
  for i := 0; i < runs; i++ {
    possibleFuture := evolve(g, a.rng)
    for _, m := range possibleMoves {
      posInPossibleFuture := toroidal(a.pos, m, g.W(), g.H())
      score := a.evaluate(
          possibleFuture,
          posInPossibleFuture,
          depthLeft-1,
          maxDepth,
          evolve,
          deterministic,
          memo,
          reducer)
      moveToScore[m] = append(moveToScore[m], score)
    }
  }
  bestMove, _ := reducer(moveToScore)
  return bestMove
}

func (a *PredictiveAgent) evaluate(
    g *substrates.Grid2d,
    pos substrates.Pos,
    depthLeft, maxDepth int,
    evolve substrates.Evolver,
    deterministic bool,
    memo map[string]memoEntry,
    reducer survivalReducer,
) float64 {
  if depthLeft < 0 {
    panic("depthLeft must be non-negative")
  }
  if depthLeft == 0 {
    if g.Get(pos) == 1 {
      return deathPenalty
    }
    if g.Get(pos) == 0 {
      return aliveReward
    }
    panic("invalid grid value at position")
  }
  if g.Get(pos) == 1 {
    return deathPenalty
  }
  var key string
  if memoize && depthLeft > 3 {
    key = stateKey(g, pos, depthLeft)
    if entry, ok := memo[key]; ok {
      return entry.score
    }
  }
  moveToScore := make(map[substrates.Move][]float64)
  runs := rolloutBudget(depthLeft, maxDepth, deterministic)
  for i := 0; i < runs; i++ {
    possibleFuture := evolve(g, a.rng)
    for _, m := range possibleMoves {
      nextPos := toroidal(pos, m, g.W(), g.H())
      score := a.evaluate(
          possibleFuture,
          nextPos,
          depthLeft-1,
          maxDepth,
          evolve,
          deterministic,
          memo,
          reducer,)
      moveToScore[m] = append(moveToScore[m], score)
    }
  }
  bestMove, bestScore := reducer(moveToScore)
  if memoize && depthLeft > 3 {
    memo[key] = memoEntry{ move: bestMove, score: bestScore }
  }
  return gamma * bestScore + aliveReward
}

func stateKey(
    g      *substrates.Grid2d,
    pos    substrates.Pos,
    depth  int,
) string {
  var b strings.Builder
  for y := 0; y < g.H(); y++ {
    for x := 0; x < g.W(); x++ {
      b.WriteByte('0' + byte(g.XY(x, y)))
    }
    b.WriteByte('|')
  }
  b.WriteString(fmt.Sprintf("P%d,%d|D%d|", pos.X, pos.Y, depth))
  return b.String()
}

func rolloutBudget(depth, horizon int, deterministic bool) int {
  const (
    maxRollouts = 3
    minRollouts = 1
  )
  return rolloutBudgetMM(
      depth, horizon, deterministic, minRollouts, maxRollouts)
}

func rolloutBudgetMM(depth, horizon int, deterministic bool,
    minimum, maximum int,
  ) int {
  if horizon <= 0 {
    panic("invalid horizon")
  }
  if depth > horizon {
    panic("depth exceeds horizon")
  }
  if deterministic {
    return 1
  }
  frac := float64(depth) / float64(horizon)
  runs := int(math.Round(
      float64(minimum) + frac*float64(maximum-minimum),
  ))
  if runs > maximum {
    return maximum
  }
  return runs
}
