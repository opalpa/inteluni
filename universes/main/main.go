package main
import "time"
import "fmt"
import "oscarkilo.com/inteluni/substrates"
import "oscarkilo.com/inteluni/universes"

func printGrid(g *substrates.Grid2d) {
  for y := 0; y < g.H(); y++ {
    for x := 0; x < g.W(); x++ {
      val := g.XY(x, y)
      switch val {
        case 0:
          fmt.Print("_")
        case 1:
          fmt.Print("#")
        default:
          emsg := fmt.Sprintf("unexpected %d at (%d, %d)", val, x, y)
          panic(emsg)
      }
    }
    fmt.Println()
  }
}

func main() {
  width := 10
  height := 6
  noise := 0.2
  complexity := 30
  seed := time.Now().UnixNano()
  rng := substrates.NewSplitMix64(uint64(seed))

  u := universes.NewNoisyUniverse(
      width, height, noise, complexity, rng)

  fmt.Println("initial grid:")
  printGrid(u.Grid())
  fmt.Println()

  for i := 0; i < 5; i++ {
    fmt.Println("iteration ", i+1)
    u.Advance()
    printGrid(u.Grid())
    fmt.Println()
  }
}
