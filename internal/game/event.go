package game

import "go-tcg/internal/model"

type gameEvent struct {
	Sequence uint64         `json:"sequence"`
	Kind     string         `json:"kind"`
	Card     cardInstanceID `json:"card"`
}

// eventBatch 保存同一原因產生的一批可重播事件。
type eventBatch struct {
	Player       *model.Player      `json:"player"`
	Cause        string             `json:"cause"`
	ParentFlow   string             `json:"parent_flow,omitempty"`
	Simultaneous bool               `json:"simultaneous"`
	Events       []gameEvent        `json:"events"`
	CauseChain   []replacementCause `json:"cause_chain,omitempty"`
}
