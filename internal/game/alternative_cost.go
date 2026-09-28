package game

import (
	"fmt"

	"go-tcg/internal/model"
	tcgErrors "go-tcg/internal/tcg_errors"
)

type alternativeCostSpec struct {
	Selection cardSelectionSpec
	Payment   alternativeCostPayment
}

type alternativeCostPayment string

const alternativeCostBanishGraveyard alternativeCostPayment = "banish_graveyard"

// alternativeCostDeclaration 保存尚未提交的替代費用選牌；取消不移動任何卡牌。
type alternativeCostDeclaration struct {
	Controller *model.Player    `json:"controller"`
	Source     cardInstanceID   `json:"source"`
	Selected   []cardInstanceID `json:"selected"`
}

// alternativeCostFor 回傳來源卡宣告的替代費用；未宣告時回傳 false，無副作用。
func (g *Game) alternativeCostFor(source cardInstanceID) (alternativeCostSpec, bool) {
	if _, exists := g.state.Cards[source]; !exists {
		return alternativeCostSpec{}, false
	}
	ability, exists := g.compiledAlly(source)
	if !exists || ability.alternative == nil {
		return alternativeCostSpec{}, false
	}
	return *ability.alternative, true
}

// alternativeCostCards 搜尋第一組完整合法付款；沒有組合時回傳 nil，無副作用。
func (g *Game) alternativeCostCards(player *model.Player, source cardInstanceID) []cardInstanceID {
	spec, exists := g.alternativeCostFor(source)
	if !exists {
		return nil
	}
	return g.findCardSelection(player, spec.Selection, source, nil)
}

// beginAlternativeCostDeclaration 建立可取消的逐張選牌宣告；來源卡與付款尚不移動。
func (g *Game) beginAlternativeCostDeclaration(player *model.Player, source cardInstanceID) error {
	if !g.canActivateAction(player, source) || cardIndex(g.state.Zones[player.UID].Hand, source) < 0 || len(g.alternativeCostCards(player, source)) == 0 {
		return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, source)
	}
	declaration := &alternativeCostDeclaration{
		Controller: player,
		Source:     source,
	}
	g.state.Knowledge.AlternativeCost = declaration
	g.setAlternativeCostChoice(declaration)
	return nil
}

// submitAlternativeCostChoice 收下一張可完成付款的卡牌；完成時一次提交費用與啟動。
func (g *Game) submitAlternativeCostChoice(player *model.Player, card cardInstanceID) error {
	declaration := g.state.Knowledge.AlternativeCost
	if declaration == nil || !samePlayer(declaration.Controller, player) || !containsCard(g.alternativeCostChoiceCards(declaration), card) {
		return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, card)
	}
	spec, exists := g.alternativeCostFor(declaration.Source)
	if !exists {
		return fmt.Errorf("invalid alternative cost source %q", declaration.Source)
	}
	selected := append(append([]cardInstanceID(nil), declaration.Selected...), card)
	if g.validCardSelection(player, spec.Selection, declaration.Source, selected) {
		completed := *declaration
		completed.Selected = selected
		if err := g.commitAlternativeCost(&completed); err != nil {
			return fmt.Errorf("commit alternative cost: %w", err)
		}
		g.state.Knowledge.AlternativeCost = nil
		g.state.Knowledge.Choice = nil
		return nil
	}
	declaration.Selected = selected
	g.setAlternativeCostChoice(declaration)
	return nil
}

// setAlternativeCostChoice 顯示仍有完整付款組合的下一張牌，並允許玩家取消宣告。
func (g *Game) setAlternativeCostChoice(declaration *alternativeCostDeclaration) {
	options := make([]objectID, 0)
	for _, card := range g.alternativeCostChoiceCards(declaration) {
		options = append(options, objectID(card))
	}
	g.setDeclarationChoice(declaration.Controller, options)
	g.state.Knowledge.Choice.CanPass = true
}

// alternativeCostChoiceCards 依來源區域順序回傳每張仍可完成精確付款的候選牌，無副作用。
func (g *Game) alternativeCostChoiceCards(declaration *alternativeCostDeclaration) []cardInstanceID {
	spec, exists := g.alternativeCostFor(declaration.Source)
	if !exists {
		return nil
	}
	return g.nextCardSelectionOptions(declaration.Controller, spec.Selection, declaration.Source, declaration.Selected)
}

// commitAlternativeCost 重新驗證所有來源與付款卡，成功時一次移動付款並建立 Ally activation。
func (g *Game) commitAlternativeCost(declaration *alternativeCostDeclaration) error {
	spec, exists := g.alternativeCostFor(declaration.Source)
	if !exists || !g.validCardSelection(declaration.Controller, spec.Selection, declaration.Source, declaration.Selected) {
		return fmt.Errorf("invalid alternative cost")
	}
	zones := g.state.Zones[declaration.Controller.UID]
	if cardIndex(zones.Hand, declaration.Source) < 0 {
		return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, declaration.Source)
	}
	switch spec.Payment {
	case alternativeCostBanishGraveyard:
		var err error
		zones, err = banishCardsFromGraveyard(zones, declaration.Selected)
		if err != nil {
			return fmt.Errorf("pay alternative cost: %w", err)
		}
	default:
		return fmt.Errorf("unknown alternative cost payment %q", spec.Payment)
	}
	index := cardIndex(zones.Hand, declaration.Source)
	zones.Hand = append([]cardInstanceID(nil), zones.Hand...)
	zones.Hand = removeCardAt(zones.Hand, index)
	return g.commitAllyActivationTransaction(
		declaration.Controller,
		declaration.Source,
		zones,
	)
}
