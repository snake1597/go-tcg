package tcgerrors

import "fmt"

var (
	ErrGameFinished      = fmt.Errorf("game is finished")
	ErrStaleRevision     = fmt.Errorf("stale revision")
	ErrUnknownPlayer     = fmt.Errorf("unknown player")
	ErrInvalidViewHandle = fmt.Errorf("invalid view handle")
)
