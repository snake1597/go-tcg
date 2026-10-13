package game

import (
	"fmt"

	"go-tcg/internal/model"
)

type abilityInstanceID uint64

type abilityInstance struct {
	ID          abilityInstanceID              `json:"id"`
	Slot        AbilitySlotID                  `json:"slot,omitempty"`
	Controller  *model.Player                  `json:"controller"`
	Source      cardInstanceID                 `json:"source"`
	SourceLKI   cardInstanceID                 `json:"source_lki"`
	Target      objectID                       `json:"target,omitempty"`
	RuntimeCopy bool                           `json:"runtime_copy,omitempty"`
	Bindings    map[resolutionBinding]entityID `json:"bindings,omitempty"`
	Operations  []effectOperation              `json:"operations"`
}

// resolutionBinding 是 Resolution Frame 中供後續操作引用的具名選擇結果。
type resolutionBinding string

type effectOperationKind string

const (
	effectOperationChoose                 effectOperationKind = "choose"
	effectOperationMove                   effectOperationKind = "move"
	effectOperationRestSource             effectOperationKind = "rest_source"
	effectOperationBanishSource           effectOperationKind = "banish_source"
	effectOperationCardistryDiscount      effectOperationKind = "cardistry_discount"
	effectOperationPreventNonCombatDamage effectOperationKind = "prevent_non_combat_damage"
	effectOperationDraw                   effectOperationKind = "draw"
	effectOperationDrawToMemory           effectOperationKind = "draw_to_memory"
	effectOperationCounter                effectOperationKind = "counter"
	effectOperationDamage                 effectOperationKind = "damage"
	effectOperationContinuousModifier     effectOperationKind = "continuous_modifier"
	effectOperationChooseZoneCard         effectOperationKind = "choose_zone_card"
	effectOperationDiscard                effectOperationKind = "discard"
	effectOperationDeploy                 effectOperationKind = "deploy"
	// effectOperationPutAllyOnField 以通用 ability runtime 結算被 play 的 Ally source。
	effectOperationPutAllyOnField         effectOperationKind = "put_ally_on_field"
	effectOperationRestGrantImmortality   effectOperationKind = "rest_grant_immortality"
	effectOperationSuitedThresholdCounter effectOperationKind = "suited_threshold_counter"
	effectOperationSuitedThresholdDamage  effectOperationKind = "suited_threshold_damage"
	effectOperationCopyAction             effectOperationKind = "copy_action"
	effectOperationSacrificeAllyPower     effectOperationKind = "sacrifice_ally_power"
	effectOperationRetargetAttack         effectOperationKind = "retarget_attack"
	effectOperationGrantTaunt             effectOperationKind = "grant_taunt"
)

// effectOperation 只保存可序列化的操作資料，讓能力與待選狀態可以重播。
// 操作可使用宣告時的目標，也可透過 choose 在結算期間取得目標；不保存執行閉包。
type effectOperation struct {
	Kind                    effectOperationKind `json:"kind"`
	Target                  objectID            `json:"target,omitempty"`
	TargetReference         abilityReference    `json:"target_reference,omitempty"`
	Source                  objectID            `json:"source,omitempty"`
	Amount                  int                 `json:"amount,omitempty"`
	Counter                 string              `json:"counter,omitempty"`
	MoveSourceToGraveyard   bool                `json:"move_source_to_graveyard,omitempty"`
	Value                   *valueExpression    `json:"value,omitempty"`
	CanPass                 bool                `json:"can_pass,omitempty"`
	CardZone                cardZone            `json:"card_zone,omitempty"`
	CardSelection           cardSelectionSpec   `json:"card_selection,omitempty"`
	Binding                 resolutionBinding   `json:"binding,omitempty"`
	Selector                selectorKind        `json:"selector,omitempty"`
	Duration                modifierDuration    `json:"duration,omitempty"`
	RequiresRetargetSuccess bool                `json:"requires_retarget_success,omitempty"`
	ContinuousEffect        continuousEffect    `json:"continuous_effect,omitempty"`
	Options                 []objectID          `json:"options,omitempty"`
}

