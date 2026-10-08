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

func (g *Game) beginObjectAbility(player *model.Player, source objectID) error {
	object, exists := g.state.Objects[source]
	if !exists || !samePlayer(object.Owner, player) || !g.canActivateObjectAbility(player, source) {
		return fmt.Errorf("invalid object ability")
	}
	if abilityDefinition, supported := g.compiledObject(object.Card); supported {
		operations := abilityDefinition.operations()
		resolutionStart, err := g.payObjectAbilityCosts(player, source, operations)
		if err != nil {
			return fmt.Errorf("invalid Object ability definition")
		}
		ability := g.newAbilityInstance(
			player,
			object.Card,
			"",
			operations[resolutionStart:],
		)
		ability.Slot = abilityDefinition.slot
		g.pushAbility(ability)
		g.grantOpportunity(player)
		return nil
	}
	return fmt.Errorf("unsupported object ability")
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
	if ability, exists := g.compiledObject(object.Card); exists {
		operations := ability.operations()
		if len(operations) > 0 && operations[0].Kind == effectOperationRestSource && object.Rested {
			return false
		}
		return g.objectAbilityCanBeDeclared(player, ability)
	}
	return false
}

// objectAbilityCanBeDeclared 判定已編譯 Object ability 的公開條件與必要選擇是否同時可在宣告時滿足。
// 輸入為啟動者與已驗證的能力定義；輸出為當下是否可合法宣告，副作用為零。
func (g *Game) objectAbilityCanBeDeclared(player *model.Player, ability compiledAbilityDefinition) bool {
	if ability.objectCondition != nil && !g.opponentControlsElementChampion(
		player,
		ability.objectCondition.opponentChampionElement,
	) {
		return false
	}
	for _, effect := range ability.effects {
		if effect.kind == effectKindChooseObject && effect.chooseObject.selector == selectorAllAllies && len(g.allAllies()) == 0 {
			return false
		}
	}
	return true
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

// payObjectAbilityCosts 依定義順序原子支付 Object 的可選 rest 與必要 banish 費用。
// 輸入為控制者、場上來源與完整 operations；輸出為第一個結算 operation 索引或錯誤，成功時會休息並放逐來源。
func (g *Game) payObjectAbilityCosts(player *model.Player, source objectID, operations []effectOperation) (int, error) {
	object, exists := g.state.Objects[source]
	if !exists || !samePlayer(object.Owner, player) {
		return 0, fmt.Errorf("invalid object ability")
	}
	operationIndex := 0
	if len(operations) > 0 && operations[0].Kind == effectOperationRestSource {
		if object.Rested {
			return 0, fmt.Errorf("object source is already rested")
		}
		object.Rested = true
		g.state.Objects[source] = object
		operationIndex++
	}
	if operationIndex >= len(operations)-1 || operations[operationIndex].Kind != effectOperationBanishSource {
		return 0, fmt.Errorf("object source banish cost is missing")
	}
	if _, err := g.banishObjectForAbility(player, source); err != nil {
		return 0, err
	}
	return operationIndex + 1, nil
}

// cardistryObserverTriggers 由場上來源的 triggered definitions 建立 Ally Cardistry 觀察者能力。
// 輸入為啟動玩家與 Cardistry Ally；輸出為依物件 ID 排序的 stack items，副作用為零。
func (g *Game) cardistryObserverTriggers(player *model.Player, source objectID) []effectStackItem {
	object, exists := g.state.Objects[source]
	if !exists || !containsString(g.state.Cards[object.Card].Types, "ALLY") {
		return nil
	}
	triggers := []effectStackItem{}
	observerIDs := make([]objectID, 0, len(g.state.Objects))
	for id := range g.state.Objects {
		observerIDs = append(observerIDs, id)
	}
	sort.Slice(observerIDs, func(first, second int) bool {
		return observerIDs[first] < observerIDs[second]
	})
	for _, observerID := range observerIDs {
		observer := g.state.Objects[observerID]
		if !samePlayer(observer.Owner, player) {
			continue
		}
		definition, definitionExists := g.definitions[g.state.Cards[observer.Card].Definition]
		if !definitionExists {
			continue
		}
		for _, authored := range definition.abilities {
			if authored.kind != abilityKindTriggered || authored.trigger == nil || authored.trigger.event != eventKindCardistryActivated {
				continue
			}
			ability := g.newAbilityInstance(
				player,
				observer.Card,
				source,
				authored.operations(),
			)
			ability.Slot = authored.slot
			triggers = append(triggers, effectStackItem{
				Kind:       effectStackAbility,
				Controller: player,
				Source:     observer.Card,
				SourceLKI:  observer.Card,
				Target:     source,
				Ability:    &ability,
			})
		}
	}
	return triggers
}

// copyAction 驗證指定區域中的合格 Action，放逐原牌並建立獨立的 runtime copy 與能力。
// 複製能力不支付原 action 費用，並移除最後的來源進墓地操作；原牌保留在放逐區。
// runtime copy 完成結算或被略過時由能力流程銷毀，不作為一般牌移入墓地。
func (g *Game) copyAction(player *model.Player, source cardInstanceID, selection cardSelectionSpec, target objectID) (abilityInstance, error) {
	if selection.Zone != cardZoneGraveyard || selection.RequiredType != "ACTION" || !containsCard(g.cardSelectionCandidates(player, selection, ""), source) {
		return abilityInstance{}, fmt.Errorf("invalid Action copy source %q", source)
	}
	zones := g.state.Zones[player.UID]
	zones.Graveyard = removeCardAt(zones.Graveyard, cardIndex(zones.Graveyard, source))
	zones.Banishment = append(zones.Banishment, source)
	g.state.Zones[player.UID] = zones
	g.recordPublicEvent(player, "ability", "banish", source)
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

// allAllies 回傳場上所有 Ally，依物件 ID 排序。
// 此查詢不改變遊戲狀態，並排除 Champion 與非 Ally 物件。
func (g *Game) allAllies() []objectID {
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

// retargetableAttackTargets 將目前 attack 的可重導目標限制為控制者的 Suited ally，且不得維持原目標。
// 此查詢同時使用攻擊目標的共用合法性規則，避免在 stealth、taunt 或 true sight 改變後提供過期選項。
func (g *Game) retargetableAttackTargets(player *model.Player) []objectID {
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

// retargetAttack 將 active attack 的目標改為 player 控制的合格 Suited Ally。
// 輸入為控制者與新目標；輸出為錯誤或 nil，成功時僅改變 combat stack target，修正由後續 typed effect 套用。
func (g *Game) retargetAttack(player *model.Player, target objectID) error {
	for index := len(g.state.EffectsStack) - 1; index >= 0; index-- {
		item := &g.state.EffectsStack[index]
		if item.Kind != effectStackCombat {
			continue
		}
		targets := g.retargetableAttackTargets(player)
		if !containsObject(targets, target) {
			return fmt.Errorf("invalid retarget attack target")
		}
		item.Target = target
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

// enqueueAllyDeath 依已編譯的死亡能力建立持續效果，讓卡牌不需透過 Card ID 分支結算。
// 輸入為死亡 Ally 的控制者與卡牌實例；輸出為零值，副作用為建立 Ability Instance 並刷新 trigger stack。
func (g *Game) enqueueAllyDeath(player *model.Player, card cardInstanceID) {
	abilityDefinition, exists := g.compiledAbility(card, abilityKindTriggered)
	if !exists || abilityDefinition.trigger == nil || abilityDefinition.trigger.event != eventKindDestroy || abilityDefinition.deathModifier == nil {
		return
	}
	expiresAtTurn, err := g.expiresAtTurnForModifierDuration(
		player,
		abilityDefinition.deathModifier.duration,
	)
	if err != nil {
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
				ExpiresAtTurn: expiresAtTurn,
				Modifier:      abilityDefinition.deathModifier.modifier,
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
	ability.Slot = abilityDefinition.slot
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

// expiresAtTurnForModifierDuration 將已編譯的持續時間轉成 scheduler 使用的到期回合。
// 輸入為效果控制者與受支援的持續時間；輸出為到期回合或驗證錯誤，副作用為零。
func (g *Game) expiresAtTurnForModifierDuration(player *model.Player, duration modifierDuration) (uint64, error) {
	switch duration {
	case modifierDurationEndOfTurn:
		return g.state.Scheduler.TurnNumber + 1, nil
	case modifierDurationEndOfNextTurn:
		return g.endOfNextTurn(player), nil
	default:
		return 0, fmt.Errorf("unsupported modifier duration %q", duration)
	}
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
