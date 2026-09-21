package game

import (
	"go-tcg/internal/constants"
	"go-tcg/internal/model"
)

type schedulerKind string

const (
	schedulerStable   schedulerKind = "stable"
	schedulerFinished schedulerKind = "finished"
)

// schedulerFrame 保存標準回合排程目前停留的位置與行動機會。
type schedulerFrame struct {
	Kind              schedulerKind   `json:"kind"`
	TurnPlayer        *model.Player   `json:"turn_player"`
	Phase             constants.Phase `json:"phase,omitempty"`
	OpportunityHolder *model.Player   `json:"opportunity_holder,omitempty"`
	ConsecutivePasses int             `json:"consecutive_passes"`
	TurnNumber        uint64          `json:"turn_number"`
}
