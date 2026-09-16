package game

import (
	"fmt"
	"sort"

	"go-tcg/internal/model"
)

const (
	duchessCardID                   CardID = "qzv380ujf5"
	heatedVengeanceCardID           CardID = "td460e8ig0"
	pepperedChefCardID              CardID = "lcy0lw1veb"
	redHareCardID                   CardID = "5du8f077ua"
	smokeBombsCardID                CardID = "ScGcOmkoQt"
	duchessThornesCardID            CardID = "bEXmm4rKOs"
	trumpSetCardID                  CardID = "w7g91ru45w"
	veritaCardID                    CardID = "4qc47amgpp"
	safeguardAmuletCardID           CardID = "yj2rJBREH8"
	infernalVesselCardID            CardID = "vgWgu1DUYv"
	grandCrusadersRingCardID        CardID = "2gv7DC0KID"
	viridianProtectiveTrinketCardID CardID = "s3572j3oda"
	waterResonanceBaubleCardID      CardID = "dSSRtNnPtw"
	windResonanceBaubleCardID       CardID = "bHGUNMFLg9"
)

type objectAbilityDeclaration struct {
	Controller *model.Player
	Source     objectID
}

func (g *Game) beginObjectAbility(player *model.Player, source objectID) error {
	object, exists := g.state.Objects[source]
	if !exists || !samePlayer(object.Owner, player) || !g.canActivateObjectAbility(player, source) {
		return fmt.Errorf("invalid object ability")
	}
	if g.state.Cards[object.Card].Definition == duchessThornesCardID {
		return g.activateDuchessThornes(player, source)
	}
	if g.state.Cards[object.Card].Definition == safeguardAmuletCardID {
		return g.activateSafeguardAmulet(player, source)
	}
	if g.state.Cards[object.Card].Definition == grandCrusadersRingCardID {
		return g.activateBanishDrawAbility(player, source)
	}
	if g.state.Cards[object.Card].Definition == waterResonanceBaubleCardID || g.state.Cards[object.Card].Definition == windResonanceBaubleCardID {
		return g.activateBanishDrawAbility(player, source)
	}
	if g.state.Cards[object.Card].Definition != smokeBombsCardID {
		return fmt.Errorf("unsupported object ability")
	}
	options := g.smokeBombsTargets()
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

// legalObjectAbilities 回傳目前玩家可啟動的場上物件能力，並以物件 ID 穩定排序。
// 輸入為玩家；輸出為合法能力來源，副作用為零。
func (g *Game) legalObjectAbilities(player *model.Player) []objectID {
	sources := make([]objectID, 0, len(g.state.Objects))
	for source := range g.state.Objects {
		if g.canActivateObjectAbility(player, source) {
			sources = append(sources, source)
		}
	}
	sort.Slice(sources, func(first, second int) bool {
		return sources[first] < sources[second]
	})
	return sources
}

// canActivateObjectAbility 驗證受控物件是否有已實作且當下可啟動的能力。
// 輸入為玩家與物件來源；輸出為合法性，副作用為零。
func (g *Game) canActivateObjectAbility(player *model.Player, source objectID) bool {
	object, exists := g.state.Objects[source]
	if !exists || !samePlayer(object.Owner, player) {
		return false
	}
	switch g.state.Cards[object.Card].Definition {
	case duchessThornesCardID:
		return !object.Rested
	case smokeBombsCardID:
		return len(g.smokeBombsTargets()) > 0
	case safeguardAmuletCardID, grandCrusadersRingCardID:
		return true
	case waterResonanceBaubleCardID:
		return g.opponentControlsElementChampion(player, "WATER")
	case windResonanceBaubleCardID:
		return g.opponentControlsElementChampion(player, "WIND")
	default:
		return false
	}
}

// opponentControlsElementChampion 檢查任一對手目前的 Champion 是否具有指定元素。
// 輸入為檢查者與元素名稱；輸出為條件是否成立，副作用為零。
func (g *Game) opponentControlsElementChampion(player *model.Player, element string) bool {
	for _, champion := range g.state.Champions {
		if samePlayer(champion.Owner, player) || !containsString(g.state.Cards[champion.Card].Elements, element) {
			continue
		}
		return true
	}
	return false
}

// activateBanishDrawAbility 原子地放逐能力來源並將抽一張牌的能力推入 Effects Stack。
// 輸入為控制者與可啟動的 Regalia 物件；成功時改變場上物件、放逐區、堆疊與公開事件，錯誤時不改變狀態。
func (g *Game) activateBanishDrawAbility(player *model.Player, source objectID) error {
	object, err := g.banishObjectForAbility(player, source)
	if err != nil {
		return err
	}
	g.pushAbility(g.newAbilityInstance(player, object.Card, "", []effectOperation{
		{
			Kind:   effectOperationDraw,
			Amount: 1,
		},
	}))
	g.grantOpportunity(player)
	return nil
}

// banishObjectForAbility 驗證並一次提交場上物件移除與其卡牌放逐。
// 輸入為控制者與物件來源；輸出為移除前物件或錯誤，成功時更新 Objects、Banishment 與公開事件。
func (g *Game) banishObjectForAbility(player *model.Player, source objectID) (fieldObject, error) {
	object, exists := g.state.Objects[source]
	if !exists || !samePlayer(object.Owner, player) {
		return fieldObject{}, fmt.Errorf("invalid object ability")
	}
	delete(g.state.Objects, source)
	zones := g.state.Zones[player.UID]
	zones.Banishment = append(zones.Banishment, object.Card)
	g.state.Zones[player.UID] = zones
	g.recordPublicEvent(player, "ability", "banish", object.Card)
	return object, nil
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

func (g *Game) activateSmokeBombs(player *model.Player, source, target objectID) error {
	object, exists := g.state.Objects[source]
	targetObject, targetExists := g.state.Objects[target]
	if !exists || !targetExists || !samePlayer(object.Owner, player) || g.state.Cards[object.Card].Definition != smokeBombsCardID || !containsObject(g.smokeBombsTargets(), target) || !containsString(targetObject.Types, "ALLY") {
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

// smokeBombsTargets 回傳場上全部可作為 Smoke Bombs 目標的 Ally，依物件 ID 排序。
// 此查詢不改變遊戲狀態，並排除 Champion 與非 Ally 物件。
func (g *Game) smokeBombsTargets() []objectID {
	targets := make([]objectID, 0, len(g.state.Objects))
	for id, object := range g.state.Objects {
		if containsString(object.Types, "ALLY") {
			targets = append(targets, id)
		}
	}
	sort.Slice(
		targets,
		func(first, second int) bool {
			return targets[first] < targets[second]
		},
	)
	return targets
}

// trumpSetTargets 將目前 attack 的可重導目標限制為控制者的 Suited ally，且不得維持原目標。
// 此查詢同時使用攻擊目標的共用合法性規則，避免在 stealth、taunt 或 true sight 改變後提供過期選項。
func (g *Game) trumpSetTargets(player *model.Player) []objectID {
	for _, item := range g.state.EffectsStack {
		if item.Kind != effectStackCombat {
			continue
		}
		attacker := item.Attacker
		if attacker == "" {
			champion, exists := g.state.Champions[item.Controller.UID]
			if !exists {
				return nil
			}
			attacker = champion.ID
		}
		targets := make([]objectID, 0)
		for _, target := range g.controlledSuitedAllies(player) {
			if target != item.Target && g.isLegalAttackTarget(item.Controller, attacker, target) {
				targets = append(targets, target)
			}
		}
		return targets
	}
	return nil
}

// retargetAttackWithTrumpSet 將 active attack 的目標改為 player 控制的合格 Suited Ally。
// 成功時為新目標加入到回合結束的 +3 power／+3 life；目標、攻擊或攻擊來源失效時回傳錯誤且不改變狀態。
func (g *Game) retargetAttackWithTrumpSet(player *model.Player, target objectID) error {
	for index := len(g.state.EffectsStack) - 1; index >= 0; index-- {
		item := &g.state.EffectsStack[index]
		if item.Kind != effectStackCombat {
			continue
		}
		if !containsObject(g.trumpSetTargets(player), target) {
			return fmt.Errorf("invalid Trump Set target")
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
	operations := make([]effectOperation, 0)
	for _, target := range g.controlledSuitedAllies(player) {
		operations = append(operations, effectOperation{
			Kind:   effectOperationContinuousModifier,
			Target: target,
			ContinuousEffect: continuousEffect{
				Scope:         effectScopeObject,
				Layer:         effectLayerModifier,
				PowerLife:     powerLifeModify,
				ExpiresAtTurn: g.endOfNextTurn(player),
				Modifier: continuousModifier{
					PowerDelta: 1,
				},
			},
		})
	}
	if len(operations) == 0 {
		return
	}
	ability := g.newAbilityInstance(
		player,
		card,
		"",
		operations,
	)
	g.flushTriggers([]effectStackItem{
		{
			Kind:       effectStackAbility,
			Controller: player,
			Source:     card,
			SourceLKI:  card,
			Ability:    &ability,
		},
	})
}

// endOfNextTurn 回傳指定玩家下個回合結束後的過期回合編號。
// 輸入為效果控制者；輸出供 evaluator 的 ExpiresAtTurn 使用，副作用為零。
func (g *Game) endOfNextTurn(player *model.Player) uint64 {
	turns := uint64(0)
	current := g.state.Scheduler.TurnPlayer
	for {
		turns++
		current = g.nextPlayer(current)
		if samePlayer(current, player) {
			return g.state.Scheduler.TurnNumber + turns
		}
	}
}
