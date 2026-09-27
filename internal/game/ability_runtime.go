package game

import (
	"fmt"

	"go-tcg/internal/model"
)

type abilityInstanceID uint64

type abilityInstance struct {
	ID          abilityInstanceID `json:"id"`
	Slot        AbilitySlotID     `json:"slot,omitempty"`
	Controller  *model.Player     `json:"controller"`
	Source      cardInstanceID    `json:"source"`
	SourceLKI   cardInstanceID    `json:"source_lki"`
	Target      objectID          `json:"target,omitempty"`
	RuntimeCopy bool              `json:"runtime_copy,omitempty"`
	Operations  []effectOperation `json:"operations"`
}

type effectOperationKind string

const (
	effectOperationChoose             effectOperationKind = "choose"
	effectOperationMove               effectOperationKind = "move"
	effectOperationDraw               effectOperationKind = "draw"
	effectOperationDrawToMemory       effectOperationKind = "draw_to_memory"
	effectOperationCounter            effectOperationKind = "counter"
	effectOperationDamage             effectOperationKind = "damage"
	effectOperationContinuousModifier effectOperationKind = "continuous_modifier"
	effectOperationChooseZoneCard     effectOperationKind = "choose_zone_card"
	effectOperationChooseMemoryAlly   effectOperationKind = "choose_memory_ally"
	effectOperationDiscard            effectOperationKind = "discard"
	effectOperationDeploy             effectOperationKind = "deploy"
	// effectOperationPutAllyOnField 以通用 ability runtime 結算被 play 的 Ally source。
	effectOperationPutAllyOnField        effectOperationKind = "put_ally_on_field"
	effectOperationSuitedThresholdDamage effectOperationKind = "suited_threshold_damage"
	effectOperationChooseDuchessCopy     effectOperationKind = "choose_duchess_copy"
	effectOperationCopyDuchessAction     effectOperationKind = "copy_duchess_action"
	effectOperationSacrificeForChef      effectOperationKind = "sacrifice_for_chef"
	effectOperationRetargetAttack        effectOperationKind = "retarget_attack"
)

// effectOperation 只保存可序列化的操作資料，讓能力與待選狀態可以重播。
// 操作可使用宣告時的目標，也可透過 choose 在結算期間取得目標；不保存執行閉包。
type effectOperation struct {
	Kind                      effectOperationKind `json:"kind"`
	Target                    objectID            `json:"target,omitempty"`
	Source                    objectID            `json:"source,omitempty"`
	Amount                    int                 `json:"amount,omitempty"`
	Counter                   string              `json:"counter,omitempty"`
	MoveSourceToGraveyard     bool                `json:"move_source_to_graveyard,omitempty"`
	DistinctSuitedCostsDamage bool                `json:"distinct_suited_costs_damage,omitempty"`
	CanPass                   bool                `json:"can_pass,omitempty"`
	CardZone                  cardZone            `json:"card_zone,omitempty"`
	ContinuousEffect          continuousEffect    `json:"continuous_effect,omitempty"`
	Options                   []objectID          `json:"options,omitempty"`
}

