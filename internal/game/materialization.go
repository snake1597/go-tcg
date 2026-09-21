package game

import (
	"fmt"
	"go-tcg/internal/constants"
	"go-tcg/internal/model"
	tcgErrors "go-tcg/internal/tcg_errors"
)

const (
	spiritOfFireCardID       CardID = "LMyKyVC2O9"
	tonorisCardID            CardID = "zb14m4c8lj"
	tonorisOnEnterCause             = "ability:zb14m4c8lj:front:on-enter-taunt"
	impactHammerCardID       CardID = "chsbalegbs"
	impactHammerOnWieldCause        = "ability:chsbalegbs:front:on-wield-self-damage"
)

// legalMaterializations 回傳目前玩家在 Materialize 階段可從 Material Deck 使用的牌。
// 輸入為目前行動玩家；輸出依 Material Deck 原始順序排列；不會改變對局狀態。
func (g *Game) legalMaterializations(player *model.Player) []cardInstanceID {
	zones := g.state.Zones[player.UID]
	materialDeckSize := len(zones.MaterialDeck)
	cards := make([]cardInstanceID, 0, materialDeckSize)
	for _, card := range zones.MaterialDeck {
		if g.canMaterialize(
			player,
			card,
		) {
			cards = append(cards, card)
		}
	}
	return cards
}

// canMaterialize 判斷指定 Material Deck 牌是否符合目前階段、所有權與牌類型的進場條件。
// 輸入為玩家與候選牌；輸出為可否 materialize；不會改變對局狀態。
func (g *Game) canMaterialize(player *model.Player, card cardInstanceID) bool {
	scheduler := g.state.Scheduler
	if scheduler.Kind != schedulerStable || scheduler.Phase != constants.PhaseMaterialize || !samePlayer(scheduler.TurnPlayer, player) || scheduler.OpportunityHolder != nil {
		return false
	}
	candidate, exists := g.state.Cards[card]
	if !exists || !samePlayer(candidate.Owner, player) {
		return false
	}
	if len(g.state.Zones[player.UID].Memory) < g.characteristicsForCard(card).MemoryCost {
		return false
	}
	if containsString(candidate.Types, "REGALIA") {
		return true
	}
	return g.canMaterializeChampionLevelUp(player, candidate)
}

// canMaterializeChampionLevelUp 判斷候選 Champion 是否能覆蓋目前 Spirit of Fire Champion。
// 輸入為玩家與候選牌；輸出為是否符合既有 Tonoris 升級條件；不會改變對局狀態。
func (g *Game) canMaterializeChampionLevelUp(player *model.Player, candidate cardInstance) bool {
	if candidate.Definition != tonorisCardID {
		return false
	}
	champion, exists := g.state.Champions[player.UID]
	if !exists {
		return false
	}
	current, exists := g.state.Cards[champion.Card]
	if !exists || current.Definition != spiritOfFireCardID || candidate.Level != current.Level+1 {
		return false
	}
	return candidate.Level == current.Level+1
}

// materialize 驗證玩家對 Material Deck 牌的宣告，並將付款與 Stack 建立交由付款流程處理。
// 輸入為玩家與其公開的 materialization 牌；輸出為驗證或付款錯誤；成功時會改變區域與 Stack。
func (g *Game) materialize(player *model.Player, card cardInstanceID) error {
	if !g.canMaterialize(
		player,
		card,
	) {
		return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, card)
	}
	zones := g.state.Zones[player.UID]
	materialDeckIndex := cardIndex(
		zones.MaterialDeck,
		card,
	)
	if materialDeckIndex < 0 {
		return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, card)
	}
	if err := g.payMaterialization(player, card); err != nil {
		return fmt.Errorf("pay materialization: %w", err)
	}
	return nil
}

// payMaterialization 重新檢查進場與付款條件，再移走 Material Deck 牌並隨機放逐 Memory。
// 來源保留在 EffectSources，materialization 效果入堆疊後授予玩家行動機會；此時尚未進場。
func (g *Game) payMaterialization(player *model.Player, card cardInstanceID) error {
	if !g.canMaterialize(
		player,
		card,
	) {
		return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, card)
	}
	zones := g.state.Zones[player.UID]
	materialDeckIndex := cardIndex(
		zones.MaterialDeck,
		card,
	)
	if materialDeckIndex < 0 {
		return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, card)
	}
	candidate := g.state.Cards[card]
	zones.MaterialDeck = removeCardAt(
		zones.MaterialDeck,
		materialDeckIndex,
	)
	for paymentCount := 0; paymentCount < g.characteristicsForCard(card).MemoryCost; paymentCount++ {
		memoryCount := len(zones.Memory)
		randomValue := g.nextRandom()
		paymentIndex := int(randomValue % uint64(memoryCount))
		payment := zones.Memory[paymentIndex]
		zones.Memory = removeCardAt(zones.Memory, paymentIndex)
		zones.Banishment = append(zones.Banishment, payment)
		g.recordPublicEvent(player, "materialize:payment", "materialize-banish-memory", payment)
	}
	g.state.Zones[player.UID] = zones
	g.state.EffectSources = append(g.state.EffectSources, card)
	g.state.EffectsStack = append(
		g.state.EffectsStack,
		effectStackItem{
			Kind:       effectStackMaterialization,
			Controller: player,
			Source:     candidate.ID,
		},
	)
	g.grantOpportunity(player)
	return nil
}

