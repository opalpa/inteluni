package universes
import "oscarkilo.com/inteluni/substrates"

type Universe interface {
  Grid() *substrates.Grid2d           // current substrate view
  Advance()                           // step forward by one tick
  MakeEvolver() substrates.Evolver    // used to see possible futures
  Deterministic() bool                // is this universe deterministic
}
