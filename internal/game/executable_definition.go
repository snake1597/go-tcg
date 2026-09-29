package game

import (
	"fmt"
	"regexp"
	"strings"
)

type abilityDefinitionKind string

const (
	abilityKindCardistry   abilityDefinitionKind = "cardistry"
	abilityKindAction      abilityDefinitionKind = "action"
	abilityKindTriggered   abilityDefinitionKind = "triggered"
	abilityKindStatic      abilityDefinitionKind = "static"
	abilityKindReplacement abilityDefinitionKind = "replacement"
	abilityKindAlly        abilityDefinitionKind = "ally"
)

type abilityTiming string

const (
	timingMainPhase abilityTiming = "main-phase"
	timingFast      abilityTiming = "fast"
)

type abilityUsage string

const usageOncePerObject abilityUsage = "once-per-object"

type abilityCostReduction string

const reductionDistinctSuitedCosts abilityCostReduction = "distinct-suited-reserve-costs"

type authoredEffectKind string

const (
	effectKindDraw                  authoredEffectKind = "draw"
	effectKindDrawToMemory          authoredEffectKind = "draw-to-memory"
	effectKindDamage                authoredEffectKind = "damage"
	effectKindChooseCard            authoredEffectKind = "choose-card"
	effectKindChooseObject          authoredEffectKind = "choose-object"
	effectKindDiscard               authoredEffectKind = "discard"
	effectKindDeploy                authoredEffectKind = "deploy"
	effectKindCounter               authoredEffectKind = "counter"
	effectKindModifier              authoredEffectKind = "modifier"
	effectKindCopyAction            authoredEffectKind = "copy-action"
	effectKindRetargetAttack        authoredEffectKind = "retarget-attack"
	effectKindMoveSourceToGraveyard authoredEffectKind = "move-source-to-graveyard"
)

type abilityReference string

const (
	referenceController     abilityReference = "controller"
	referenceDeclaredTarget abilityReference = "declared-target"
	referenceEventTarget    abilityReference = "event-target"
)

type eventKind string

const (
	eventKindWield   eventKind = "wield"
	eventKindDestroy eventKind = "destroy"
)

type selectorKind string

const (
	selectorUnits                            selectorKind = "units"
	selectorControlledSuited                 selectorKind = "controlled-suited"
	selectorControlledSuitedAlly             selectorKind = "controlled-suited-ally"
	selectorRetargetableControlledSuitedAlly selectorKind = "retargetable-controlled-suited-ally"
)

type targetSelector struct {
	kind selectorKind
}

type valueKind string

const (
	valueConstant             valueKind = "constant"
	valueAdd                  valueKind = "add"
	valueDistinctPrintedCosts valueKind = "distinct-printed-costs"
)

// valueExpression 是可序列化的整數運算樹；左右運算元僅用於加法。
type valueExpression struct {
	Kind     valueKind        `json:"kind"`
	Constant int              `json:"constant,omitempty"`
	Left     *valueExpression `json:"left,omitempty"`
	Right    *valueExpression `json:"right,omitempty"`
	Selector selectorKind     `json:"selector,omitempty"`
	Player   abilityReference `json:"player,omitempty"`
}

type drawEffectDefinition struct {
	amount    int
	recipient abilityReference
}

type chooseObjectEffectDefinition struct {
	selector selectorKind
}

type deployEffectDefinition struct {
	binding resolutionBinding
}

type counterEffectDefinition struct {
	counter string
	amount  int
}

type modifierEffectDefinition struct {
	duration                modifierDuration
	modifier                continuousModifier
	requiresRetargetSuccess bool
}

type damageEffectDefinition struct {
	target abilityReference
	amount valueExpression
}

type actionCostKind string

const actionCostSacrificeControlledWeapon actionCostKind = "sacrifice-controlled-weapon"

// actionCostDefinition 宣告 Action 在建立 Ability Instance 前必須原子支付的額外費用。
// 輸入為受支援的費用種類；輸出由 compiler 複製的不可變費用資料，副作用為零。
type actionCostDefinition struct {
	kind actionCostKind
}

type chooseCardEffectDefinition struct {
	selection cardSelectionSpec
	binding   resolutionBinding
}

type discardEffectDefinition struct {
	binding resolutionBinding
}

type copyActionEffectDefinition struct {
	binding resolutionBinding
}

type triggerDefinition struct {
	event eventKind
}

type staticPredicateKind string

const (
	staticPredicateSelf                       staticPredicateKind = "self"
	staticPredicateControlsQualifiedHumanAlly staticPredicateKind = "controls-qualified-human-ally"
	staticPredicateOtherControlledSuitedAlly  staticPredicateKind = "other-controlled-suited-ally"
)

type effectSubLayer string

const effectSubLayerNone effectSubLayer = "none"

type staticDuration string

const staticDurationWhilePredicate staticDuration = "while-predicate"

type sourcePresence string

const sourcePresenceRequired sourcePresence = "required"

type staticEffectDefinition struct {
	predicate      staticPredicateKind
	layer          effectLayer
	sublayer       effectSubLayer
	modifier       continuousModifier
	duration       staticDuration
	sourcePresence sourcePresence
}

type replacementEvent string

const replacementEventRecover replacementEvent = "recover"

type replacementTransformation string

const replacementTransformReduce replacementTransformation = "recover-reduce"

type replacementDefinition struct {
	event          replacementEvent
	transformation replacementTransformation
	amount         int
}

type modifierDuration string

const (
	modifierDurationEndOfTurn     modifierDuration = "end-of-turn"
	modifierDurationEndOfNextTurn modifierDuration = "end-of-next-turn"
)

type deathModifierDefinition struct {
	selector selectorKind
	duration modifierDuration
	modifier continuousModifier
}

type authoredEffectDefinition struct {
	kind         authoredEffectKind
	draw         *drawEffectDefinition
	damage       *damageEffectDefinition
	chooseCard   *chooseCardEffectDefinition
	chooseObject *chooseObjectEffectDefinition
	discard      *discardEffectDefinition
	deploy       *deployEffectDefinition
	counter      *counterEffectDefinition
	modifier     *modifierEffectDefinition
	copyAction   *copyActionEffectDefinition
}

type authoredAbilityDefinition struct {
	slot               AbilitySlotID
	kind               abilityDefinitionKind
	timing             abilityTiming
	usage              abilityUsage
	reduction          abilityCostReduction
	baseCost           int
	classCostReduction int
	target             *targetSelector
	trigger            *triggerDefinition
	static             *staticEffectDefinition
	replacement        *replacementDefinition
	alternative        *alternativeCostSpec
	deathModifier      *deathModifierDefinition
	cost               *actionCostDefinition
	effects            []authoredEffectDefinition
}

// compiledEffect 保存已驗證的具體效果 payload，不讓 authoring tag 進入對局狀態。
type compiledEffect struct {
	kind         authoredEffectKind
	draw         drawEffectDefinition
	damage       damageEffectDefinition
	chooseCard   chooseCardEffectDefinition
	chooseObject chooseObjectEffectDefinition
	discard      discardEffectDefinition
	deploy       deployEffectDefinition
	counter      counterEffectDefinition
	modifier     modifierEffectDefinition
	copyAction   copyActionEffectDefinition
}