func cardIndex(cards []cardInstanceID, want cardInstanceID) int {
	for index, card := range cards {
		if card == want {
			return index
		}
	}
	return -1
}

// removeCardAt 保留其餘牌的順序，並可能改寫原切片的底層陣列。
// 呼叫端須保證 index 在範圍內；找不到牌時不可直接傳入 cardIndex 回傳的 -1。
func removeCardAt(cards []cardInstanceID, index int) []cardInstanceID {
	return append(
		cards[:index],
		cards[index+1:]...,
	)
}

// resolveMaterialization 消耗 Stack 來源，並依牌類型結算 Champion 升級或 Regalia 進場。
// 輸入為待結算 Stack 項目；無輸出；會移除來源並改變戰場或放逐區。
func (g *Game) resolveMaterialization(item effectStackItem) {
	sourceIndex := cardIndex(g.state.EffectSources, item.Source)
	if sourceIndex < 0 {
		return
	}
	g.state.EffectSources = removeCardAt(g.state.EffectSources, sourceIndex)
	candidate, exists := g.state.Cards[item.Source]
	if !exists {
		return
	}
	if containsString(candidate.Types, "REGALIA") {
		g.putMaterialRegaliaOnField(item.Controller, item.Source)
		return
	}
	g.resolveChampionLevelUp(item)
}

// resolveChampionLevelUp 重新檢查 Champion 升級條件，失效時將來源放逐且不退回已付費用。
// 輸入為來源已自 EffectSources 移除的 Stack 項目；無輸出；成功時替換 champion 並加入 Tonoris 入場觸發。
func (g *Game) resolveChampionLevelUp(item effectStackItem) {
	if !g.canResolveChampionLevelUp(item) {
		zones := g.state.Zones[item.Controller.UID]
		zones.Banishment = append(zones.Banishment, item.Source)
		g.state.Zones[item.Controller.UID] = zones
		return
	}
	champion := g.state.Champions[item.Controller.UID]
	previousTop := champion.Card
	sourceEntity := entityID(item.Source)
	champion.Card = item.Source
	champion.InnerLineage = append(champion.InnerLineage, previousTop)
	g.state.Champions[item.Controller.UID] = champion
	for _, player := range g.players {
		g.grantCardTracking(player, sourceEntity)
	}
	g.recordPublicEvent(
		item.Controller,
		"materialize:level-up",
		"level-up",
		item.Source,
	)
	g.state.EffectsStack = append(
		g.state.EffectsStack,
		effectStackItem{
			Kind:       effectStackTonorisTaunt,
			Controller: item.Controller,
			Source:     item.Source,
		},
	)
}

func (g *Game) canResolveChampionLevelUp(item effectStackItem) bool {
	candidate, exists := g.state.Cards[item.Source]
	if !exists || !samePlayer(candidate.Owner, item.Controller) || candidate.Definition != tonorisCardID {
		return false
	}
	champion, exists := g.state.Champions[item.Controller.UID]
	if !exists {
		return false
	}
	current, exists := g.state.Cards[champion.Card]
	if !exists || current.Definition != spiritOfFireCardID {
		return false
	}
	return candidate.Level == current.Level+1
}

func (g *Game) resolveTonorisTaunt(item effectStackItem) {
	champion := g.state.Champions[item.Controller.UID]
	playerCount := len(g.players)
	champion.TauntUntilTurn = g.state.Scheduler.TurnNumber + uint64(playerCount)
	g.state.Champions[item.Controller.UID] = champion
	g.recordPublicEvent(
		item.Controller,
		tonorisOnEnterCause,
		"taunt-granted",
		item.Source,
	)
}

func (g *Game) expireTimedChampionEffects() {
	g.expireContinuousEffects()
	g.expireReplacementEffects()
	for _, player := range g.players {
		champion := g.state.Champions[player.UID]
		if champion.TauntUntilTurn > 0 && champion.TauntUntilTurn <= g.state.Scheduler.TurnNumber {
			champion.TauntUntilTurn = 0
			g.state.Champions[player.UID] = champion
			g.recordPublicEvent(
				player,
				tonorisOnEnterCause,
				"taunt-expired",
				champion.Card,
			)
		}
	}
}

func (g *Game) recordPublicEvent(player *model.Player, cause, kind string, card cardInstanceID) {
	cardEntity := entityID(card)
	g.state.NextEvent++
	g.state.Events = append(
		g.state.Events,
		eventBatch{
			Player: player,
			Cause:  cause,
			Events: []gameEvent{
				{
					Sequence: g.state.NextEvent,
					Kind:     kind,
					Card:     card,
				},
			},
		},
	)
	for _, viewer := range g.players {
		g.recordVisibleEvent(viewer, kind, cardEntity)
	}
}
