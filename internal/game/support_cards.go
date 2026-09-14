package game

import (
	"fmt"
	"sort"

	"go-tcg/internal/model"
)

const (
	duchessCardID         CardID = "qzv380ujf5"
	heatedVengeanceCardID CardID = "td460e8ig0"
	pepperedChefCardID    CardID = "lcy0lw1veb"
	redHareCardID         CardID = "5du8f077ua"
	smokeBombsCardID      CardID = "ScGcOmkoQt"
	duchessThornesCardID  CardID = "bEXmm4rKOs"
	trumpSetCardID        CardID = "w7g91ru45w"
	veritaCardID          CardID = "4qc47amgpp"
)

type objectAbilityDeclaration struct {
	Controller *model.Player
	Source     objectID
}

func (g *Game) beginObjectAbility(player *model.Player, source objectID) error {
	object, exists := g.state.Objects[source]
	if !exists || !samePlayer(object.Owner, player) {
		return fmt.Errorf("invalid object ability")
	}
	if g.state.Cards[object.Card].Definition == duchessThornesCardID {
		return g.activateDuchessThornes(player, source)
	}
	if g.state.Cards[object.Card].Definition != smokeBombsCardID {
		return fmt.Errorf("unsupported object ability")
	}
	options := []objectID{}
	for id, target := range g.state.Objects {
		if containsString(target.Types, "ALLY") {
			options = append(options, id)
		}
	}
	if len(options) == 0 {
		return fmt.Errorf("Smoke Bombs has no ally target")
	}
	g.state.Knowledge.ObjectAbility = &objectAbilityDeclaration{
		Controller: player,
		Source:     source,
	}
	g.setDeclarationChoice(player, options)
	return nil
}

func (g *Game) submitObjectAbilityChoice(player *model.Player, target objectID) error {
	declaration := g.state.Knowledge.ObjectAbility
	if declaration == nil || !samePlayer(declaration.Controller, player) {
		return fmt.Errorf("invalid object ability choice")
	}
	g.state.Knowledge.ObjectAbility = nil
	g.state.Knowledge.Choice = nil
	return g.activateSmokeBombs(player, declaration.Source, target)
}

// activateDuchessThornes pays its printed rest-and-banish cost and records a
// one-shot Cardistry discount. The discount is consumed by the next successful
// Cardistry activation, rather than by an attempted declaration.
func (g *Game) activateDuchessThornes(player *model.Player, source objectID) error {
	object, exists := g.state.Objects[source]
	if !exists || !samePlayer(object.Owner, player) || object.Rested || g.state.Cards[object.Card].Definition != duchessThornesCardID {
		return fmt.Errorf("invalid Duchess's Thornes activation")
	}
	object.Rested = true
	g.state.Objects[source] = object
	delete(g.state.Objects, source)
	zones := g.state.Zones[player.UID]
	zones.Banishment = append(zones.Banishment, object.Card)
	g.state.Zones[player.UID] = zones
	g.state.CardistryDiscounts[player.UID] = 6
	g.recordPublicEvent(player, "ability", "banish", object.Card)
	return nil
}

func (g *Game) cardistryObserverTriggers(player *model.Player, source objectID) []effectStackItem {
	object, exists := g.state.Objects[source]
	if !exists || !containsString(g.state.Cards[object.Card].Types, "ALLY") {
		return nil
	}
	triggers := []effectStackItem{}
	for _, observer := range g.state.Objects {
		if !samePlayer(observer.Owner, player) || g.state.Cards[observer.Card].Definition != duchessThornesCardID {
			continue
		}
		ability := g.newAbilityInstance(player, observer.Card, source, []effectOperation{
			g.temporaryModifierOperation(1, 0),
			{
				Kind: effectOperationContinuousModifier,
				ContinuousEffect: continuousEffect{
					Scope:         effectScopeObject,
					Layer:         effectLayerAbility,
					ExpiresAtTurn: g.state.Scheduler.TurnNumber + 1,
					Modifier:      continuousModifier{GrantTrueSight: true},
				},
			},
		})
		triggers = append(triggers, effectStackItem{Kind: effectStackAbility, Controller: player, Source: observer.Card, SourceLKI: observer.Card, Target: source, Ability: &ability})
	}
	return triggers
}