// compiledAbilityDefinition 是建局前完成驗證的能力中介表示。
type compiledAbilityDefinition struct {
	slot               AbilitySlotID
	kind               abilityDefinitionKind
	timing             abilityTiming
	usage              abilityUsage
	reduction          abilityCostReduction
	baseCost           int
	classCostReduction int
	target             *targetSelector
	trigger            *triggerDefinition
	static             *staticEffectDefinition
	replacement        *replacementDefinition
	alternative        *alternativeCostSpec
	deathModifier      *deathModifierDefinition
	cost               *actionCostDefinition
	effects            []compiledEffect
}

var abilitySlotKeyPattern = regexp.MustCompile(`^[a-z][a-z0-9]*(?:-[a-z0-9]+)*$`)

// wonderlandsReignAbilities 集中宣告 Wonderland's Reign 的 Cardistry Slot、費用與抽牌效果。
// 輸入為零；輸出為獨立的 Go 編寫資料，副作用為零。
func wonderlandsReignAbilities() []authoredAbilityDefinition {
	return []authoredAbilityDefinition{
		{
			slot:      "ability:0mf1ug6yfi:front:cardistry-draw",
			kind:      abilityKindCardistry,
			timing:    timingMainPhase,
			usage:     usageOncePerObject,
			reduction: reductionDistinctSuitedCosts,
			baseCost:  10,
			effects: []authoredEffectDefinition{
				{
					kind: effectKindDraw,
					draw: &drawEffectDefinition{
						amount:    1,
						recipient: referenceController,
					},
				},
			},
		},
	}
}

// threeOfHeartsAbilities 集中宣告 Three of Hearts 的抽牌、具名選牌及棄牌 Cardistry Slot。
// 輸入為零；輸出為獨立的 Go 編寫資料，副作用為零。
func threeOfHeartsAbilities() []authoredAbilityDefinition {
	return []authoredAbilityDefinition{
		{
			slot:      "ability:1db8hz4prm:front:cardistry-draw-discard",
			kind:      abilityKindCardistry,
			timing:    timingMainPhase,
			usage:     usageOncePerObject,
			reduction: reductionDistinctSuitedCosts,
			baseCost:  3,
			effects: []authoredEffectDefinition{
				{
					kind: effectKindDraw,
					draw: &drawEffectDefinition{
						amount:    1,
						recipient: referenceController,
					},
				},
				{
					kind: effectKindChooseCard,
					chooseCard: &chooseCardEffectDefinition{
						selection: cardSelectionSpec{
							Zone: cardZoneHand,
						},
						binding: "discard-card",
					},
				},
				{
					kind: effectKindDiscard,
					discard: &discardEffectDefinition{
						binding: "discard-card",
					},
				},
			},
		},
	}
}

// duchessAbilities 宣告 Duchess 的墓地選牌與免費 Action 複製 Cardistry slot。
// 輸入為零；輸出為不共享可變狀態的 Go 編寫資料，副作用為零。
func duchessAbilities() []authoredAbilityDefinition {
	return []authoredAbilityDefinition{
		{
			slot:      "ability:qzv380ujf5:front:cardistry-copy-action",
			kind:      abilityKindCardistry,
			timing:    timingMainPhase,
			usage:     usageOncePerObject,
			reduction: reductionDistinctSuitedCosts,
			baseCost:  6,
			effects: []authoredEffectDefinition{
				{
					kind: effectKindChooseCard,
					chooseCard: &chooseCardEffectDefinition{
						selection: cardSelectionSpec{
							Zone:               cardZoneGraveyard,
							RequiredType:       "ACTION",
							RequiredElement:    "FIRE",
							MaximumReserveCost: 2,
						},
						binding: "copy-source",
					},
				},
				{
					kind: effectKindCopyAction,
					copyAction: &copyActionEffectDefinition{
						binding: "copy-source",
					},
				},
			},
		},
	}
}

// fiveOfSpadesAbilities 宣告 Five of Spades 的 Cardistry Slot 與來源自身的暫時 Power 修正。
// 輸入為零；輸出為不共享可變狀態的 Go 編寫資料，副作用為零。
func fiveOfSpadesAbilities() []authoredAbilityDefinition {
	return []authoredAbilityDefinition{
		{
			slot:      "ability:i9hf5lhl5f:front:cardistry-power",
			kind:      abilityKindCardistry,
			timing:    timingMainPhase,
			usage:     usageOncePerObject,
			reduction: reductionDistinctSuitedCosts,
			baseCost:  5,
			effects: []authoredEffectDefinition{
				{
					kind: effectKindModifier,
					modifier: &modifierEffectDefinition{
						duration: modifierDurationEndOfTurn,
						modifier: continuousModifier{
							PowerDelta: 5,
						},
					},
				},
			},
		},
	}
}

// fourOfSpadesAbilities 宣告 Four of Spades 的 Cardistry Slot 與抽至 Memory 效果。
// 輸入為零；輸出為不共享可變狀態的 Go 編寫資料，副作用為零。
func fourOfSpadesAbilities() []authoredAbilityDefinition {
	return []authoredAbilityDefinition{
		{
			slot:      "ability:8bolq2y5qp:front:cardistry-draw-to-memory",
			kind:      abilityKindCardistry,
			timing:    timingMainPhase,
			usage:     usageOncePerObject,
			reduction: reductionDistinctSuitedCosts,
			baseCost:  4,
			effects: []authoredEffectDefinition{
				{
					kind: effectKindDrawToMemory,
					draw: &drawEffectDefinition{
						amount:    1,
						recipient: referenceController,
					},
				},
			},
		},
	}
}

// fourOfHeartsAbilities 宣告 Four of Hearts 的抽至 Memory、選擇合格 Ally 與部署 Cardistry Slot。
// 輸入為零；輸出為不共享可變狀態的 Go 編寫資料，副作用為零。
func fourOfHeartsAbilities() []authoredAbilityDefinition {
	return []authoredAbilityDefinition{
		{
			slot:      "ability:xgax8bbjqj:front:cardistry-deploy",
			kind:      abilityKindCardistry,
			timing:    timingMainPhase,
			usage:     usageOncePerObject,
			reduction: reductionDistinctSuitedCosts,
			baseCost:  4,
			effects: []authoredEffectDefinition{
				{
					kind: effectKindDrawToMemory,
					draw: &drawEffectDefinition{
						amount:    1,
						recipient: referenceController,
					},
				},
				{
					kind: effectKindChooseCard,
					chooseCard: &chooseCardEffectDefinition{
						selection: cardSelectionSpec{
							Zone:               cardZoneMemory,
							RequiredType:       "ALLY",
							RequiredSubtype:    "SUITED",
							RequiredElements:   []string{"FIRE", "NORM"},
							MaximumReserveCost: 3,
						},
						binding: "deploy-card",
					},
				},
				{
					kind: effectKindDeploy,
					deploy: &deployEffectDefinition{
						binding: "deploy-card",
					},
				},
			},
		},
	}
}

// threeOfSpadesAbilities 宣告 Three of Spades 的選擇受控 Suited Ally 與暫時 Life 修正 Cardistry Slot。
// 輸入為零；輸出為不共享可變狀態的 Go 編寫資料，副作用為零。
func threeOfSpadesAbilities() []authoredAbilityDefinition {
	return []authoredAbilityDefinition{
		{
			slot:      "ability:o09csnorqv:front:cardistry-life",
			kind:      abilityKindCardistry,
			timing:    timingMainPhase,
			usage:     usageOncePerObject,
			reduction: reductionDistinctSuitedCosts,
			baseCost:  3,
			effects: []authoredEffectDefinition{
				{
					kind: effectKindChooseObject,
					chooseObject: &chooseObjectEffectDefinition{
						selector: selectorControlledSuitedAlly,
					},
				},
				{
					kind: effectKindModifier,
					modifier: &modifierEffectDefinition{
						duration: modifierDurationEndOfTurn,
						modifier: continuousModifier{
							LifeDelta: 2,
						},
					},
				},
			},
		},
	}
}

