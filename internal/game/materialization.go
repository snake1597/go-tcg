package game

import (
	"fmt"
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

type effectStackItemKind string

const (
	effectStackMaterialization effectStackItemKind = "materialization"
	effectStackTonorisTaunt    effectStackItemKind = "tonoris_on_enter_taunt"
	effectStackCombat          effectStackItemKind = "combat"
	effectStackAbility         effectStackItemKind = "ability"
)

type effectStackItem struct {
	Kind       effectStackItemKind `json:"kind"`
	Controller *model.Player       `json:"controller"`
	Source     cardInstanceID      `json:"source"`
	Target     objectID            `json:"target,omitempty"`
	Attacker   objectID            `json:"attacker,omitempty"`
	SourceLKI  cardInstanceID      `json:"source_lki,omitempty"`
	Ability    *abilityInstance    `json:"ability,omitempty"`
}

func (g *Game) legalChampionMaterializations(player *model.Player) []cardInstanceID {
	zones := g.state.Zones[player.UID]
	materialDeckSize := len(zones.MaterialDeck)
	cards := make([]cardInstanceID, 0, materialDeckSize)
	for _, card := range zones.MaterialDeck {
		if g.canMaterializeChampion(
			player,
			card,
		) {
			cards = append(cards, card)
		}
	}
	return cards
}

func (g *Game) canMaterializeChampion(player *model.Player, card cardInstanceID) bool {
	scheduler := g.state.Scheduler
	if scheduler.Kind != schedulerStable || scheduler.Phase != PhaseMaterialize || !samePlayer(scheduler.TurnPlayer, player) || scheduler.OpportunityHolder != nil {
		return false
	}
	candidate, exists := g.state.Cards[card]
	if !exists || !samePlayer(candidate.Owner, player) || candidate.Definition != tonorisCardID {
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
	return len(g.state.Zones[player.UID].Memory) >= g.characteristicsForCard(card).MemoryCost
}

func (g *Game) materializeChampion(player *model.Player, card cardInstanceID) error {
	if !g.canMaterializeChampion(
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
	return g.payChampionMaterialization(player, card)
}

// payChampionMaterialization 重新檢查升級與付款條件，再移走物質牌並隨機放逐 Memory。
// 來源保留在 EffectSources，升級效果入堆疊後授予玩家行動機會；此時尚未替換 champion。
func (g *Game) payChampionMaterialization(player *model.Player, card cardInstanceID) error {
	if !g.canMaterializeChampion(
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

// resolveTopEffectStack 先移除堆疊頂項目，再依種類結算，因此結算中加入的新項目會留在堆疊。
// 呼叫端須保證堆疊非空，能力項目須有 Ability；未知種類會 panic。
func (g *Game) resolveTopEffectStack() {
	lastIndex := len(g.state.EffectsStack) - 1
	item := g.state.EffectsStack[lastIndex]
	g.state.EffectsStack = g.state.EffectsStack[:lastIndex]
	switch item.Kind {
	case effectStackMaterialization:
		g.resolveChampionLevelUp(item)
	case effectStackTonorisTaunt:
		g.resolveTonorisTaunt(item)
	case effectStackCombat:
		g.resolveCombat(item)
	case effectStackAbility:
		g.resolveAbility(*item.Ability)
	default:
		panic(fmt.Sprintf("unknown effect stack item %q", item.Kind))
	}
}

// resolveChampionLevelUp 消耗 EffectSources 中的來源並重新檢查升級條件。
// 來源已消失時不處理；條件失效時將來源放逐，不退回已付費用。
// 成功時保留原 champion 物件，把舊牌納入 InnerLineage，公開新牌並加入 Tonoris 入場觸發。
func (g *Game) resolveChampionLevelUp(item effectStackItem) {
	sourceIndex := cardIndex(g.state.EffectSources, item.Source)
	if sourceIndex < 0 {
		return
	}
	g.state.EffectSources = removeCardAt(g.state.EffectSources, sourceIndex)
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
