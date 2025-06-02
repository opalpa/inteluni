package main
import "flag"
import "time"
import "oscarkilo.com/inteluni/substrates"
import "oscarkilo.com/inteluni/universes"
import "oscarkilo.com/inteluni/agents"
import "oscarkilo.com/inteluni/sim"

const (
  gridW         = 32
  gridH         = 32
  stepsPerRun   = 100

  noiseStart    = 0.1
  noiseEnd      = 0.9
  noiseStep     = 0.1

  complexStart  = 10
  complexEnd    = 90
  complexStep   = 10

  foresightMin  = 2
  foresightMax  = 5
  foresightStep = 1

  numReactive   = 5
  numPredictive = 5
)

var seedFlag = flag.Uint64(
    "seed", uint64(time.Now().UnixNano()), "random seed",)

func main() {
  flag.Parse()
  runID := 0
  for noise := noiseStart; noise <= noiseEnd; noise += noiseStep {
    for comp := complexStart; comp <= complexEnd; comp += complexStep {
      for fs := foresightMin; fs <= foresightMax; fs += foresightStep {
        runSimulation(runID, noise, comp, fs, *seedFlag,)
        runID++
      }
    }
  }
}

func runSimulation(
    id int, noise float64, comp int, fores int, seed uint64,) {
  rng := substrates.NewSplitMix64(seed + uint64(id))
  u := universes.NewNoisyUniverse(
      gridW, gridH, noise, comp, rng,
  )
  agentsPop := agents.Spawn(
      u.Grid(), numReactive, numPredictive, fores, rng,
  )
  frames := sim.SimulateSteps(u, &agentsPop, stepsPerRun)
  sim.Report(id, frames, u, noise, comp, fores,
      numReactive, numPredictive, agentsPop, rng)
}