// twoOfHeartsAbilities 宣告 Two of Hearts 的 Cardistry Slot 與來源自身的暫時 Power 修正。
// 輸入為零；輸出為不共享可變狀態的 Go 編寫資料，副作用為零。
func twoOfHeartsAbilities() []authoredAbilityDefinition {
	return []authoredAbilityDefinition{
		{
			slot:      "ability:rufki4o41y:front:cardistry-power",
			kind:      abilityKindCardistry,
			timing:    timingMainPhase,
			usage:     usageOncePerObject,
			reduction: reductionDistinctSuitedCosts,
			baseCost:  2,
			effects: []authoredEffectDefinition{
				{
					kind: effectKindModifier,
					modifier: &modifierEffectDefinition{
						duration: modifierDurationEndOfTurn,
						modifier: continuousModifier{
							PowerDelta: 2,
						},
					},
				},
			},
		},
	}
}

// twoOfSpadesAbilities 宣告 Two of Spades 的 Fast Cardistry Slot 與自身 BUFF counter 效果。
// 輸入為零；輸出為不共享可變狀態的 Go 編寫資料，副作用為零。
func twoOfSpadesAbilities() []authoredAbilityDefinition {
	return []authoredAbilityDefinition{
		{
			slot:      "ability:e8ygl32jef:front:cardistry-buff-counter",
			kind:      abilityKindCardistry,
			timing:    timingFast,
			usage:     usageOncePerObject,
			reduction: reductionDistinctSuitedCosts,
			baseCost:  2,
			effects: []authoredEffectDefinition{
				{
					kind: effectKindCounter,
					counter: &counterEffectDefinition{
						counter: "BUFF",
						amount:  1,
					},
				},
			},
		},
	}
}

// straightFlareAbilities 宣告行動目標與結算時求值的傷害公式。
// 輸入為零；輸出為新的 Go 編寫資料，副作用為零。
func straightFlareAbilities() []authoredAbilityDefinition {
	return []authoredAbilityDefinition{
		{
			slot: "ability:28bjn8g50v:front:action-damage",
			kind: abilityKindAction,
			target: &targetSelector{
				kind: selectorUnits,
			},
			effects: []authoredEffectDefinition{
				{
					kind: effectKindDamage,
					damage: &damageEffectDefinition{
						target: referenceDeclaredTarget,
						amount: valueExpression{
							Kind: valueAdd,
							Left: &valueExpression{
								Kind:     valueConstant,
								Constant: 1,
							},
							Right: &valueExpression{
								Kind:     valueDistinctPrintedCosts,
								Selector: selectorControlledSuited,
								Player:   referenceController,
							},
						},
					},
				},
				{
					kind: effectKindMoveSourceToGraveyard,
				},
			},
		},
	}
}

// blazingThrowAbilities 宣告 Blazing Throw 的武器犧牲費用、單位目標、傷害與來源移動。
// 輸入為零；輸出為不可共用的 Go 編寫資料，副作用為零。
func blazingThrowAbilities() []authoredAbilityDefinition {
	return []authoredAbilityDefinition{
		{
			slot: "ability:iohZMWh5v5:front:action-sacrifice-weapon-damage",
			kind: abilityKindAction,
			cost: &actionCostDefinition{
				kind: actionCostSacrificeControlledWeapon,
			},
			target: &targetSelector{
				kind: selectorUnits,
			},
			effects: []authoredEffectDefinition{
				{
					kind: effectKindDamage,
					damage: &damageEffectDefinition{
						target: referenceDeclaredTarget,
						amount: valueExpression{
							Kind:     valueConstant,
							Constant: 4,
						},
					},
				},
				{
					kind: effectKindMoveSourceToGraveyard,
				},
			},
		},
	}
}

// fieryInterferenceAbilities 宣告 Fiery Interference 的傷害、禁止 recover 暫時修正與來源移動。
// 輸入為零；輸出為不可共用的 Go 編寫資料，副作用為零。
func fieryInterferenceAbilities() []authoredAbilityDefinition {
	return []authoredAbilityDefinition{
		{
			slot: "ability:gt2zqtgs42:front:action-damage-prohibit-recover",
			kind: abilityKindAction,
			target: &targetSelector{
				kind: selectorUnits,
			},
			effects: []authoredEffectDefinition{
				{
					kind: effectKindDamage,
					damage: &damageEffectDefinition{
						target: referenceDeclaredTarget,
						amount: valueExpression{
							Kind:     valueConstant,
							Constant: 2,
						},
					},
				},
				{
					kind: effectKindModifier,
					modifier: &modifierEffectDefinition{
						duration: modifierDurationEndOfTurn,
						modifier: continuousModifier{
							ProhibitRecover: true,
						},
					},
				},
				{
					kind: effectKindMoveSourceToGraveyard,
				},
			},
		},
	}
}

// trumpSetAbilities 宣告重導攻擊所需目標、暫時修正與來源移動。
// 輸入為零；輸出為不可共用的 Go 編寫資料，副作用為零。
func trumpSetAbilities() []authoredAbilityDefinition {
	return []authoredAbilityDefinition{
		{
			slot:               "ability:w7g91ru45w:front:action-retarget-attack",
			kind:               abilityKindAction,
			classCostReduction: 1,
			target: &targetSelector{
				kind: selectorRetargetableControlledSuitedAlly,
			},
			effects: []authoredEffectDefinition{
				{
					kind: effectKindRetargetAttack,
				},
				{
					kind: effectKindModifier,
					modifier: &modifierEffectDefinition{
						duration:                modifierDurationEndOfTurn,
						requiresRetargetSuccess: true,
						modifier: continuousModifier{
							PowerDelta: 3,
							LifeDelta:  3,
						},
					},
				},
				{
					kind: effectKindMoveSourceToGraveyard,
				},
			},
		},
	}
}

// impactHammerAbilities 宣告 Impact Hammer 觀察 wield event 後對該事件 unit 造成傷害的 triggered Slot。
// 輸入為零；輸出為新的 Go 編寫資料，副作用為零。
func impactHammerAbilities() []authoredAbilityDefinition {
	return []authoredAbilityDefinition{
		{
			slot: "ability:chsbalegbs:front:on-wield-self-damage",
			kind: abilityKindTriggered,
			trigger: &triggerDefinition{
				event: eventKindWield,
			},
			effects: []authoredEffectDefinition{
				{
					kind: effectKindDamage,
					damage: &damageEffectDefinition{
						target: referenceEventTarget,
						amount: valueExpression{
							Kind:     valueConstant,
							Constant: 3,
						},
					},
				},
			},
		},
	}
}

