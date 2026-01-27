package game

type PlayCardCommand struct {
	PlayerID string `json:"playerID"`
	CardID   string `json:"cardID"`
	_Type    string `json:"type"`
}

func (pcc PlayCardCommand) GetPlayerID() string {
	return pcc.PlayerID
}

func (pcc PlayCardCommand) GetCardID() string {
	return pcc.CardID
}

func (pcc PlayCardCommand) Type() string {
	return pcc._Type
}

type DrawCardCommand struct {
	PlayerID string `json:"playerID"`
	_Type    string `json:"type"`
}

func (pcc DrawCardCommand) Type() string {
	return pcc._Type
}

func (pcc DrawCardCommand) GetPlayerID() string {
	return pcc.PlayerID
}
