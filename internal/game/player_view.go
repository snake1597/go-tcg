package game

import (
	"fmt"
	"go-tcg/internal/constants"
	"go-tcg/internal/model"
	tcgErrors "go-tcg/internal/tcg_errors"
)

type LegalAction struct {
	Handle   ViewHandle           `json:"handle"`
	Kind     constants.ActionKind `json:"kind"`
	CardName string               `json:"card_name,omitempty"`
	// ReserveCost 與 ReserveOptions 定義 activation 必須提交的手牌付款張數與可選 handles。
	ReserveCost    int           `json:"reserve_cost,omitempty"`
	ReserveOptions []VisibleCard `json:"reserve_options,omitempty"`
	// MemoryPaymentRequired 指出本次 Memory Cost 至少必須使用的非隨機付款來源數量。
	MemoryPaymentRequired int           `json:"memory_payment_required,omitempty"`
	MemoryPaymentOptions  []VisibleCard `json:"memory_payment_options,omitempty"`
	HeuristicRank         int           `json:"heuristic_rank"`
}

// VisibleChoice 提供 PendingChoice 選項的玩家可見描述與啟發式優先級。
// Handle 是提交用的不透明識別；其餘欄位只由既有 PlayerView 可見資料導出，不包含內部 ID。
type VisibleChoice struct {
	Handle        ViewHandle `json:"handle"`
	CardName      string     `json:"card_name,omitempty"`
	HeuristicRank int        `json:"heuristic_rank"`
}

type VisibleChampion struct {
	Owner    *model.Player `json:"owner"`
	CardName string        `json:"card_name"`
	Power    int           `json:"power"`
	Life     int           `json:"life"`
	Damage   int           `json:"damage"`
	Rested   bool          `json:"rested"`
	Taunt    bool          `json:"taunt"`
}

type VisibleCard struct {
	Handle ViewHandle `json:"handle"`
	Name   string     `json:"name"`
}

type VisibleFieldObject struct {
	Owner    *model.Player  `json:"owner"`
	CardName string         `json:"card_name"`
	Types    []string       `json:"types"`
	Rested   bool           `json:"rested"`
	Damage   int            `json:"damage"`
	Counters map[string]int `json:"counters,omitempty"`
}

type VisibleEffectStackItem struct {
	Kind       string        `json:"kind"`
	Controller *model.Player `json:"controller"`
	SourceName string        `json:"source_name"`
}

type VisibleEvent struct {
	Kind     string `json:"kind"`
	CardName string `json:"card_name"`
}

type PendingChoice struct {
	Options []ViewHandle    `json:"options"`
	Choices []VisibleChoice `json:"choices"`
	CanPass bool            `json:"can_pass,omitempty"`
}

type PlayerView struct {
	Revision          uint64                   `json:"revision"`
	Finished          bool                     `json:"finished"`
	Winner            *model.Player            `json:"winner,omitempty"`
	Diagnostic        string                   `json:"diagnostic,omitempty"`
	TurnPlayer        *model.Player            `json:"turn_player,omitempty"`
	TurnNumber        uint64                   `json:"turn_number,omitempty"`
	Phase             constants.Phase          `json:"phase,omitempty"`
	OpportunityHolder *model.Player            `json:"opportunity_holder,omitempty"`
	DecisionPlayer    *model.Player            `json:"decision_player,omitempty"`
	Champions         []VisibleChampion        `json:"champions,omitempty"`
	Hand              []VisibleCard            `json:"hand"`
	Field             []VisibleFieldObject     `json:"field"`
	EffectsStack      []VisibleEffectStackItem `json:"effects_stack"`
	Cards             []VisibleCard            `json:"cards"`
	VisibleEvents     []VisibleEvent           `json:"visible_events"`
	LegalActions      []LegalAction            `json:"legal_actions"`
	PendingChoice     *PendingChoice           `json:"pending_choice,omitempty"`
}

// PlayerView 投影指定玩家可見的牌、事件、合法行動與待選項目。
// 提交行動須使用此視圖的 revision 與 handle；未知玩家回傳 ErrUnknownPlayer。
func (g *Game) PlayerView(player *model.Player) (PlayerView, error) {
	if !g.hasPlayer(player) {
		return PlayerView{}, fmt.Errorf("%w %q", tcgErrors.ErrUnknownPlayer, player)
	}
	return PlayerView{
		Revision:          g.state.Revision,
		Finished:          g.state.Finished,
		Winner:            g.state.Winner,
		Diagnostic:        g.state.Diagnostic,
		TurnPlayer:        g.state.Scheduler.TurnPlayer,
		TurnNumber:        g.state.Scheduler.TurnNumber,
		Phase:             g.state.Scheduler.Phase,
		OpportunityHolder: g.state.Scheduler.OpportunityHolder,
		DecisionPlayer:    g.decisionPlayer(),
		Champions: g.visibleChampions(
			player,
		),
		Hand: g.visibleHand(
			player,
		),
		Field:        g.visibleField(),
		EffectsStack: g.visibleEffectsStack(),
		Cards: g.visibleCards(
			player,
		),
		VisibleEvents: g.visibleEvents(
			player,
		),
		LegalActions: g.legalActions(
			player,
		),
		PendingChoice: g.pendingChoice(
			player,
		),
	}, nil
}

// decisionPlayer 回傳目前依 scheduler 或 PendingChoice 應取得輸入的玩家。
// 輸入為目前單局狀態；輸出為待決玩家或 nil，無副作用且不透露任何私人卡牌資訊。
func (g *Game) decisionPlayer() *model.Player {
	if g.state.Knowledge.Choice != nil {
		return g.state.Knowledge.Choice.Actor
	}
	if g.state.Finished {
		return nil
	}
	if g.state.Scheduler.OpportunityHolder != nil {
		return g.state.Scheduler.OpportunityHolder
	}
	if g.state.Scheduler.Phase == constants.PhaseMaterialize {
		return g.state.Scheduler.TurnPlayer
	}
	return nil
}
