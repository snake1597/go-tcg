package tcgerrors

import "errors"

var (
	ErrGameFinished     = errors.New("game is finished")
	ErrStaleRevision    = errors.New("stale revision")
	ErrUnknownPlayer    = errors.New("unknown player")
	ErrUnknownInputKind = errors.New("unknown input kind")
)