// redHareAbilities 宣告 Red Hare 的 Pride 與條件式靜態修正，讓中央 evaluator 套用攻擊限制與能力。
// 輸入為零；輸出為不共享可變狀態的 Go 編寫資料，副作用為零。
func redHareAbilities() []authoredAbilityDefinition {
	pride := 3
	return []authoredAbilityDefinition{
		{
			slot: "ability:5du8f077ua:front:pride",
			kind: abilityKindStatic,
			static: &staticEffectDefinition{
				predicate: staticPredicateSelf,
				layer:     effectLayerAbility,
				sublayer:  effectSubLayerNone,
				modifier: continuousModifier{
					SetPride: &pride,
				},
				duration:       staticDurationWhilePredicate,
				sourcePresence: sourcePresenceRequired,
			},
		},
		{
			slot: "ability:5du8f077ua:front:qualified-human",
			kind: abilityKindStatic,
			static: &staticEffectDefinition{
				predicate: staticPredicateControlsQualifiedHumanAlly,
				layer:     effectLayerAbility,
				sublayer:  effectSubLayerNone,
				modifier: continuousModifier{
					RemovePride:          true,
					GrantRedHareOnAttack: true,
				},
				duration:       staticDurationWhilePredicate,
				sourcePresence: sourcePresenceRequired,
			},
		},
	}
}

// infernalVesselAbilities 宣告 Infernal Vessel 對 recover event 的 replacement 與減少數量 transformation。
// 輸入為零；輸出為不可共用的 Go 編寫資料，副作用為零。
func infernalVesselAbilities() []authoredAbilityDefinition {
	return []authoredAbilityDefinition{
		{
			slot: "ability:vgWgu1DUYv:front:recover-reduce",
			kind: abilityKindReplacement,
			replacement: &replacementDefinition{
				event:          replacementEventRecover,
				transformation: replacementTransformReduce,
				amount:         3,
			},
		},
	}
}

// veritaAbilities 宣告 Verita 的一般 Ally 結算、墓地替代付款、靜態 Immortality 與死亡後暫時修正。
// 輸入為零；輸出為不共享可變狀態的 Go 編寫資料，副作用為零。
func veritaAbilities() []authoredAbilityDefinition {
	return []authoredAbilityDefinition{
		{
			slot: "ability:4qc47amgpp:front:ally-activation",
			kind: abilityKindAlly,
			alternative: &alternativeCostSpec{
				Selection: cardSelectionSpec{
					Zone:                     cardZoneGraveyard,
					RequiredType:             "ALLY",
					RequiredSubtype:          "SUITED",
					MinimumCards:             3,
					MatchPrintedReserveTotal: true,
					PrintedReserveTotal:      10,
				},
				Payment: alternativeCostBanishGraveyard,
			},
		},
		{
			slot: "ability:4qc47amgpp:front:other-suited-immortality",
			kind: abilityKindStatic,
			static: &staticEffectDefinition{
				predicate: staticPredicateOtherControlledSuitedAlly,
				layer:     effectLayerAbility,
				sublayer:  effectSubLayerNone,
				modifier: continuousModifier{
					GrantImmortality: true,
				},
				duration:       staticDurationWhilePredicate,
				sourcePresence: sourcePresenceRequired,
			},
		},
		{
			slot: "ability:4qc47amgpp:front:on-death-suited-power",
			kind: abilityKindTriggered,
			trigger: &triggerDefinition{
				event: eventKindDestroy,
			},
			deathModifier: &deathModifierDefinition{
				selector: selectorControlledSuited,
				duration: modifierDurationEndOfNextTurn,
				modifier: continuousModifier{
					PowerDelta: 1,
				},
			},
		},
	}
}

// authoredAbilitiesForCard 將固定卡牌的 Go 能力編寫資料交給同一個 compiler。
// 輸入為 Card ID；輸出為該卡目前已遷移的能力資料，副作用為零。
func authoredAbilitiesForCard(id CardID) []authoredAbilityDefinition {
	switch id {
	case wonderlandsReignCardID:
		return wonderlandsReignAbilities()
	case straightFlareCardID:
		return straightFlareAbilities()
	case blazingThrowCardID:
		return blazingThrowAbilities()
	case fieryInterferenceCardID:
		return fieryInterferenceAbilities()
	case trumpSetCardID:
		return trumpSetAbilities()
	case threeOfHeartsCardID:
		return threeOfHeartsAbilities()
	case duchessCardID:
		return duchessAbilities()
	case fiveOfSpadesCardID:
		return fiveOfSpadesAbilities()
	case fourOfSpadesCardID:
		return fourOfSpadesAbilities()
	case fourOfHeartsCardID:
		return fourOfHeartsAbilities()
	case threeOfSpadesCardID:
		return threeOfSpadesAbilities()
	case twoOfHeartsCardID:
		return twoOfHeartsAbilities()
	case twoOfSpadesCardID:
		return twoOfSpadesAbilities()
	case impactHammerCardID:
		return impactHammerAbilities()
	case redHareCardID:
		return redHareAbilities()
	case infernalVesselCardID:
		return infernalVesselAbilities()
	case veritaCardID:
		return veritaAbilities()
	}
	return nil
}