// resolutionFrame 保存能力在等待玩家選擇時的完整可序列化續行資料。
// 它以具名 binding 保存所選 Card Instance，不保存 closure 或可變 definition pointer。
type resolutionFrame struct {
	Instance   abilityInstance   `json:"instance"`
	Operations []effectOperation `json:"operations"`
	Binding    resolutionBinding `json:"binding,omitempty"`
	Candidates []cardInstanceID  `json:"candidates,omitempty"`
	CardZone   cardZone          `json:"card_zone,omitempty"`
	CanPass    bool              `json:"can_pass,omitempty"`
}

func (g *Game) newAbilityInstance(controller *model.Player, source cardInstanceID, target objectID, operations []effectOperation) abilityInstance {
	g.state.NextAbility++
	return abilityInstance{
		ID:         abilityInstanceID(g.state.NextAbility),
		Controller: controller,
		Source:     source,
		SourceLKI:  source,
		Target:     target,
		Operations: operations,
	}
}

func (g *Game) pushAbility(instance abilityInstance) {
	g.state.EffectsStack = append(g.state.EffectsStack, effectStackItem{
		Kind:       effectStackAbility,
		Controller: instance.Controller,
		Source:     instance.Source,
		SourceLKI:  instance.SourceLKI,
		Target:     instance.Target,
		Ability:    &instance,
	})
}

