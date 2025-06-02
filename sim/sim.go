package sim
import "fmt"
import "oscarkilo.com/inteluni/substrates"
import "oscarkilo.com/inteluni/universes"
import "oscarkilo.com/inteluni/agents"
import "oscarkilo.com/inteluni/metrics"
import "sync"

func SimulateSteps(
    u universes.Universe,
    agentsPop *[]agents.Agent,
    stepsPerRun int,
) []*substrates.Grid2d {
  frames := make([]*substrates.Grid2d, 0, stepsPerRun+1,)
  frames = append(frames, u.Grid().Clone(),)
  for step := 0; step < stepsPerRun; step++ {
    moves := collectMoves(u, *agentsPop)
    u.Advance()
    frames = append(frames, u.Grid().Clone(),)
    applyMoves(*agentsPop, moves, u.Grid())
    *agentsPop, _ = resolveCollisions(*agentsPop, u.Grid())
    if len(*agentsPop) == 0 {
      break
    }
  }
  return frames
}

func collectMoves(
    u universes.Universe,
    agentsPop []agents.Agent,
) []substrates.Move {
  moves := make([]substrates.Move, len(agentsPop),)
  evolver := u.MakeEvolver()
  det := u.Deterministic()
  for i, ag := range agentsPop {
    moves[i] = ag.Decide(u.Grid(), evolver, det,)
  }
  return moves
}

func applyMoves(
    agentsPop []agents.Agent,
    moves []substrates.Move,
    grid *substrates.Grid2d,
) {
  for i, ag := range agentsPop {
    ag.Apply(moves[i], grid)
  }
}

func resolveCollisions(
  agentsPop []agents.Agent,
  grid *substrates.Grid2d,
) ([]agents.Agent, []agents.Agent) {
  survivors := make([]agents.Agent, 0, len(agentsPop))
  dead := make([]agents.Agent, 0, len(agentsPop))
  for _, ag := range agentsPop {
    pos := ag.Pos()
    if grid.XY(pos.X, pos.Y) == 1 {
      // obstacle death: agent moved into an obstacle
      dead = append(dead, ag)
      continue
    }
    survivors = append(survivors, ag)
  }
  return survivors, dead
}

var headerOnce sync.Once

func Report(
    id int,
    frames []*substrates.Grid2d,
    univ universes.Universe,
    noise float64,
    complexity int,
    foresight int,
    originalNumReactive int,
    originalNumPredictive int,
    agentsPop []agents.Agent,
    rng *substrates.SplitMix64,
) {
  headerOnce.Do(func() {
    fmt.Printf("noise,complexity,foresight,K,TauL,C_react,C_pred\n")
  })
  var reactCount, predCount int
  for _, ag := range agentsPop {
    switch ag.(type) {
      case *agents.ReactiveAgent:
        reactCount++
      case *agents.PredictiveAgent:
        predCount++
      default:
        panic("unknown agent type in report")
    }
  }
  K := metrics.KolmogorovProxy(frames)
  tau := metrics.TauL(univ, rng)
  collisionsReactive := originalNumReactive - reactCount
  collisionsPredictive := originalNumPredictive - predCount
  fmt.Printf(
      "%0.2f,%d,%d,%0.3f,%0.3f,%d,%d\n",
      noise, complexity, foresight, K, tau,
      collisionsReactive, collisionsPredictive,
  )
}
