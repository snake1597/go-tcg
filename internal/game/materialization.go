package game

import (
	"fmt"
	"go-tcg/internal/constants"
	"go-tcg/internal/model"
	tcgErrors "go-tcg/internal/tcg_errors"
)

const (
	spiritOfFireCardID CardID = "LMyKyVC2O9"
	tonorisCardID      CardID = "zb14m4c8lj"
	impactHammerCardID CardID = "chsbalegbs"
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
	if !g.canPayMemoryCost(player, g.characteristicsForCard(card).MemoryCost) {
		return false
	}
	if containsString(candidate.Types, "REGALIA") {
		return true
	}
	return g.canMaterializeChampionLevelUp(player, candidate)
}

// canMaterializeChampionLevelUp 以 Champion 類型、compiled setup／enter slots 與等級判斷候選牌能否升級。
// 輸入為玩家與候選牌；輸出為是否符合通用升級條件；不會改變對局狀態。
func (g *Game) canMaterializeChampionLevelUp(player *model.Player, candidate cardInstance) bool {
	if !containsString(candidate.Types, "CHAMPION") {
		return false
	}
	champion, exists := g.state.Champions[player.UID]
	if !exists {
		return false
	}
	current, exists := g.state.Cards[champion.Card]
	if !exists || !containsString(current.Types, "CHAMPION") || len(g.compiledTriggeredAbilities(current.ID, eventKindSetup)) == 0 || len(g.compiledTriggeredAbilities(candidate.ID, eventKindEnter)) == 0 {
		return false
	}
	return candidate.Level == current.Level+1
}

// materialize 驗證玩家對 Material Deck 牌的宣告，並以通用 Memory 付款與 Stack 建立交由付款流程處理。
// 輸入為玩家、其公開的 materialization 牌與非隨機付款 handles；輸出為驗證或付款錯誤；成功時會改變區域與 Stack。
func (g *Game) materialize(player *model.Player, card cardInstanceID, memoryPayment []ViewHandle) error {
	return g.commitDeclarationTransaction(
		func() error {
			return g.materializeTransaction(
				player,
				card,
				memoryPayment,
			)
		},
	)
}

// materializeTransaction 在候選狀態中驗證 Material Deck 來源與 Memory Cost，再建立 materialization Stack Item。
// 輸入為玩家、其公開的 materialization 牌與非隨機付款 handles；輸出為驗證或付款錯誤；成功時只修改目前的候選權威狀態。
func (g *Game) materializeTransaction(player *model.Player, card cardInstanceID, memoryPayment []ViewHandle) error {
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
	if err := g.payMaterialization(player, card, memoryPayment); err != nil {
		return fmt.Errorf("pay materialization: %w", err)
	}
	return nil
}

// payMaterialization 重新檢查進場與付款條件，再以通用 Memory 付款移走 Material Deck 牌。
// 輸入為玩家、候選牌與非隨機付款 handles；來源保留在 EffectSources，成功時入堆疊並授予玩家行動機會。
func (g *Game) payMaterialization(player *model.Player, card cardInstanceID, memoryPayment []ViewHandle) error {
	if !g.canMaterialize(
		player,
		card,
	) {
		return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, card)
	}
	if err := g.payMemoryCost(player, g.characteristicsForCard(card).MemoryCost, memoryPayment, "materialize-banish-floating-memory", "materialize-banish-memory"); err != nil {
		return fmt.Errorf("pay memory cost: %w", err)
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
// 輸入為來源已自 EffectSources 移除的 Stack 項目；無輸出；成功時替換 champion 並加入已編譯的進場觸發。
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
	g.enqueueEnterAbilities(
		item.Controller,
		champion.ID,
		item.Source,
	)
}

func (g *Game) canResolveChampionLevelUp(item effectStackItem) bool {
	candidate, exists := g.state.Cards[item.Source]
	if !exists || !samePlayer(candidate.Owner, item.Controller) {
		return false
	}
	champion, exists := g.state.Champions[item.Controller.UID]
	if !exists {
		return false
	}
	current, exists := g.state.Cards[champion.Card]
	return exists && g.canMaterializeChampionLevelUp(item.Controller, candidate) && candidate.Level == current.Level+1
}

// grantTemporaryTaunt 授予控制者 Champion 到其下回合的 Taunt，並記錄來源 ability slot。
// 輸入為控制者與能力來源；無輸出；副作用為更新 Champion 與公開事件。
func (g *Game) grantTemporaryTaunt(player *model.Player, source cardInstanceID) {
	champion := g.state.Champions[player.UID]
	playerCount := len(g.players)
	champion.TauntUntilTurn = g.state.Scheduler.TurnNumber + uint64(playerCount)
	g.state.Champions[player.UID] = champion
	g.recordPublicEvent(
		player,
		"ability:champion:on-enter-taunt",
		"taunt-granted",
		source,
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
				"ability:champion:on-enter-taunt",
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