// compileAbilityDefinitions 驗證 Definition、Face、Slot、kind、payload 與 reference，產生 typed 中介表示。
// 輸入為卡牌定義和 Go 編寫資料；輸出為不可共用 authoring pointer 的能力資料或具完整脈絡的錯誤，副作用為零。
func compileAbilityDefinitions(definition CardDefinition, authored []authoredAbilityDefinition) ([]compiledAbilityDefinition, error) {
	compiled := make([]compiledAbilityDefinition, 0, len(authored))
	seen := make(map[AbilitySlotID]struct{}, len(authored))
	allowed := make(map[AbilitySlotID]struct{})
	registeredAbilities := authoredAbilitiesForCard(definition.id)
	for _, registered := range registeredAbilities {
		allowed[registered.slot] = struct{}{}
	}
	prefix := fmt.Sprintf("ability:%s:front:", definition.id)
	cardistryCount := 0
	for _, ability := range authored {
		context := fmt.Sprintf("definition %q face %q slot %q", definition.id, definition.face.id, ability.slot)
		slotKey, matchesFace := strings.CutPrefix(string(ability.slot), prefix)
		if !matchesFace || definition.face.id != CardFaceID("face:"+string(definition.id)+":front") || !abilitySlotKeyPattern.MatchString(slotKey) {
			return nil, fmt.Errorf("%s: slot has invalid definition or face relationship", context)
		}
		if _, exists := seen[ability.slot]; exists {
			return nil, fmt.Errorf("%s: duplicate slot", context)
		}
		seen[ability.slot] = struct{}{}
		if _, exists := allowed[ability.slot]; !exists {
			return nil, fmt.Errorf("%s: slot is not registered for definition", context)
		}
		switch ability.kind {
		case abilityKindCardistry:
			cardistryCount++
			if cardistryCount > 1 {
				return nil, fmt.Errorf("%s: kind permits only one Cardistry slot per face", context)
			}
			if ability.timing != timingMainPhase && ability.timing != timingFast {
				return nil, fmt.Errorf("%s: timing unsupported %q", context, ability.timing)
			}
			if ability.usage != usageOncePerObject {
				return nil, fmt.Errorf("%s: usage unsupported %q", context, ability.usage)
			}
			if ability.reduction != reductionDistinctSuitedCosts {
				return nil, fmt.Errorf("%s: cost_reduction unsupported %q", context, ability.reduction)
			}
			if ability.baseCost < 0 {
				return nil, fmt.Errorf("%s: base_cost must be nonnegative", context)
			}
			if ability.classCostReduction != 0 || ability.cost != nil || ability.target != nil {
				return nil, fmt.Errorf("%s: target unsupported for Cardistry", context)
			}
		case abilityKindAction:
			if ability.timing != "" || ability.usage != "" || ability.reduction != "" || ability.baseCost != 0 {
				return nil, fmt.Errorf("%s: action cost or timing payload unsupported", context)
			}
			if ability.target == nil || (ability.target.kind != selectorUnits && ability.target.kind != selectorRetargetableControlledSuitedAlly) {
				return nil, fmt.Errorf("%s: target selector unsupported", context)
			}
			if ability.trigger != nil {
				return nil, fmt.Errorf("%s: trigger unsupported for Action", context)
			}
			if ability.cost != nil && ability.cost.kind != actionCostSacrificeControlledWeapon {
				return nil, fmt.Errorf("%s: action cost unsupported", context)
			}
			if ability.classCostReduction < 0 || ability.classCostReduction > 1 {
				return nil, fmt.Errorf("%s: class cost reduction unsupported", context)
			}
		case abilityKindAlly:
			if ability.timing != "" || ability.usage != "" || ability.reduction != "" || ability.baseCost != 0 || ability.classCostReduction != 0 || ability.target != nil || ability.trigger != nil || ability.static != nil || ability.replacement != nil || ability.deathModifier != nil || ability.cost != nil || len(ability.effects) != 0 {
				return nil, fmt.Errorf("%s: Ally declaration payload unsupported", context)
			}
			if err := validateAlternativeCostSpec(ability.alternative); err != nil {
				return nil, fmt.Errorf("%s: alternative cost: %w", context, err)
			}
		case abilityKindTriggered:
			if ability.timing != "" || ability.usage != "" || ability.reduction != "" || ability.baseCost != 0 || ability.classCostReduction != 0 || ability.target != nil || ability.cost != nil {
				return nil, fmt.Errorf("%s: triggered declaration payload unsupported", context)
			}
			if ability.trigger == nil || (ability.trigger.event != eventKindWield && ability.trigger.event != eventKindDestroy) {
				return nil, fmt.Errorf("%s: trigger.event unsupported", context)
			}
			if ability.trigger.event == eventKindDestroy {
				if err := validateDeathModifierDefinition(ability.deathModifier); err != nil || len(ability.effects) != 0 {
					return nil, fmt.Errorf("%s: destroy trigger payload unsupported", context)
				}
			} else if ability.deathModifier != nil {
				return nil, fmt.Errorf("%s: wield trigger death modifier unsupported", context)
			}
		case abilityKindStatic:
			if ability.timing != "" || ability.usage != "" || ability.reduction != "" || ability.baseCost != 0 || ability.classCostReduction != 0 || ability.target != nil || ability.trigger != nil || ability.cost != nil || len(ability.effects) != 0 {
				return nil, fmt.Errorf("%s: static declaration payload unsupported", context)
			}
			if err := validateStaticEffectDefinition(ability.static); err != nil {
				return nil, fmt.Errorf("%s: static: %w", context, err)
			}
		case abilityKindReplacement:
			if ability.timing != "" || ability.usage != "" || ability.reduction != "" || ability.baseCost != 0 || ability.classCostReduction != 0 || ability.target != nil || ability.trigger != nil || ability.static != nil || ability.cost != nil || len(ability.effects) != 0 {
				return nil, fmt.Errorf("%s: replacement declaration payload unsupported", context)
			}
			if err := validateReplacementDefinition(ability.replacement); err != nil {
				return nil, fmt.Errorf("%s: replacement: %w", context, err)
			}
		default:
			return nil, fmt.Errorf("%s: unknown kind %q", context, ability.kind)
		}
		if ability.kind != abilityKindStatic && ability.kind != abilityKindReplacement && ability.kind != abilityKindAlly && !(ability.kind == abilityKindTriggered && ability.trigger != nil && ability.trigger.event == eventKindDestroy) && len(ability.effects) == 0 {
			return nil, fmt.Errorf("%s: effects must not be empty", context)
		}
		result := compiledAbilityDefinition{
			slot:               ability.slot,
			kind:               ability.kind,
			timing:             ability.timing,
			usage:              ability.usage,
			reduction:          ability.reduction,
			baseCost:           ability.baseCost,
			classCostReduction: ability.classCostReduction,
			effects:            make([]compiledEffect, 0, len(ability.effects)),
		}
		if ability.target != nil {
			selector := *ability.target
			result.target = &selector
		}
		if ability.trigger != nil {
			trigger := *ability.trigger
			result.trigger = &trigger
		}
		if ability.static != nil {
			static := cloneStaticEffectDefinition(*ability.static)
			result.static = &static
		}
		if ability.replacement != nil {
			replacement := *ability.replacement
			result.replacement = &replacement
		}
		if ability.alternative != nil {
			alternative := *ability.alternative
			result.alternative = &alternative
		}
		if ability.deathModifier != nil {
			deathModifier := *ability.deathModifier
			result.deathModifier = &deathModifier
		}
		if ability.cost != nil {
			cost := *ability.cost
			result.cost = &cost
		}
		bindings := make(map[resolutionBinding]cardSelectionSpec)
		for index, effect := range ability.effects {
			field := fmt.Sprintf("effects[%d]", index)
			switch effect.kind {
			case effectKindDraw:
				if ability.kind != abilityKindCardistry || effect.draw == nil || effect.damage != nil {
					return nil, fmt.Errorf("%s: %s.draw payload is required only for Cardistry", context, field)
				}
				if effect.draw.amount <= 0 {
					return nil, fmt.Errorf("%s: %s.draw.amount must be positive", context, field)
				}
				if effect.draw.recipient != referenceController {
					return nil, fmt.Errorf("%s: %s.draw.recipient unknown or wrong reference %q", context, field, effect.draw.recipient)
				}
				result.effects = append(result.effects, compiledEffect{
					kind: effectKindDraw,
					draw: *effect.draw,
				})
			case effectKindDrawToMemory:
				if ability.kind != abilityKindCardistry || effect.draw == nil || effect.damage != nil {
					return nil, fmt.Errorf("%s: %s.draw_to_memory payload is required only for Cardistry", context, field)
				}
				if effect.draw.amount <= 0 || effect.draw.recipient != referenceController {
					return nil, fmt.Errorf("%s: %s.draw_to_memory requires a positive controller draw", context, field)
				}
				result.effects = append(result.effects, compiledEffect{
					kind: effectKindDrawToMemory,
					draw: *effect.draw,
				})
			case effectKindDamage:
				if (ability.kind != abilityKindAction && ability.kind != abilityKindTriggered) || effect.damage == nil || effect.draw != nil {
					return nil, fmt.Errorf("%s: %s.damage payload is required only for Action or triggered", context, field)
				}
				expectedTarget := referenceDeclaredTarget
				if ability.kind == abilityKindTriggered {
					expectedTarget = referenceEventTarget
				}
				if effect.damage.target != expectedTarget {
					return nil, fmt.Errorf("%s: %s.damage.target has wrong reference %q", context, field, effect.damage.target)
				}
				if err := validateValueExpression(effect.damage.amount); err != nil {
					return nil, fmt.Errorf("%s: %s.damage.amount: %w", context, field, err)
				}
				result.effects = append(result.effects, compiledEffect{
					kind: effectKindDamage,
					damage: damageEffectDefinition{
						target: effect.damage.target,
						amount: cloneValueExpression(effect.damage.amount),
					},
				})
			case effectKindChooseCard:
				if ability.kind != abilityKindCardistry || effect.chooseCard == nil || effect.draw != nil || effect.damage != nil || effect.discard != nil || effect.copyAction != nil {
					return nil, fmt.Errorf("%s: %s.choose_card payload is required only for Cardistry", context, field)
				}
				if (effect.chooseCard.selection.Zone != cardZoneHand && effect.chooseCard.selection.Zone != cardZoneGraveyard && effect.chooseCard.selection.Zone != cardZoneMemory) || effect.chooseCard.binding == "" {
					return nil, fmt.Errorf("%s: %s.choose_card requires supported zone and binding", context, field)
				}
				if _, exists := bindings[effect.chooseCard.binding]; exists {
					return nil, fmt.Errorf("%s: %s.choose_card.binding %q is duplicated", context, field, effect.chooseCard.binding)
				}
				bindings[effect.chooseCard.binding] = effect.chooseCard.selection
				result.effects = append(result.effects, compiledEffect{
					kind:       effectKindChooseCard,
					chooseCard: *effect.chooseCard,
				})
			case effectKindChooseObject:
				if ability.kind != abilityKindCardistry || effect.chooseObject == nil || effect.chooseObject.selector != selectorControlledSuitedAlly {
					return nil, fmt.Errorf("%s: %s.choose_object requires a controlled Suited Ally selector", context, field)
				}
				result.effects = append(result.effects, compiledEffect{
					kind:         effectKindChooseObject,
					chooseObject: *effect.chooseObject,
				})
			case effectKindDiscard:
				if ability.kind != abilityKindCardistry || effect.discard == nil || effect.draw != nil || effect.damage != nil || effect.chooseCard != nil || effect.copyAction != nil {
					return nil, fmt.Errorf("%s: %s.discard payload is required only for Cardistry", context, field)
				}
				if effect.discard.binding == "" {
					return nil, fmt.Errorf("%s: %s.discard.binding must not be empty", context, field)
				}
				if _, exists := bindings[effect.discard.binding]; !exists {
					return nil, fmt.Errorf("%s: %s.discard.binding %q has no preceding choice", context, field, effect.discard.binding)
				}
				result.effects = append(result.effects, compiledEffect{
					kind:    effectKindDiscard,
					discard: *effect.discard,
				})
			case effectKindDeploy:
				if ability.kind != abilityKindCardistry || effect.deploy == nil || effect.deploy.binding == "" {
					return nil, fmt.Errorf("%s: %s.deploy.binding must reference a Memory Ally choice", context, field)
				}
				selection, exists := bindings[effect.deploy.binding]
				if !exists || selection.Zone != cardZoneMemory || selection.RequiredType != "ALLY" {
					return nil, fmt.Errorf("%s: %s.deploy.binding must reference a Memory Ally choice", context, field)
				}
				result.effects = append(result.effects, compiledEffect{
					kind:   effectKindDeploy,
					deploy: *effect.deploy,
				})
			case effectKindCounter:
				if ability.kind != abilityKindCardistry || effect.counter == nil || effect.counter.counter != "BUFF" || effect.counter.amount != 1 {
					return nil, fmt.Errorf("%s: %s.counter requires exactly one BUFF counter", context, field)
				}
				result.effects = append(result.effects, compiledEffect{
					kind:    effectKindCounter,
					counter: *effect.counter,
				})
			case effectKindModifier:
				if effect.modifier == nil || effect.modifier.duration != modifierDurationEndOfTurn || (ability.kind == abilityKindCardistry && (!isCardistryModifier(effect.modifier.modifier) || effect.modifier.requiresRetargetSuccess)) || (ability.kind == abilityKindAction && (!isActionModifier(effect.modifier.modifier) || (effect.modifier.requiresRetargetSuccess && (ability.target == nil || ability.target.kind != selectorRetargetableControlledSuitedAlly)))) || (ability.kind != abilityKindCardistry && ability.kind != abilityKindAction) {
					return nil, fmt.Errorf("%s: %s.modifier payload unsupported", context, field)
				}
				result.effects = append(result.effects, compiledEffect{
					kind:     effectKindModifier,
					modifier: *effect.modifier,
				})
			case effectKindCopyAction:
				if ability.kind != abilityKindCardistry || effect.copyAction == nil || effect.draw != nil || effect.damage != nil || effect.chooseCard != nil || effect.discard != nil {
					return nil, fmt.Errorf("%s: %s.copy_action payload is required only for Cardistry", context, field)
				}
				selection, exists := bindings[effect.copyAction.binding]
				if effect.copyAction.binding == "" || !exists || selection.Zone != cardZoneGraveyard || selection.RequiredType != "ACTION" {
					return nil, fmt.Errorf("%s: %s.copy_action.binding %q must reference a graveyard Action choice", context, field, effect.copyAction.binding)
				}
				result.effects = append(result.effects, compiledEffect{
					kind:       effectKindCopyAction,
					copyAction: *effect.copyAction,
					chooseCard: chooseCardEffectDefinition{
						selection: selection,
					},
				})
			case effectKindRetargetAttack:
				if ability.kind != abilityKindAction || ability.target == nil || ability.target.kind != selectorRetargetableControlledSuitedAlly {
					return nil, fmt.Errorf("%s: %s.retarget_attack requires a retargetable controlled Suited Ally target", context, field)
				}
				result.effects = append(result.effects, compiledEffect{
					kind: effectKindRetargetAttack,
				})
			case effectKindMoveSourceToGraveyard:
				if ability.kind != abilityKindAction || index != len(ability.effects)-1 {
					return nil, fmt.Errorf("%s: %s.move_source_to_graveyard must be the final Action effect", context, field)
				}
				result.effects = append(result.effects, compiledEffect{
					kind: effectKindMoveSourceToGraveyard,
				})
			default:
				return nil, fmt.Errorf("%s: %s.kind unknown %q", context, field, effect.kind)
			}
		}
		if ability.kind == abilityKindAction && (len(result.effects) == 0 || result.effects[len(result.effects)-1].kind != effectKindMoveSourceToGraveyard) {
			return nil, fmt.Errorf("%s: Action must end by moving its source to graveyard", context)
		}
		compiled = append(compiled, result)
	}
	for _, registered := range registeredAbilities {
		if _, exists := seen[registered.slot]; !exists {
			return nil, fmt.Errorf("definition %q face %q slot %q: slot is missing", definition.id, definition.face.id, registered.slot)
		}
	}
	return compiled, nil
}

