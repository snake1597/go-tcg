package game

// playerZones 保存單一玩家各區域中的卡牌實例。
type playerZones struct {
	MainDeck        []cardInstanceID `json:"main_deck"`
	Hand            []cardInstanceID `json:"hand"`
	MaterialDeck    []cardInstanceID `json:"material_deck"`
	Memory          []cardInstanceID `json:"memory"`
	Banishment      []cardInstanceID `json:"banishment"`
	Graveyard       []cardInstanceID `json:"graveyard"`
	OutsideGamePool []cardInstanceID `json:"outside_game_pool"`
}