func (g *Game) eligibleDuchessCopies(player *model.Player) []cardInstanceID {
	cards := []cardInstanceID{}
	for _, card := range g.state.Zones[player.UID].Graveyard {
		candidate := g.state.Cards[card]
		if samePlayer(candidate.Owner, player) && containsString(candidate.Types, "ACTION") && containsString(candidate.Elements, "FIRE") && candidate.ReserveCost <= 2 {
			cards = append(cards, card)
		}
	}
	sort.Slice(cards, func(first, second int) bool { return cards[first] < cards[second] })
	return cards
}

// copyDuchessAction 驗證墓地中的合格 Fire action，放逐原牌並建立獨立的 runtime copy 與能力。
// 複製能力不支付原 action 費用，並移除最後的來源進墓地操作；原牌保留在放逐區。
// runtime copy 完成結算或被略過時由能力流程銷毀，不作為一般牌移入墓地。
func (g *Game) copyDuchessAction(player *model.Player, source cardInstanceID, target objectID) (abilityInstance, error) {
	if cardIndex(g.state.Zones[player.UID].Graveyard, source) < 0 || !containsCard(g.eligibleDuchessCopies(player), source) {
		return abilityInstance{}, fmt.Errorf("invalid Duchess copy source %q", source)
	}
	zones := g.state.Zones[player.UID]
	zones.Graveyard = removeCardAt(zones.Graveyard, cardIndex(zones.Graveyard, source))
	zones.Banishment = append(zones.Banishment, source)
	g.state.Zones[player.UID] = zones
	copyID := cardInstanceID(fmt.Sprintf("copy:%s:%d", source, g.state.NextAbility+1))
	copyCard := g.state.Cards[source]
	copyCard.ID = copyID
	g.state.Cards[copyID] = copyCard
	g.state.Entities[entityID(copyID)] = g.state.Entities[entityID(source)]
	declaration := &actionDeclaration{Controller: player, Source: copyID, Target: target}
	instance := g.actionAbilityInstance(declaration)
	instance.Operations = instance.Operations[:len(instance.Operations)-1]
	instance.RuntimeCopy = true
	return instance, nil
}

func containsCard(cards []cardInstanceID, want cardInstanceID) bool {
	for _, card := range cards {
		if card == want {
			return true
		}
	}
	return false
}

func (g *Game) canUseVeritaAlternativeCost(player *model.Player, cards []cardInstanceID) bool {
	if len(cards) < 3 {
		return false
	}
	total := 0
	seen := make(map[cardInstanceID]bool, len(cards))
	for _, card := range cards {
		if seen[card] || cardIndex(g.state.Zones[player.UID].Graveyard, card) < 0 || !g.cardHasSubtype(card, "SUITED") || !containsString(g.state.Cards[card].Types, "ALLY") {
			return false
		}
		seen[card] = true
		total += g.printedReserveCost(card)
	}
	return total == 10
}

// veritaAlternativeCostCards 依墓地順序搜尋第一組至少三張、印刷 reserve cost 總和為 10 的 Suited ally。
// 沒有合格組合時回傳 nil；此函式只找付款組合，實際放逐由 payVeritaAlternativeCost 執行。
func (g *Game) veritaAlternativeCostCards(player *model.Player) []cardInstanceID {
	graveyard := g.state.Zones[player.UID].Graveyard
	return g.findVeritaAlternativeCostCards(player, graveyard, nil, 0)
}

func (g *Game) findVeritaAlternativeCostCards(player *model.Player, cards, selected []cardInstanceID, start int) []cardInstanceID {
	if len(selected) >= 3 && g.canUseVeritaAlternativeCost(player, selected) {
		return append([]cardInstanceID(nil), selected...)
	}
	for index := start; index < len(cards); index++ {
		candidate := append(selected, cards[index])
		if result := g.findVeritaAlternativeCostCards(player, cards, candidate, index+1); len(result) > 0 {
			return result
		}
	}
	return nil
}

func (g *Game) payVeritaAlternativeCost(player *model.Player, cards []cardInstanceID) error {
	if !g.canUseVeritaAlternativeCost(player, cards) {
		return fmt.Errorf("invalid Verita alternative cost")
	}
	zones := g.state.Zones[player.UID]
	for _, card := range cards {
		index := cardIndex(zones.Graveyard, card)
		zones.Graveyard = removeCardAt(zones.Graveyard, index)
		zones.Banishment = append(zones.Banishment, card)
	}
	g.state.Zones[player.UID] = zones
	return nil
}