// validateReplacementDefinition 限制 replacement event filter 與 transformation 為 Infernal Vessel 已證明的 recover 規則。
// 輸入為編寫的 replacement 定義；輸出為驗證錯誤或 nil，副作用為零。
func validateReplacementDefinition(definition *replacementDefinition) error {
	if definition == nil {
		return fmt.Errorf("definition is required")
	}
	if definition.event != replacementEventRecover {
		return fmt.Errorf("event unsupported %q", definition.event)
	}
	if definition.transformation != replacementTransformReduce {
		return fmt.Errorf("transformation unsupported %q", definition.transformation)
	}
	if definition.amount != 3 {
		return fmt.Errorf("amount must be 3")
	}
	return nil
}

// validateStaticEffectDefinition 限制目前支援集合可使用的 static predicate、layer、duration 與 modifier。
// 輸入為編寫的靜態效果；輸出為驗證錯誤或 nil，副作用為零。
func validateStaticEffectDefinition(effect *staticEffectDefinition) error {
	if effect == nil {
		return fmt.Errorf("definition is required")
	}
	if effect.layer != effectLayerAbility || effect.sublayer != effectSubLayerNone || effect.duration != staticDurationWhilePredicate || effect.sourcePresence != sourcePresenceRequired {
		return fmt.Errorf("layer, sublayer, duration, or source presence unsupported")
	}
	switch effect.predicate {
	case staticPredicateSelf:
		if !isPrideStaticModifier(effect.modifier) {
			return fmt.Errorf("self predicate requires nonnegative Pride modifier")
		}
	case staticPredicateControlsQualifiedHumanAlly:
		if !isQualifiedHumanStaticModifier(effect.modifier) {
			return fmt.Errorf("qualified Human predicate requires Pride removal and granted On Attack")
		}
	case staticPredicateOtherControlledSuitedAlly:
		if !isOtherControlledSuitedAllyModifier(effect.modifier) {
			return fmt.Errorf("other controlled Suited Ally predicate requires Immortality")
		}
	default:
		return fmt.Errorf("predicate unsupported %q", effect.predicate)
	}
	return nil
}

