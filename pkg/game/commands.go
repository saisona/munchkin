package game

type PlayCardCommand struct {
	playerID string
	cardID   string
	_type    string
}

func (pcc PlayCardCommand) GetPlayerID() string {
	return pcc.playerID
}

func (pcc PlayCardCommand) GetCardID() string {
	return pcc.cardID
}

func (pcc PlayCardCommand) Type() string {
	return pcc._type
}

type DrawCardCommand struct {
	playerID string
	_type    string
}

func (pcc DrawCardCommand) Type() string {
	return pcc._type
}

func (pcc DrawCardCommand) GetPlayerID() string {
	return pcc.playerID
}