func (g *Game) redHareObeys(player *model.Player, redHare objectID) bool {
	champion, exists := g.state.Champions[player.UID]
	if !exists || g.state.Cards[champion.Card].Level >= 3 {
		return true
	}
	for id, object := range g.state.Objects {
		card := g.state.Cards[object.Card]
		if id != redHare && samePlayer(object.Owner, player) && containsString(card.Types, "UNIQUE") && containsString(card.Types, "ALLY") && containsString(card.Subtypes, "HUMAN") && (containsString(card.Elements, "FIRE") || containsString(card.Elements, "TERA")) {
			return true
		}
	}
	return false
}

func (g *Game) activateSmokeBombs(player *model.Player, source, target objectID) error {
	object, exists := g.state.Objects[source]
	targetObject, targetExists := g.state.Objects[target]
	if !exists || !targetExists || !samePlayer(object.Owner, player) || g.state.Cards[object.Card].Definition != smokeBombsCardID || !containsString(targetObject.Types, "ALLY") {
		return fmt.Errorf("invalid Smoke Bombs activation")
	}
	delete(g.state.Objects, source)
	zones := g.state.Zones[player.UID]
	zones.Banishment = append(zones.Banishment, object.Card)
	g.state.Zones[player.UID] = zones
	g.addContinuousEffect(continuousEffect{Source: source, Controller: player, Target: target, Scope: effectScopeObject, Layer: effectLayerAbility, ExpiresAtTurn: g.state.Scheduler.TurnNumber + 1, Modifier: continuousModifier{GrantStealth: true}})
	g.drawCards(player, 1)
	return nil
}

func (g *Game) retargetAttackWithTrumpSet(player *model.Player, target objectID) error {
	if !g.isLegalTarget(target) || !samePlayer(g.state.Objects[target].Owner, player) || !g.cardHasSubtype(g.state.Objects[target].Card, "SUITED") {
		return fmt.Errorf("invalid Trump Set target")
	}
	for index := len(g.state.EffectsStack) - 1; index >= 0; index-- {
		item := &g.state.EffectsStack[index]
		if item.Kind != effectStackCombat || item.Target == target {
			continue
		}
		item.Target = target
		g.addContinuousEffect(continuousEffect{Controller: player, Target: target, Scope: effectScopeObject, Layer: effectLayerModifier, PowerLife: powerLifeModify, ExpiresAtTurn: g.state.Scheduler.TurnNumber + 1, Modifier: continuousModifier{PowerDelta: 3, LifeDelta: 3}})
		return nil
	}
	return fmt.Errorf("no active attack to retarget")
}

func (g *Game) sacrificeForPepperedChef(player *model.Player, chef, sacrifice objectID) error {
	chefObject, chefExists := g.state.Objects[chef]
	sacrificed, sacrificeExists := g.state.Objects[sacrifice]
	if !chefExists || !sacrificeExists || chef == sacrifice || !samePlayer(chefObject.Owner, player) || !samePlayer(sacrificed.Owner, player) || g.state.Cards[chefObject.Card].Definition != pepperedChefCardID || !containsString(sacrificed.Types, "ALLY") {
		return fmt.Errorf("invalid Peppered Chef sacrifice")
	}
	delete(g.state.Objects, sacrifice)
	zones := g.state.Zones[player.UID]
	zones.Graveyard = append(zones.Graveyard, sacrificed.Card)
	g.state.Zones[player.UID] = zones
	g.addContinuousEffect(continuousEffect{Source: chef, Controller: player, Target: chef, Scope: effectScopeObject, Layer: effectLayerModifier, PowerLife: powerLifeModify, ExpiresAtTurn: g.state.Scheduler.TurnNumber + 1, Modifier: continuousModifier{PowerDelta: 2}})
	return nil
}

func (g *Game) enqueueVeritaDeath(player *model.Player, card cardInstanceID) {
	if g.state.Cards[card].Definition != veritaCardID {
		return
	}
	for target, object := range g.state.Objects {
		if samePlayer(object.Owner, player) && containsString(object.Types, "ALLY") && g.cardHasSubtype(object.Card, "SUITED") {
			g.addContinuousEffect(continuousEffect{Controller: player, Target: target, Scope: effectScopeObject, Layer: effectLayerModifier, PowerLife: powerLifeModify, ExpiresAtTurn: g.state.Scheduler.TurnNumber + uint64(len(g.players)), Modifier: continuousModifier{PowerDelta: 1}})
		}
	}
}