// validateAlternativeCostSpec 限制目前 Support Set 可用的墓地放逐替代費用。
// 輸入為編寫的替代費用；輸出為驗證錯誤或 nil，副作用為零。
func validateAlternativeCostSpec(spec *alternativeCostSpec) error {
	if spec == nil || spec.Selection.Zone != cardZoneGraveyard || spec.Selection.RequiredType != "ALLY" || spec.Selection.RequiredSubtype != "SUITED" || spec.Selection.MinimumCards != 3 || !spec.Selection.MatchPrintedReserveTotal || spec.Selection.PrintedReserveTotal != 10 || spec.Payment != alternativeCostBanishGraveyard {
		return fmt.Errorf("unsupported selection or payment")
	}
	return nil
}

// validateDeathModifierDefinition 限制死亡觸發以受控 Suited Ally 為目標並持續至擁有者下回合結束。
// 輸入為編寫的死亡後修正；輸出為驗證錯誤或 nil，副作用為零。
func validateDeathModifierDefinition(definition *deathModifierDefinition) error {
	if definition == nil || definition.selector != selectorControlledSuited || definition.duration != modifierDurationEndOfNextTurn || definition.modifier.PowerDelta != 1 {
		return fmt.Errorf("unsupported selector, duration, or modifier")
	}
	return nil
}

// isCardistryModifier 驗證 Cardistry 暫時修正只改變 Power 或 Life，且不混入其他 continuous modifier。
// 輸入為編寫的修正；輸出為是否屬於受支援的 Cardistry 修正，副作用為零。
func isCardistryModifier(modifier continuousModifier) bool {
	return modifier.SetPower == nil &&
		modifier.SetLife == nil &&
		(modifier.PowerDelta != 0 || modifier.LifeDelta != 0) &&
		modifier.ReserveCostDelta == 0 &&
		!modifier.GrantImmortality &&
		!modifier.ProhibitRecover &&
		!modifier.SwitchPowerLife &&
		!modifier.GrantStealth &&
		!modifier.GrantTrueSight &&
		modifier.SetPride == nil &&
		!modifier.RemovePride &&
		!modifier.GrantRedHareOnAttack
}

// isActionModifier 驗證 Action 的暫時修正僅為 Power/Life 或 recovery prohibition，且不混入其他機制。
// 輸入為編寫的 continuous modifier；輸出為是否符合已支援 Action payload，副作用為零。
func isActionModifier(modifier continuousModifier) bool {
	return isCardistryModifier(modifier) || (modifier.SetPower == nil &&
		modifier.SetLife == nil &&
		modifier.PowerDelta == 0 &&
		modifier.LifeDelta == 0 &&
		modifier.ReserveCostDelta == 0 &&
		!modifier.GrantImmortality &&
		modifier.ProhibitRecover &&
		!modifier.SwitchPowerLife &&
		!modifier.GrantStealth &&
		!modifier.GrantTrueSight &&
		modifier.SetPride == nil &&
		!modifier.RemovePride &&
		!modifier.GrantRedHareOnAttack)
}

// isPrideStaticModifier 驗證 Pride static 只設定非負 Pride，不混入未支援的特徵修正。
// 輸入為靜態能力的 modifier；輸出為是否符合 Pride 定義，副作用為零。
func isPrideStaticModifier(modifier continuousModifier) bool {
	return modifier.SetPride != nil &&
		*modifier.SetPride >= 0 &&
		modifier.SetPower == nil &&
		modifier.SetLife == nil &&
		modifier.PowerDelta == 0 &&
		modifier.LifeDelta == 0 &&
		modifier.ReserveCostDelta == 0 &&
		!modifier.GrantImmortality &&
		!modifier.ProhibitRecover &&
		!modifier.SwitchPowerLife &&
		!modifier.GrantStealth &&
		!modifier.GrantTrueSight &&
		!modifier.RemovePride &&
		!modifier.GrantRedHareOnAttack
}

// isQualifiedHumanStaticModifier 驗證條件式 static 只移除 Pride 並授予對應的 On Attack ability。
// 輸入為靜態能力的 modifier；輸出為是否符合 qualified Human 定義，副作用為零。
func isQualifiedHumanStaticModifier(modifier continuousModifier) bool {
	return modifier.SetPride == nil &&
		modifier.SetPower == nil &&
		modifier.SetLife == nil &&
		modifier.PowerDelta == 0 &&
		modifier.LifeDelta == 0 &&
		modifier.ReserveCostDelta == 0 &&
		!modifier.GrantImmortality &&
		!modifier.ProhibitRecover &&
		!modifier.SwitchPowerLife &&
		!modifier.GrantStealth &&
		!modifier.GrantTrueSight &&
		modifier.RemovePride &&
		modifier.GrantRedHareOnAttack
}

// isOtherControlledSuitedAllyModifier 驗證 Verita 靜態能力只授予 Immortality。
// 輸入為連續效果修正；輸出為是否符合已支援的修正形狀，副作用為零。
func isOtherControlledSuitedAllyModifier(modifier continuousModifier) bool {
	return modifier.SetPride == nil && modifier.SetPower == nil && modifier.SetLife == nil && modifier.PowerDelta == 0 && modifier.LifeDelta == 0 && modifier.ReserveCostDelta == 0 && modifier.GrantImmortality && !modifier.ProhibitRecover && !modifier.SwitchPowerLife && !modifier.GrantStealth && !modifier.GrantTrueSight && !modifier.RemovePride && !modifier.GrantRedHareOnAttack
}

// cloneStaticEffectDefinition 複製 static 定義及其指標欄位，隔離編寫資料和編譯結果。
// 輸入為已驗證的靜態效果；輸出為獨立的靜態效果，副作用為零。
func cloneStaticEffectDefinition(effect staticEffectDefinition) staticEffectDefinition {
	clone := effect
	if effect.modifier.SetPride != nil {
		pride := *effect.modifier.SetPride
		clone.modifier.SetPride = &pride
	}
	return clone
}

// validateValueExpression 驗證整數樹的 kind、運算元與 reference 型別。
// 輸入為編寫的整數樹；輸出為驗證錯誤或 nil，副作用為零。
func validateValueExpression(value valueExpression) error {
	return validateValueNode(&value, make(map[*valueExpression]bool), 0)
}