type abilityChoice struct {
	Instance   abilityInstance   `json:"instance"`
	Operations []effectOperation `json:"operations"`
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
		if target == "" {
			target = instance.Target
		}
		switch operation.Kind {
		case effectOperationDamage:
			amount := operation.Amount
			if operation.DistinctSuitedCostsDamage {
				amount += g.distinctSuitedPrintedReserveCosts(instance.Controller)
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
			if !g.isLegalTarget(target) {
				continue
			}
			effect := operation.ContinuousEffect
			effect.Controller = instance.Controller
			effect.Source = objectID(instance.SourceLKI)
			effect.Target = target
			g.addContinuousEffect(effect)
			g.recordPublicEvent(instance.Controller, "ability", "continuous-effect", instance.SourceLKI)
		case effectOperationDraw:
			g.drawCards(instance.Controller, operation.Amount)
		case effectOperationDrawToMemory:
			g.drawToMemory(instance.Controller, operation.Amount)
		case effectOperationCounter:
			if g.addCounter(target, operation.Counter, operation.Amount) {
				g.recordPublicEvent(instance.Controller, "ability", "counter", instance.SourceLKI)
			}
		case effectOperationChooseZoneCard:
			zone := operation.CardZone
			if zone == "" {
				zone = cardZoneHand
			}
			g.beginAbilityCardChoice(
				instance,
				operationIndex,
				instance.Controller,
				cardsInZone(g.state.Zones[instance.Controller.UID], zone),
				operation.CanPass,
			)
			if g.state.AbilityChoice != nil {
				g.advanceKnowledgeRevision()
			}
			return
		case effectOperationChooseMemoryAlly:
			g.beginAbilityCardChoice(
				instance,
				operationIndex,
				instance.Controller,
				g.qualifiedMemoryAllies(instance.Controller),
				operation.CanPass,
			)
			if g.state.AbilityChoice != nil {
				g.advanceKnowledgeRevision()
			}
			return
		case effectOperationChooseDuchessCopy:
			g.beginAbilityCardChoice(
				instance,
				operationIndex,
				instance.Controller,
				g.eligibleDuchessCopies(instance.Controller),
				operation.CanPass,
			)
			if g.state.AbilityChoice != nil {
				g.advanceKnowledgeRevision()
			}
			return
		case effectOperationCopyDuchessAction:
			copied, err := g.copyDuchessAction(
				instance.Controller,
				cardInstanceID(target),
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
		case effectOperationSacrificeForChef:
			if err := g.sacrificeForPepperedChef(
				instance.Controller,
				operation.Source,
				target,
			); err != nil {
				return
			}
		case effectOperationRetargetAttack:
			// 重導失敗時此牌照常結算後續移動；已支付的費用不會退回。
			_ = g.retargetAttackWithTrumpSet(instance.Controller, target)
		case effectOperationDiscard:
			g.discardCard(instance.Controller, operation.CardZone, cardInstanceID(target))
		case effectOperationDeploy:
			g.deployAlly(instance.Controller, cardInstanceID(target))
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
		case effectOperationMove:
			if operation.MoveSourceToGraveyard {
				g.removeEffectSource(instance.Source)
				g.putInGraveyard(instance.Source)
			}
		case effectOperationChoose:
			if len(operation.Options) == 0 {
				return
			}
			completed = false
			canPass := operation.CanPass || instance.RuntimeCopy
			g.state.AbilityChoice = &abilityChoice{
				Instance:   instance,
				Operations: append([]effectOperation(nil), instance.Operations[operationIndex+1:]...),
				CanPass:    canPass,
			}
			g.setDeclarationChoice(instance.Controller, operation.Options)
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
func (g *Game) beginAbilityCardChoice(instance abilityInstance, operationIndex int, player *model.Player, cards []cardInstanceID, canPass bool) {
	if len(cards) == 0 {
		return
	}
	options := make([]objectID, 0, len(cards))
	for _, card := range cards {
		options = append(options, objectID(card))
	}
	g.state.AbilityChoice = &abilityChoice{
		Instance:   instance,
		Operations: append([]effectOperation(nil), instance.Operations[operationIndex+1:]...),
		CanPass:    canPass,
	}
	g.setDeclarationChoice(player, options)
	g.state.Knowledge.Choice.CanPass = canPass
}

// discardCard 將指定區域的卡牌棄至墓地；來源未指定時依規則使用手牌。
// 輸入為玩家、來源區域與卡牌；卡牌不在該區域時不產生副作用，成功時記錄公開事件。
func (g *Game) discardCard(player *model.Player, from cardZone, card cardInstanceID) {
	if from == "" {
		from = cardZoneHand
	}
	zones := g.state.Zones[player.UID]
	updated, err := moveCardBetweenZones(zones, from, cardZoneGraveyard, card)
	if err != nil {
		return
	}
	g.state.Zones[player.UID] = updated
	g.recordPublicEvent(player, "ability", "discard", card)
}

// drawToMemory 逐張由主牌組頂抽至 Memory，並於抽空時停止結算及使玩家敗北。
// 輸入為玩家及張數；輸出為零，副作用為區域、公開事件、追蹤權及可能的終局狀態變更。
func (g *Game) drawToMemory(player *model.Player, amount int) {
	for range amount {
		card, drawn := g.drawOneCard(player, cardZoneMemory)
		if !drawn {
			return
		}
		g.recordDrawEvent(player, "draw-to-memory", card)
	}
}

// drawCards 逐張由主牌組頂抽至手牌，並於抽空時停止結算及使玩家敗北。
// 輸入為玩家及張數；輸出為零，副作用為區域、公開事件、追蹤權及可能的終局狀態變更。
func (g *Game) drawCards(player *model.Player, amount int) {
	for range amount {
		card, drawn := g.drawOneCard(player, cardZoneHand)
		if !drawn {
			return
		}
		g.recordDrawEvent(player, "draw", card)
	}
}

// recordDrawEvent 保存能力抽牌事件，僅讓抽牌玩家看見私有區域的卡名。
// 輸入為玩家、抽牌種類與卡牌；輸出為零，副作用為新增 Game Event 與該玩家的可見事件。
func (g *Game) recordDrawEvent(player *model.Player, kind string, card cardInstanceID) {
	g.state.NextEvent++
	g.state.Events = append(g.state.Events, eventBatch{
		Player: player,
		Cause:  "ability",
		Events: []gameEvent{
			{
				Sequence: g.state.NextEvent,
				Kind:     kind,
				Card:     card,
			},
		},
	})
	cardEntity := entityID(card)
	g.recordVisibleEvent(player, kind, cardEntity)
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
