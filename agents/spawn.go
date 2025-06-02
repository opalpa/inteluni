package agents
import "oscarkilo.com/inteluni/substrates"

func Spawn(
    grid *substrates.Grid2d,
    numReactive, numPredictive, foresight int,
    rng *substrates.SplitMix64,
) []Agent {

  total := numReactive + numPredictive
  totalCells := grid.W() * grid.H()
  if total > totalCells {
    panic("not enough cells to spawn all agents")
  }

  result := make([]Agent, 0, total)
  occupied := make(map[substrates.Pos]struct{}, total)
      // Initially we make agents start on different cells.
      // Can change that.

  // Shuffle agent type spawn order
  types := make([]string, 0, total)
  for i := 0; i < numReactive; i++ {
    types = append(types, "reactive")
  }
  for i := 0; i < numPredictive; i++ {
    types = append(types, "predictive")
  }
  for i := total - 1; i > 0; i-- {
    j := rng.Intn(i+1)
    types[i], types[j] = types[j], types[i]
  }

  nextID := 1
  attempts := 0
  maxAttempts := totalCells * 5
  remainingReactive := numReactive  // used to double check correcntess
  remainingPredictive := numPredictive

  for len(result) < total {
    if attempts >= maxAttempts {
      panic("failed to spawn all agents: too many attempts")
    }
    attempts++

    x := rng.Intn(grid.W())
    y := rng.Intn(grid.H())
    pos := substrates.Pos{X: x, Y: y}

    if grid.XY(x, y) != 0 {
      continue // obstacle
    }
    if _, taken := occupied[pos]; taken {
      continue // already has agent
    }

    agentType := types[len(result)]
    var ag Agent
    agentRng  := rng.NewFromSelf()
    if agentType == "reactive" {
      ag = NewReactiveAgent(nextID, pos, agentRng)
      remainingReactive--
    } else if agentType == "predictive" {
      ag = NewPredictiveAgent(nextID, pos, foresight, agentRng)
      remainingPredictive--
    } else {
      panic("unknown agent type: " + agentType)
    }

    result = append(result, ag)
    occupied[pos] = struct{}{}
    nextID++
  }

  if remainingReactive != 0 {
    panic("not all reactive agents spawned")
  }
  if remainingPredictive != 0 {
    panic("not all predictive agents spawned")
  }

  return result
}
