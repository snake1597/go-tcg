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
	effectOperationChoose effectOperationKind = "choose"
	// effectOperationMoveSourceToGraveyard 以共通來源提交器將已結算的 Source Card 移入墓地。
	effectOperationMoveSourceToGraveyard  effectOperationKind = "move_source_to_graveyard"
	effectOperationRestSource             effectOperationKind = "rest_source"
	effectOperationBanishSource           effectOperationKind = "banish_source"
	effectOperationCardistryDiscount      effectOperationKind = "cardistry_discount"
	effectOperationPreventNonCombatDamage effectOperationKind = "prevent_non_combat_damage"
	effectOperationDraw                   effectOperationKind = "draw"
	effectOperationDrawToMemory           effectOperationKind = "draw_to_memory"
	effectOperationCounter                effectOperationKind = "counter"
	effectOperationDamage                 effectOperationKind = "damage"
	effectOperationContinuousModifier     effectOperationKind = "continuous_modifier"
	effectOperationDiscard                effectOperationKind = "discard"
	effectOperationDeploy                 effectOperationKind = "deploy"
	// effectOperationPutAllyOnField 以通用 ability runtime 結算被 play 的 Ally source。
	effectOperationPutAllyOnField       effectOperationKind = "put_ally_on_field"
	effectOperationRestGrantImmortality effectOperationKind = "rest_grant_immortality"
	effectOperationCopyAction           effectOperationKind = "copy_action"
	effectOperationSacrifice            effectOperationKind = "sacrifice"
	effectOperationRetargetAttack       effectOperationKind = "retarget_attack"
	effectOperationGrantTaunt           effectOperationKind = "grant_taunt"
)

// effectOperation 只保存可序列化的操作資料，讓能力與待選狀態可以重播。
// 操作可使用宣告時的目標，也可透過 choose 在結算期間取得目標；不保存執行閉包。
type effectOperation struct {
	Kind                    effectOperationKind `json:"kind"`
	Target                  objectID            `json:"target,omitempty"`
	TargetReference         abilityReference    `json:"target_reference,omitempty"`
	TargetBinding           resolutionBinding   `json:"target_binding,omitempty"`
	Source                  objectID            `json:"source,omitempty"`
	Amount                  int                 `json:"amount,omitempty"`
	Counter                 string              `json:"counter,omitempty"`
	Value                   *valueExpression    `json:"value,omitempty"`
	CanPass                 bool                `json:"can_pass,omitempty"`
	CardZone                cardZone            `json:"card_zone,omitempty"`
	CardSelection           cardSelectionSpec   `json:"card_selection,omitempty"`
	Binding                 resolutionBinding   `json:"binding,omitempty"`
	RequiredType            string              `json:"required_type,omitempty"`
	ExcludedObject          objectID            `json:"excluded_object,omitempty"`
	Selector                selectorKind        `json:"selector,omitempty"`
	Duration                modifierDuration    `json:"duration,omitempty"`
	RequiresRetargetSuccess bool                `json:"requires_retarget_success,omitempty"`
	ContinuousEffect        continuousEffect    `json:"continuous_effect,omitempty"`
	Options                 []objectID          `json:"options,omitempty"`
}

