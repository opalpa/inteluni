package substrates
import "fmt"

type Pos struct {
  X int
  Y int
}

type Move struct {
  dx int
  dy int
}

func (m Move) DX() int { return m.dx }
func (m Move) DY() int { return m.dy }

var (
  South = Move{dx:  0, dy:  1}
  West  = Move{dx: -1, dy:  0}
  North = Move{dx:  0, dy: -1}
  East  = Move{dx:  1, dy:  0}
  Stay  = Move{dx:  0, dy:  0}
)

type Grid2d struct {
  w int
  h int
  v [][]int
}

type Evolver func(g *Grid2d, rng *SplitMix64) *Grid2d

func NewGrid2d(w, h int) *Grid2d {
  v := make([][]int, h)
  for y := range v {
    v[y] = make([]int, w)
  }
  return &Grid2d{
    w: w,
    h: h,
    v: v,
  }
}

func (g *Grid2d) Map(fn func(x, y, val int) int) {
  for y := 0; y < g.h; y++ {
    for x := 0; x < g.w; x++ {
      g.v[y][x] = fn(x, y, g.v[y][x])
    }
  }
}

func (g *Grid2d) SetXY(x, y, val int) {
  if !g.InBoundsXY(x, y) {
    panic("out of bounds access")
  }
  g.v[y][x] = val
}

func (g *Grid2d) InBoundsXY(x, y int) bool {
  return x >= 0 && x < g.w &&
         y >= 0 && y < g.h
}
func (g *Grid2d) InBounds(p Pos) bool {
  return g.InBoundsXY(p.X, p.Y)
}

func (g *Grid2d) XY(x, y int) int {
  if !g.InBoundsXY(x, y) {
    panic("out of bounds access")
  }
  return g.v[y][x]
}
func (g *Grid2d) Get(p Pos) int {
  return g.XY(p.X, p.Y)
}

func (g *Grid2d) H() int { return g.h }
func (g *Grid2d) W() int { return g.w }

// Deep copy of the grid.
func (g *Grid2d) Clone() *Grid2d {
  dup := NewGrid2d(g.w, g.h)
  for y := 0; y < g.h; y++ {
    copy(dup.v[y], g.v[y])
  }
  return dup
}

func (g *Grid2d) OntoStdout() {
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

func (g *Grid2d) OntoStdoutAgent(ax, ay int) {
  for y := 0; y < g.H(); y++ {
    for x := 0; x < g.W(); x++ {
      val := g.XY(x, y)
      switch val {
        case 0:
          if x == ax && y == ay {
            fmt.Print("A")
          } else {
            fmt.Print("_")
          }
        case 1:
          if x == ax && y == ay {
            fmt.Print("a")
          } else {
            fmt.Print("#")
          }
        default:
          emsg := fmt.Sprintf("unexpected %d at (%d, %d)", val, x, y)
          panic(emsg)
      }
    }
    fmt.Println()
  }
}
