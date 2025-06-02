package agents
import "oscarkilo.com/inteluni/substrates"

// Agent is an interface for an autonomous entity moving in the world.
//
// Each agent has a position, unique ID, and foresight depth.
// Agents choose a Move based on the current state of the grid.
// They know the entire substrate (grid) state at time of decision.
// They can simulate the universe with a provided Evolver.
//
// They do NOT know:
//   - The future moves of other agents.
//   - Other agents' positions.
//   - Other agents' policies.
//   - The random number seeds, if any, used by the universe.
// We are modeling intellect in universe, not adversarial agent game.
type Agent interface {
  ID() int
  Pos() substrates.Pos
  Foresight() int
  Decide(
      grid *substrates.Grid2d,
      evolve substrates.Evolver,
      deterministic bool,
  ) substrates.Move
  Apply(m substrates.Move, g *substrates.Grid2d)
}

// baseAgent handles shared identity and position logic.
type baseAgent struct {
  id        int
  pos       substrates.Pos
  foresight int
  rng       *substrates.SplitMix64
}

func (a *baseAgent) ID() int {
  return a.id
}

func (a *baseAgent) Pos() substrates.Pos {
  return a.pos
}

func (a *baseAgent) Foresight() int {
  return a.foresight
}

// Apply executes the move and wraps around the grid (toroidal world).
func (a *baseAgent) Apply(m substrates.Move, g *substrates.Grid2d) {
  w := g.W()
  h := g.H()
  a.pos.X = (a.pos.X + m.DX() + w) % w
  a.pos.Y = (a.pos.Y + m.DY() + h) % h
}

// ReactiveAgent is a depth‑0 agent that only sees the present.
// Decision is made based on immediate neighborhood conditions.
type ReactiveAgent struct {
  baseAgent
}

// NewReactiveAgent creates a reactive agent with foresight depth 0.
func NewReactiveAgent(
    id int,
    pos substrates.Pos,
    rng *substrates.SplitMix64,
) *ReactiveAgent {
  return &ReactiveAgent{
    baseAgent: baseAgent{
      id:        id,
      pos:       pos,
      foresight: 0,
      rng:       rng,
    },
  }
}

// Decide chooses one of the 5 reachable actions (N/E/S/W/Stay).
//
// Decision strategy:
//   - Evaluate all adjacent cells (wraparound grid).
//   - Prefer empty squares (value == 0).
//   - Break ties randomly.
//   - Avoid obvious collisions if possible.
func (a *ReactiveAgent) Decide(
    g *substrates.Grid2d,
    _ substrates.Evolver,
    _ bool,
  ) substrates.Move {
  moves := []substrates.Move{
    substrates.North,
    substrates.South,
    substrates.East,
    substrates.West,
    substrates.Stay,
  }

  w := g.W()
  h := g.H()
  px := a.pos.X
  py := a.pos.Y

  var open []substrates.Move

  for _, m := range moves {  // consider all of N/S/E/W/Stay
    nx := (px + m.DX() + w) % w
    ny := (py + m.DY() + h) % h

    empty := g.XY(nx, ny) == 0
    if empty {
      open = append(open, m)
    }
  }

  if len(open) == 0 {  // no open cells out of the 5
    return substrates.Stay  // all cells look like death
  }

  return open[a.rng.Intn(len(open))]
}


// PredictiveAgent simulates future universe evolution before deciding.
type PredictiveAgent struct {
  baseAgent
}

// NewPredictiveAgent creates a foresight-enabled agent.
func NewPredictiveAgent(
    id int,
    pos substrates.Pos,
    foresight int,
    rng *substrates.SplitMix64,
) *PredictiveAgent {
  return &PredictiveAgent{
    baseAgent: baseAgent{
      id:        id,
      pos:       pos,
      foresight: foresight,
      rng:       rng,
    },
  }
}

func (a *PredictiveAgent) Decide(
    g *substrates.Grid2d,
    evolve substrates.Evolver,
    deterministic bool,
) substrates.Move {
  m := a.predictiveDecide(g, evolve, deterministic)
  return m
}