// resolutionFrame 保存能力在等待玩家選擇時的完整可序列化續行資料。
// 它以具名 binding 保存所選 Card Instance 或 Object，不保存 closure 或可變 definition pointer。
type resolutionFrame struct {
	Instance      abilityInstance   `json:"instance"`
	Operations    []effectOperation `json:"operations"`
	Binding       resolutionBinding `json:"binding,omitempty"`
	Candidates    []entityID        `json:"candidates,omitempty"`
	CardSelection cardSelectionSpec `json:"card_selection,omitempty"`
	Selector      selectorKind      `json:"selector,omitempty"`
	Options       []objectID        `json:"options,omitempty"`
	CanPass       bool              `json:"can_pass,omitempty"`
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
		if operation.TargetBinding != "" {
			bound, exists := instance.Bindings[operation.TargetBinding]
			if !exists {
				return
			}
			target = objectID(bound)
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
			amount := operation.Amount
			if operation.Value != nil {
				amount = g.evaluateValue(*operation.Value, instance.Controller)
			}
			if amount > 0 && g.addCounter(target, operation.Counter, amount) {
				g.recordPublicEvent(instance.Controller, "ability", "counter", instance.SourceLKI)
			}
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
		case effectOperationSacrifice:
			excludedObject := operation.ExcludedObject
			if excludedObject == "" {
				excludedObject = instance.Target
			}
			_, err := g.commitObjectSacrifice(
				objectSacrificeRequest{
					Controller:     instance.Controller,
					Object:         target,
					RequiredType:   operation.RequiredType,
					ExcludedObject: excludedObject,
					Cause:          "ability:sacrifice",
				},
			)
			if err != nil {
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
			g.resolveAllyEffectSource(instance.Controller, instance.Source)
		case effectOperationRestGrantImmortality:
			g.restAndGrantTemporaryImmortality(target)
		case effectOperationGrantTaunt:
			g.grantTemporaryTaunt(instance.Controller, instance.Source)
		case effectOperationMoveSourceToGraveyard:
			g.resolveEffectSourceToGraveyard(instance.Source)
		case effectOperationChoose:
			options := g.choiceCandidates(instance.Controller, operation)
			if len(options) == 0 {
				return
			}
			completed = false
			canPass := operation.CanPass || instance.RuntimeCopy
			g.beginAbilityChoice(instance, operationIndex, operation, options, canPass)
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

// choiceCandidates 依 operation 的明示選項、Card selector 或 Object selector 建立當前合法候選。
// 輸入為選擇玩家與可序列化 operation；輸出為可放入 PendingChoice 的 entity；副作用為零。
func (g *Game) choiceCandidates(player *model.Player, operation effectOperation) []entityID {
	if operation.CardSelection.Zone != "" {
		cards := g.cardSelectionCandidates(player, operation.CardSelection, "")
		candidates := make([]entityID, 0, len(cards))
		for _, card := range cards {
			candidates = append(candidates, entityID(card))
		}
		return candidates
	}
	options := operation.Options
	if operation.Selector == "" {
		options = filterLegalTargets(g, options)
	}
	if operation.Selector == selectorControlledSuitedAlly {
		options = g.controlledSuitedAllies(player)
	}
	if operation.Selector == selectorAllAllies {
		options = g.allAllies()
	}
	if operation.Selector == selectorUnits {
		options = g.legalTargets()
	}
	if operation.Selector == selectorOtherControlledAllies {
		options = g.otherControlledAllies(player, operation.Source)
	}
	candidates := make([]entityID, 0, len(options))
	for _, option := range options {
		candidates = append(candidates, entityID(option))
	}
	return candidates
}

// filterLegalTargets 排除已離場或不再是合法 Unit 的明示 Object 選項。
// 輸入為遊戲與原始 Object ID；輸出為目前合法的 Object ID；副作用為零。
func filterLegalTargets(g *Game, options []objectID) []objectID {
	legal := make([]objectID, 0, len(options))
	for _, option := range options {
		if g.isLegalTarget(option) {
			legal = append(legal, option)
		}
	}
	return legal
}

// beginAbilityChoice 保存 Card 或 Object 選擇的續行資料，並建立指定玩家專屬的 PendingChoice。
// 輸入為能力、目前操作及候選 entity；輸出為零；副作用為寫入 Resolution Frame、Knowledge Choice 與可略過設定。
func (g *Game) beginAbilityChoice(instance abilityInstance, operationIndex int, operation effectOperation, candidates []entityID, canPass bool) {
	g.state.ResolutionFrame = &resolutionFrame{
		Instance:      instance,
		Operations:    append([]effectOperation(nil), instance.Operations[operationIndex+1:]...),
		Binding:       operation.Binding,
		Candidates:    append([]entityID(nil), candidates...),
		CardSelection: operation.CardSelection,
		Selector:      operation.Selector,
		Options:       append([]objectID(nil), operation.Options...),
		CanPass:       canPass,
	}
	options := make([]objectID, 0, len(candidates))
	for _, candidate := range candidates {
		options = append(options, objectID(candidate))
	}
	g.setDeclarationChoice(instance.Controller, options)
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
