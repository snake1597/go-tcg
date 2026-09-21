package game

import "go-tcg/internal/model"

// objectID 識別單局遊戲中的場上物件。
type objectID string

type championObject struct {
	ID             objectID         `json:"id"`
	Card           cardInstanceID   `json:"card"`
	Owner          *model.Player    `json:"owner"`
	InnerLineage   []cardInstanceID `json:"inner_lineage"`
	Rested         bool             `json:"rested"`
	Counters       map[string]int   `json:"counters"`
	CombatRole     string           `json:"combat_role"`
	TauntUntilTurn uint64           `json:"taunt_until_turn"`
	Damage         int              `json:"damage"`
	DamageTurn     uint64           `json:"damage_turn,omitempty"`
}

type fieldObject struct {
	ID       objectID       `json:"id"`
	Card     cardInstanceID `json:"card"`
	Owner    *model.Player  `json:"owner"`
	Types    []string       `json:"types"`
	Rested   bool           `json:"rested"`
	Counters map[string]int `json:"counters,omitempty"`
	Damage   int            `json:"damage"`
}
