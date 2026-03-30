package game

import "errors"

var ErrCardNotImplemented = errors.New("card not implemented yet")

var (
	ErrInvalidPhase  = errors.New("action is not allowed in the current phase")
	ErrNotYourTurn   = errors.New("action is only allowed for the active player")
	ErrNoPlayersInGame = errors.New("cannot apply commands without players")
)
