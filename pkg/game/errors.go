package game

import "errors"

var ErrCardNotImplemented = errors.New("card not implemented yet")

var (
	ErrInvalidPhase               = errors.New("action is not allowed in the current phase")
	ErrNotYourTurn                = errors.New("action is only allowed for the active player")
	ErrNoPlayersInGame            = errors.New("cannot apply commands without players")
	ErrGameNotStarted             = errors.New("game has not been started yet")
	ErrCardNotFound               = errors.New("card was not found")
	ErrCharityRequired            = errors.New("player must reduce hand size to 5 before ending the turn")
	ErrCharityTransferUnsupported = errors.New("charity transfer to another player is not implemented in this MVP")
	ErrNoRevealedDoorCard         = errors.New("there is no revealed door card to acknowledge")
)