// validateValueNode 限制樹深並拒絕循環指標，避免壞定義在載入時造成 panic。
// 輸入為節點、目前路徑與深度；輸出為驗證錯誤或 nil，副作用僅更新呼叫內路徑。
func validateValueNode(value *valueExpression, path map[*valueExpression]bool, depth int) error {
	if depth > 32 || path[value] {
		return fmt.Errorf("value expression is cyclic or too deep")
	}
	path[value] = true
	defer delete(path, value)
	switch value.Kind {
	case valueConstant:
		if value.Constant < 0 || value.Left != nil || value.Right != nil || value.Selector != "" || value.Player != "" {
			return fmt.Errorf("invalid constant payload")
		}
	case valueAdd:
		if value.Left == nil || value.Right == nil || value.Constant != 0 || value.Selector != "" || value.Player != "" {
			return fmt.Errorf("add requires exactly two operands")
		}
		if err := validateValueNode(value.Left, path, depth+1); err != nil {
			return fmt.Errorf("left: %w", err)
		}
		if err := validateValueNode(value.Right, path, depth+1); err != nil {
			return fmt.Errorf("right: %w", err)
		}
	case valueDistinctPrintedCosts:
		if value.Selector != selectorControlledSuited || value.Player != referenceController || value.Constant != 0 || value.Left != nil || value.Right != nil {
			return fmt.Errorf("distinct printed costs requires controlled-suited selector and controller reference")
		}
	default:
		return fmt.Errorf("unknown value kind %q", value.Kind)
	}
	return nil
}

// cloneValueExpression 複製整數樹，隔離可變的 Go 編寫資料與編譯結果。
// 輸入為已驗證的 expression；輸出為獨立運算樹，副作用為零。
func cloneValueExpression(value valueExpression) valueExpression {
	clone := value
	if value.Left != nil {
		left := cloneValueExpression(*value.Left)
		clone.Left = &left
	}
	if value.Right != nil {
		right := cloneValueExpression(*value.Right)
		clone.Right = &right
	}
	return clone
}

// operations 將已編譯效果轉成目前統一 Ability Runtime 的可序列化操作。
// 輸入為已驗證的能力；輸出為每次呼叫新建的操作序列，副作用為零。
func (ability compiledAbilityDefinition) operations() []effectOperation {
	operations := make([]effectOperation, 0, len(ability.effects))
	for _, effect := range ability.effects {
		switch effect.kind {
		case effectKindDraw:
			operations = append(operations, effectOperation{
				Kind:   effectOperationDraw,
				Amount: effect.draw.amount,
			})
		case effectKindDrawToMemory:
			operations = append(operations, effectOperation{
				Kind:   effectOperationDrawToMemory,
				Amount: effect.draw.amount,
			})
		case effectKindDamage:
			amount := cloneValueExpression(effect.damage.amount)
			operations = append(operations, effectOperation{
				Kind:            effectOperationDamage,
				TargetReference: effect.damage.target,
				Value:           &amount,
			})
		case effectKindChooseCard:
			operations = append(operations, effectOperation{
				Kind:          effectOperationChooseZoneCard,
				CardSelection: effect.chooseCard.selection,
				Binding:       effect.chooseCard.binding,
			})
		case effectKindChooseObject:
			operations = append(operations, effectOperation{
				Kind:     effectOperationChoose,
				Selector: effect.chooseObject.selector,
			})
		case effectKindDiscard:
			operations = append(operations, effectOperation{
				Kind:    effectOperationDiscard,
				Binding: effect.discard.binding,
			})
		case effectKindDeploy:
			operations = append(operations, effectOperation{
				Kind:    effectOperationDeploy,
				Binding: effect.deploy.binding,
			})
		case effectKindCounter:
			operations = append(operations, effectOperation{
				Kind:    effectOperationCounter,
				Counter: effect.counter.counter,
				Amount:  effect.counter.amount,
			})
		case effectKindModifier:
			operations = append(operations, effectOperation{
				Kind:                    effectOperationContinuousModifier,
				Duration:                effect.modifier.duration,
				RequiresRetargetSuccess: effect.modifier.requiresRetargetSuccess,
				ContinuousEffect: continuousEffect{
					Scope:     effectScopeObject,
					Layer:     effectLayerModifier,
					PowerLife: powerLifeModify,
					Modifier:  effect.modifier.modifier,
				},
			})
		case effectKindCopyAction:
			operations = append(operations, effectOperation{
				Kind:          effectOperationCopyAction,
				Binding:       effect.copyAction.binding,
				CardSelection: effect.chooseCard.selection,
			})
		case effectKindRetargetAttack:
			operations = append(operations, effectOperation{
				Kind: effectOperationRetargetAttack,
			})
		case effectKindMoveSourceToGraveyard:
			operations = append(operations, effectOperation{
				Kind:                  effectOperationMove,
				MoveSourceToGraveyard: true,
			})
		}
	}
	return operations
}

// compiledAbility 取得來源卡指定 kind 的已編譯定義，不讀取可變對局狀態以外的編寫資料。
// 輸入為卡牌實例與能力種類；輸出為能力及存在旗標，副作用為零。
func (g *Game) compiledAbility(card cardInstanceID, kind abilityDefinitionKind) (compiledAbilityDefinition, bool) {
	definition, exists := g.definitions[g.state.Cards[card].Definition]
	if !exists {
		return compiledAbilityDefinition{}, false
	}
	for _, ability := range definition.abilities {
		if ability.kind == kind {
			return ability, true
		}
	}
	return compiledAbilityDefinition{}, false
}

// compiledAction 取得來源卡已編譯的 Action 定義。
// 輸入為卡牌實例；輸出為 Action 能力與存在旗標，副作用為零。
func (g *Game) compiledAction(card cardInstanceID) (compiledAbilityDefinition, bool) {
	return g.compiledAbility(card, abilityKindAction)
}

// compiledCardistry 取得來源卡的已編譯 Cardistry 定義。
// 輸入為卡牌實例；輸出為能力及存在旗標，副作用為零。
func (g *Game) compiledCardistry(card cardInstanceID) (compiledAbilityDefinition, bool) {
	return g.compiledAbility(card, abilityKindCardistry)
}

// compiledAlly 取得來源卡已編譯的 Ally 宣告定義。
// 輸入為卡牌實例；輸出為 Ally 能力及存在旗標，副作用為零。
func (g *Game) compiledAlly(card cardInstanceID) (compiledAbilityDefinition, bool) {
	return g.compiledAbility(card, abilityKindAlly)
}

// compileReplayDefinitions 重建重播所需的固定可執行定義，不將編寫資料寫入 Game State。
// 輸入為零；輸出為目前引擎版本釘選的定義或驗證錯誤，副作用為零。
func compileReplayDefinitions() (map[CardID]CardDefinition, error) {
	definitions := make(map[CardID]CardDefinition)
	for _, id := range []CardID{
		wonderlandsReignCardID,
		straightFlareCardID,
		blazingThrowCardID,
		fieryInterferenceCardID,
		trumpSetCardID,
		threeOfHeartsCardID,
		duchessCardID,
		fiveOfSpadesCardID,
		fourOfSpadesCardID,
		fourOfHeartsCardID,
		threeOfSpadesCardID,
		twoOfHeartsCardID,
		twoOfSpadesCardID,
		impactHammerCardID,
		redHareCardID,
		infernalVesselCardID,
		veritaCardID,
	} {
		definition := CardDefinition{
			id: id,
			face: CardFace{
				id: CardFaceID("face:" + string(id) + ":front"),
			},
		}
		abilities, err := compileAbilityDefinitions(definition, authoredAbilitiesForCard(id))
		if err != nil {
			return nil, fmt.Errorf("compile replay definition: %w", err)
		}
		definition.abilities = abilities
		definitions[id] = definition
	}
	return definitions, nil
}