// resolveAbility 按宣告時建立的 operation 序列結算，個別 operation 未指定目標時使用 instance.Target。
// 傷害與持續效果在結算時重新檢查目標；目標失效時略過該 operation，不退回費用。
// 選牌分支與 effectOperationChoose 都會中斷結算；有選項時保存後續操作供選擇提交後繼續。
// 只有 effectOperationChoose 設定 completed=false，讓 runtime copy 在等待期間保留。
// 其他返回路徑會由 defer 銷毀 runtime copy；未知 operation 會 panic。
func (g *Game) resolveAbility(instance abilityInstance) {
	completed := true
	retargetSucceeded := true
	defer func() {
		if completed && instance.RuntimeCopy {
			g.destroyRuntimeCopy(instance.Source)
		}
	}()
	for operationIndex, operation := range instance.Operations {
		if g.state.Finished {
			return
		}
		target := operation.Target
		if operation.TargetReference == referenceDeclaredTarget {
			target = instance.Target
		} else if target == "" {
			target = instance.Target
		}
		switch operation.Kind {
		case effectOperationDamage:
			amount := operation.Amount
			if operation.Value != nil {
				amount = g.evaluateValue(*operation.Value, instance.Controller)
			}
			if g.isLegalTarget(target) {
				continuation := instance
				continuation.Operations = append(
					[]effectOperation(nil),
					instance.Operations[operationIndex+1:]...,
				)
				if !g.damageNonCombat(target, amount, instance.SourceLKI, &continuation) {
					completed = false
					return
				}
			}
		case effectOperationContinuousModifier:
			if operation.RequiresRetargetSuccess && !retargetSucceeded {
				continue
			}
			if !g.isLegalTarget(target) {
				continue
			}
			effect := operation.ContinuousEffect
			effect.Controller = instance.Controller
			effect.Source = objectID(instance.SourceLKI)
			effect.Target = target
			if operation.Duration != "" {
				expiresAtTurn, err := g.expiresAtTurnForModifierDuration(
					instance.Controller,
					operation.Duration,
				)
				if err != nil {
					return
				}
				effect.ExpiresAtTurn = expiresAtTurn
			}
			g.addContinuousEffect(effect)
			g.recordPublicEvent(instance.Controller, "ability", "continuous-effect", instance.SourceLKI)
		case effectOperationDraw:
			g.drawCardsWithDeckOut(
				instance.Controller,
				operation.Amount,
				cardZoneHand,
				"ability",
				"draw",
			)
		case effectOperationCardistryDiscount:
			g.state.CardistryDiscounts[instance.Controller.UID] = operation.Amount
		case effectOperationPreventNonCombatDamage:
			champion, exists := g.state.Champions[instance.Controller.UID]
			if !exists {
				continue
			}
			g.state.ReplacementEffects = append(
				g.state.ReplacementEffects,
				replacementEffect{
					Kind:          replacementDamagePrevent,
					Source:        objectID(instance.SourceLKI),
					Controller:    instance.Controller,
					Target:        champion.ID,
					ExpiresAtTurn: g.state.Scheduler.TurnNumber + 1,
				},
			)
		case effectOperationDrawToMemory:
			g.drawCardsWithDeckOut(
				instance.Controller,
				operation.Amount,
				cardZoneMemory,
				"ability",
				"draw-to-memory",
			)
		case effectOperationCounter:
			if g.addCounter(target, operation.Counter, operation.Amount) {
				g.recordPublicEvent(instance.Controller, "ability", "counter", instance.SourceLKI)
			}
		case effectOperationChooseZoneCard:
			selection := operation.CardSelection
			if selection.Zone == "" {
				selection.Zone = cardZoneHand
			}
			candidates := g.cardSelectionCandidates(
				instance.Controller,
				selection,
				"",
			)
			g.beginAbilityCardChoice(
				instance,
				operationIndex,
				instance.Controller,
				candidates,
				operation.CanPass,
				operation.Binding,
				selection.Zone,
			)
			if g.state.ResolutionFrame != nil {
				g.advanceKnowledgeRevision()
			}
			return
		case effectOperationCopyAction:
			copied, err := g.copyAction(
				instance.Controller,
				cardInstanceID(instance.Bindings[operation.Binding]),
				operation.CardSelection,
				"",
			)
			if err != nil {
				return
			}
			chooseTarget := effectOperation{
				Kind:    effectOperationChoose,
				Options: g.legalTargets(),
			}
			copied.Operations = append(
				[]effectOperation{
					chooseTarget,
				},
				copied.Operations...,
			)
			g.pushAbility(copied)
			return
		case effectOperationSacrificeAllyPower:
			if err := g.sacrificeAllyAndIncreaseSourcePower(
				instance.Controller,
				instance.Target,
				target,
			); err != nil {
				return
			}
		case effectOperationRetargetAttack:
			// 重導失敗時此牌照常結算後續移動；已支付的費用不會退回。
			retargetSucceeded = g.retargetAttack(instance.Controller, target) == nil
		case effectOperationDiscard:
			card := cardInstanceID(target)
			if operation.Binding != "" {
				card = cardInstanceID(instance.Bindings[operation.Binding])
			}
			if err := g.discardCard(instance.Controller, operation.CardZone, card); err != nil {
				return
			}
		case effectOperationDeploy:
			card := cardInstanceID(target)
			if operation.Binding != "" {
				card = cardInstanceID(instance.Bindings[operation.Binding])
			}
			g.deployAlly(instance.Controller, card)
		case effectOperationPutAllyOnField:
			sourceIndex := cardIndex(g.state.EffectSources, instance.Source)
			if sourceIndex < 0 {
				return
			}
			g.state.EffectSources = removeCardAt(g.state.EffectSources, sourceIndex)
			source, exists := g.state.Cards[instance.Source]
			if !exists || !samePlayer(source.Owner, instance.Controller) || !containsString(source.Types, "ALLY") {
				if exists {
					g.putInGraveyard(instance.Source)
				}
				return
			}
			g.putAllyOnField(instance.Controller, instance.Source)
		case effectOperationRestGrantImmortality:
			g.restAndGrantTemporaryImmortality(target)
		case effectOperationSuitedThresholdCounter:
			amount := suitedThresholdAmount(g.suitedReserveTotal(instance.Controller))
			if amount > 0 && g.addCounter(target, "BUFF", amount) {
				g.recordPublicEvent(instance.Controller, "ability", "counter", instance.SourceLKI)
			}
		case effectOperationSuitedThresholdDamage:
			amount := suitedThresholdAmount(g.suitedReserveTotal(instance.Controller)) * 2
			if amount > 0 && g.isLegalTarget(target) {
				continuation := instance
				continuation.Operations = append(
					[]effectOperation(nil),
					instance.Operations[operationIndex+1:]...,
				)
				if !g.damageNonCombat(target, amount, instance.SourceLKI, &continuation) {
					completed = false
					return
				}
			}
		case effectOperationGrantTaunt:
			g.grantTemporaryTaunt(instance.Controller, instance.Source)
		case effectOperationMove:
			if operation.MoveSourceToGraveyard {
				g.removeEffectSource(instance.Source)
				g.putInGraveyard(instance.Source)
			}
		case effectOperationChoose:
			options := operation.Options
			if operation.Selector == selectorControlledSuitedAlly {
				options = g.controlledSuitedAllies(instance.Controller)
			}
			if operation.Selector == selectorAllAllies {
				options = g.allAllies()
			}
			if operation.Selector == selectorUnits {
				options = g.legalTargets()
			}
			if operation.Selector == selectorOtherControlledAllies {
				options = g.otherControlledAllies(
					instance.Controller,
					operation.Source,
				)
			}
			if len(options) == 0 {
				return
			}
			completed = false
			canPass := operation.CanPass || instance.RuntimeCopy
			g.state.ResolutionFrame = &resolutionFrame{
				Instance:   instance,
				Operations: append([]effectOperation(nil), instance.Operations[operationIndex+1:]...),
				CanPass:    canPass,
			}
			g.setDeclarationChoice(instance.Controller, options)
			g.state.Knowledge.Choice.CanPass = canPass
			g.advanceKnowledgeRevision()
			return
		default:
			panic(fmt.Sprintf("unknown effect operation %q", operation.Kind))
		}
	}
}

