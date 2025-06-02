package main
import "flag"
import "time"
import "oscarkilo.com/inteluni/substrates"
import "oscarkilo.com/inteluni/universes"
import "oscarkilo.com/inteluni/agents"
import "oscarkilo.com/inteluni/sim"
import "os"
import "runtime/pprof"

const (
  gridW         = 32
  gridH         = 32
  stepsPerRun   = 50

  noiseStart    = 0.00
  noiseEnd      = 0.70
  noiseStep     = 0.05

  complexStart  = 10
  complexEnd    = 60
  complexStep   = 5

  foresightMin  = 1
  foresightMax  = 5
  foresightStep = 1

  numReactive   = 5
  numPredictive = 5

  prof          = false
)

var seedFlag = flag.Uint64(
    "seed", uint64(time.Now().UnixNano()), "random seed",)

func main() {
  flag.Parse()
  if prof {
    f, err := os.Create("profile.out")
    if err != nil {
      panic(err)
    }
    pprof.StartCPUProfile(f)
    defer pprof.StopCPUProfile()
  }
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
  u := universes.NewGameOfNoiseUniverse(
      gridW, gridH, noise, comp, rng,
  )
  agentsPop := agents.Spawn(
      u.Grid(), numReactive, numPredictive, fores, rng,
  )
  frames := sim.SimulateSteps(u, &agentsPop, stepsPerRun)
  sim.Report(id, frames, u, noise, comp, fores,
      numReactive, numPredictive, agentsPop, rng)
}