func (g *Game) destroyRuntimeCopy(source cardInstanceID) {
	delete(g.state.Cards, source)
	delete(g.state.Entities, entityID(source))
}

// beginAbilityCardChoice 保存選牌後要繼續執行的 operation，並建立玩家專屬選項。
// 沒有可選牌時不建立選擇；呼叫此函式的結算分支仍會直接返回，不繼續後續操作。
func (g *Game) beginAbilityCardChoice(instance abilityInstance, operationIndex int, player *model.Player, cards []cardInstanceID, canPass bool, binding resolutionBinding, zone cardZone) {
	if len(cards) == 0 {
		return
	}
	options := make([]objectID, 0, len(cards))
	for _, card := range cards {
		options = append(options, objectID(card))
	}
	g.state.ResolutionFrame = &resolutionFrame{
		Instance:   instance,
		Operations: append([]effectOperation(nil), instance.Operations[operationIndex+1:]...),
		Binding:    binding,
		Candidates: append([]cardInstanceID(nil), cards...),
		CardZone:   zone,
		CanPass:    canPass,
	}
	g.setDeclarationChoice(player, options)
	g.state.Knowledge.Choice.CanPass = canPass
}

// discardCard 將指定區域的卡牌棄至墓地；來源未指定時依規則使用手牌。
// 輸入為玩家、來源區域與卡牌；輸出為錯誤或 nil；失敗無副作用，成功時移動區域、更新追蹤權並記錄公開事件。
func (g *Game) discardCard(player *model.Player, from cardZone, card cardInstanceID) error {
	if from == "" {
		from = cardZoneHand
	}
	err := g.commitPlayerCardMoves(
		playerCardMoveRequest{
			Player: player,
			Moves: []playerCardMove{
				{
					Card: card,
					From: from,
					To:   cardZoneGraveyard,
				},
			},
			Cause:        "ability",
			EventKind:    "discard",
			VisibleToAll: true,
		},
	)
	if err != nil {
		return fmt.Errorf("discard card: %w", err)
	}
	return nil
}

func (g *Game) addCounter(target objectID, counter string, amount int) bool {
	object, exists := g.state.Objects[target]
	if !exists {
		return false
	}
	if object.Counters == nil {
		object.Counters = make(map[string]int)
	}
	object.Counters[counter] += amount
	g.state.Objects[target] = object
	return true
}
