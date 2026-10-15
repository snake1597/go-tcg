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
	abilityKindObject      abilityDefinitionKind = "object"
	abilityKindPermission  abilityDefinitionKind = "permission"
	abilityKindKeyword     abilityDefinitionKind = "keyword"
)

// abilitySupportStatus 是 compiler 與建局 gate 共用的 production 支援判定。
type abilitySupportStatus string

const (
	supportStatusSupported   abilitySupportStatus = "supported"
	supportStatusUnsupported abilitySupportStatus = "unsupported"
	supportStatusNeedsRuling abilitySupportStatus = "needs-ruling"
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
	effectKindDraw                   authoredEffectKind = "draw"
	effectKindDrawToMemory           authoredEffectKind = "draw-to-memory"
	effectKindDamage                 authoredEffectKind = "damage"
	effectKindChooseCard             authoredEffectKind = "choose-card"
	effectKindChooseObject           authoredEffectKind = "choose-object"
	effectKindDiscard                authoredEffectKind = "discard"
	effectKindDeploy                 authoredEffectKind = "deploy"
	effectKindCounter                authoredEffectKind = "counter"
	effectKindModifier               authoredEffectKind = "modifier"
	effectKindCopyAction             authoredEffectKind = "copy-action"
	effectKindRetargetAttack         authoredEffectKind = "retarget-attack"
	effectKindMoveSourceToGraveyard  authoredEffectKind = "move-source-to-graveyard"
	effectKindRestSource             authoredEffectKind = "rest-source"
	effectKindBanishSource           authoredEffectKind = "banish-source"
	effectKindCardistryDiscount      authoredEffectKind = "cardistry-discount"
	effectKindPreventNonCombatDamage authoredEffectKind = "prevent-non-combat-damage"
	effectKindRestObject             authoredEffectKind = "rest-object"
	effectKindSacrificeAllyPower     authoredEffectKind = "sacrifice-ally-power"
	effectKindGrantTaunt             authoredEffectKind = "grant-taunt"
)

type abilityReference string

const (
	referenceController     abilityReference = "controller"
	referenceDeclaredTarget abilityReference = "declared-target"
	referenceEventTarget    abilityReference = "event-target"
)

type eventKind string

const (
	eventKindWield              eventKind = "wield"
	eventKindDestroy            eventKind = "destroy"
	eventKindSacrifice          eventKind = "sacrifice"
	eventKindCardistryActivated eventKind = "cardistry-activated"
	eventKindEnter              eventKind = "enter"
	eventKindAttack             eventKind = "attack"
	eventKindSetup              eventKind = "setup"
)

// triggerCondition 限制 triggered ability 可使用的可編譯條件原語。
type triggerCondition string

const (
	triggerConditionClassMatch         triggerCondition = "class-match"
	triggerConditionQualifiedHumanAlly triggerCondition = "qualified-human-ally"
)

type keywordDefinition string

const keywordHindered keywordDefinition = "hindered"

type selectorKind string

const (
	selectorUnits                            selectorKind = "units"
	selectorControlledSuited                 selectorKind = "controlled-suited"
	selectorControlledSuitedAlly             selectorKind = "controlled-suited-ally"
	selectorAllAllies                        selectorKind = "all-allies"
	selectorRetargetableControlledSuitedAlly selectorKind = "retargetable-controlled-suited-ally"
	selectorOtherControlledAllies            selectorKind = "other-controlled-allies"
)

type targetSelector struct {
	kind selectorKind
}

type valueKind string

const (
	valueConstant               valueKind = "constant"
	valueAdd                    valueKind = "add"
	valueCount                  valueKind = "count"
	valueSumPrintedReserveCosts valueKind = "sum-printed-reserve-costs"
	valueThreshold              valueKind = "threshold"
	valueDistinctPrintedCosts   valueKind = "distinct-printed-costs"
)

// valueExpression 是可序列化的整數運算樹；左右運算元僅用於加法。
type valueExpression struct {
	Kind       valueKind        `json:"kind"`
	Constant   int              `json:"constant,omitempty"`
	Left       *valueExpression `json:"left,omitempty"`
	Right      *valueExpression `json:"right,omitempty"`
	Selector   selectorKind     `json:"selector,omitempty"`
	Player     abilityReference `json:"player,omitempty"`
	Multiplier int              `json:"multiplier,omitempty"`
}

type drawEffectDefinition struct {
	amount    int
	recipient abilityReference
}

type chooseObjectEffectDefinition struct {
	selector selectorKind
	binding  resolutionBinding
}

type deployEffectDefinition struct {
	binding resolutionBinding
}

type counterEffectDefinition struct {
	counter       string
	amount        valueExpression
	targetBinding resolutionBinding
}

type modifierEffectDefinition struct {
	duration                modifierDuration
	modifier                continuousModifier
	requiresRetargetSuccess bool
}

type damageEffectDefinition struct {
	target        abilityReference
	targetBinding resolutionBinding
	amount        valueExpression
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
	canPass   bool
}

type discardEffectDefinition struct {
	binding resolutionBinding
}

type copyActionEffectDefinition struct {
	binding resolutionBinding
}

type triggerDefinition struct {
	event     eventKind
	condition triggerCondition
	optional  bool
}

type staticPredicateKind string

const (
	staticPredicateSelf                       staticPredicateKind = "self"
	staticPredicateControlsQualifiedHumanAlly staticPredicateKind = "controls-qualified-human-ally"
	staticPredicateOtherControlledSuitedAlly  staticPredicateKind = "other-controlled-suited-ally"
	staticPredicateClassMatchedSelf           staticPredicateKind = "class-matched-self"
	staticPredicateRestedSourceOtherAllies    staticPredicateKind = "rested-source-other-allies"
	staticPredicateSelfWithOtherSuitedAlly    staticPredicateKind = "self-with-other-suited-ally"
	staticPredicateSelfChampionDamaged        staticPredicateKind = "self-champion-damaged"
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

type permissionDefinition struct {
	candidateElement     string
	reserveCostDelta     int
	duringControllerTurn bool
	opponentsOnly        bool
	sourcePresence       sourcePresence
}

type modifierDuration string

const (
	modifierDurationEndOfTurn     modifierDuration = "end-of-turn"
	modifierDurationEndOfNextTurn modifierDuration = "end-of-next-turn"
)

// objectActivationCondition 描述 Object ability 在宣告時必須滿足的公開盤面條件。
// 輸入為對手 Champion 所需元素；輸出供 compiler 與合法性查詢使用；不保存可變對局狀態。
type objectActivationCondition struct {
	opponentChampionElement string
}

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
	permission         *permissionDefinition
	keyword            keywordDefinition
	alternative        *alternativeCostSpec
	deathModifier      *deathModifierDefinition
	cost               *actionCostDefinition
	objectCondition    *objectActivationCondition
	effects            []authoredEffectDefinition
	supportStatus      abilitySupportStatus
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
	permission         *permissionDefinition
	keyword            keywordDefinition
	alternative        *alternativeCostSpec
	deathModifier      *deathModifierDefinition
	cost               *actionCostDefinition
	objectCondition    *objectActivationCondition
	effects            []compiledEffect
	supportStatus      abilitySupportStatus
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
						amount: valueExpression{
							Kind:     valueConstant,
							Constant: 1,
						},
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

// impactHammerAbilities 宣告 Impact Hammer 的職業加成 static 與 wield 傷害 triggered slots。
// 輸入為零；輸出為新的 Go 編寫資料，副作用為零。
func impactHammerAbilities() []authoredAbilityDefinition {
	return []authoredAbilityDefinition{
		{
			slot: "ability:chsbalegbs:front:class-bonus",
			kind: abilityKindStatic,
			static: &staticEffectDefinition{
				predicate: staticPredicateClassMatchedSelf,
				layer:     effectLayerAbility,
				sublayer:  effectSubLayerNone,
				modifier: continuousModifier{
					PowerDelta: 1,
				},
				duration:       staticDurationWhilePredicate,
				sourcePresence: sourcePresenceRequired,
			},
		},
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

// arthurYoungHeirAbilities 宣告 Arthur 的進場暫時 Immortality 與休息中 Ally 加成 Slots。
// 輸入為零；輸出為不可共享的 Go 編寫資料，副作用為零。
func arthurYoungHeirAbilities() []authoredAbilityDefinition {
	return []authoredAbilityDefinition{
		{
			slot: "ability:GjM8b5fxqj:front:on-enter-rest-immortality",
			kind: abilityKindTriggered,
			trigger: &triggerDefinition{
				event: eventKindEnter,
			},
			effects: []authoredEffectDefinition{
				{
					kind: effectKindRestObject,
				},
				{
					kind: effectKindModifier,
					modifier: &modifierEffectDefinition{
						duration: modifierDurationEndOfNextTurn,
						modifier: continuousModifier{
							GrantImmortality: true,
						},
					},
				},
			},
		},
		{
			slot: "ability:GjM8b5fxqj:front:rested-allies-power",
			kind: abilityKindStatic,
			static: &staticEffectDefinition{
				predicate: staticPredicateRestedSourceOtherAllies,
				layer:     effectLayerModifier,
				sublayer:  effectSubLayerNone,
				modifier: continuousModifier{
					PowerDelta: 1,
				},
				duration:       staticDurationWhilePredicate,
				sourcePresence: sourcePresenceRequired,
			},
		},
	}
}

// noireAbilities 宣告 Noire 的 Suited 進場 BUFF 與條件式 Stealth Slots。
// 輸入為零；輸出為不可共享的 Go 編寫資料，副作用為零。
func noireAbilities() []authoredAbilityDefinition {
	return []authoredAbilityDefinition{
		{
			slot: "ability:wbjc9t8ycp:front:on-enter-suited-counters",
			kind: abilityKindTriggered,
			trigger: &triggerDefinition{
				event: eventKindEnter,
			},
			effects: []authoredEffectDefinition{
				{
					kind: effectKindCounter,
					counter: &counterEffectDefinition{
						counter: "BUFF",
						amount: valueExpression{
							Kind: valueThreshold,
							Left: &valueExpression{
								Kind:     valueSumPrintedReserveCosts,
								Selector: selectorControlledSuited,
								Player:   referenceController,
							},
							Multiplier: 1,
						},
					},
				},
			},
		},
		{
			slot: "ability:wbjc9t8ycp:front:suited-stealth",
			kind: abilityKindStatic,
			static: &staticEffectDefinition{
				predicate: staticPredicateSelfWithOtherSuitedAlly,
				layer:     effectLayerAbility,
				sublayer:  effectSubLayerNone,
				modifier: continuousModifier{
					GrantStealth: true,
				},
				duration:       staticDurationWhilePredicate,
				sourcePresence: sourcePresenceRequired,
			},
		},
	}
}

// heatedVengeanceAbilities 宣告 Champion 受傷加成與職業加成攻擊觸發 Slots。
// 輸入為零；輸出為不可共享的 Go 編寫資料，副作用為零。
func heatedVengeanceAbilities() []authoredAbilityDefinition {
	return []authoredAbilityDefinition{
		{
			slot: "ability:td460e8ig0:front:damaged-champion-power",
			kind: abilityKindStatic,
			static: &staticEffectDefinition{
				predicate: staticPredicateSelfChampionDamaged,
				layer:     effectLayerModifier,
				sublayer:  effectSubLayerNone,
				modifier: continuousModifier{
					PowerDelta: 3,
				},
				duration:       staticDurationWhilePredicate,
				sourcePresence: sourcePresenceRequired,
			},
		},
		{
			slot: "ability:td460e8ig0:front:on-attack-self-damage",
			kind: abilityKindTriggered,
			trigger: &triggerDefinition{
				event:     eventKindAttack,
				condition: triggerConditionClassMatch,
				optional:  true,
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

// pepperedChefAbilities 宣告進場時可犧牲另一 Ally 並讓來源本回合 +2 Power 的 Slot。
// 輸入為零；輸出為不可共享的 Go 編寫資料，副作用為零。
func pepperedChefAbilities() []authoredAbilityDefinition {
	return []authoredAbilityDefinition{
		{
			slot: "ability:lcy0lw1veb:front:on-enter-sacrifice-power",
			kind: abilityKindTriggered,
			trigger: &triggerDefinition{
				event:    eventKindEnter,
				optional: true,
			},
			effects: []authoredEffectDefinition{
				{
					kind: effectKindChooseObject,
					chooseObject: &chooseObjectEffectDefinition{
						selector: selectorOtherControlledAllies,
						binding:  "sacrifice-target",
					},
				},
				{
					kind: effectKindSacrificeAllyPower,
				},
			},
		},
	}
}

// rougeAbilities 宣告進場時依 Suited Reserve 門檻對任一單位造成傷害的 Slot。
// 輸入為零；輸出為不可共享的 Go 編寫資料，副作用為零。
func rougeAbilities() []authoredAbilityDefinition {
	return []authoredAbilityDefinition{
		{
			slot: "ability:h68dr63eo5:front:on-enter-suited-threshold-damage",
			kind: abilityKindTriggered,
			trigger: &triggerDefinition{
				event: eventKindEnter,
			},
			effects: []authoredEffectDefinition{
				{
					kind: effectKindChooseObject,
					chooseObject: &chooseObjectEffectDefinition{
						selector: selectorUnits,
						binding:  "damage-target",
					},
				},
				{
					kind: effectKindDamage,
					damage: &damageEffectDefinition{
						targetBinding: "damage-target",
						amount: valueExpression{
							Kind: valueThreshold,
							Left: &valueExpression{
								Kind:     valueSumPrintedReserveCosts,
								Selector: selectorControlledSuited,
								Player:   referenceController,
							},
							Multiplier: 2,
						},
					},
				},
			},
		},
	}
}

// spiritOfFireAbilities 宣告初始 Champion 在 setup 進場時抽七張牌的 Slot。
// 輸入為零；輸出為不可共享的 Go 編寫資料，副作用為零。
func spiritOfFireAbilities() []authoredAbilityDefinition {
	return []authoredAbilityDefinition{
		{
			slot: "ability:LMyKyVC2O9:front:on-enter-draw-seven",
			kind: abilityKindTriggered,
			trigger: &triggerDefinition{
				event: eventKindSetup,
			},
			effects: []authoredEffectDefinition{
				{
					kind: effectKindDraw,
					draw: &drawEffectDefinition{
						amount:    7,
						recipient: referenceController,
					},
				},
			},
		},
	}
}

// tonorisAbilities 宣告 Champion 升級進場後直到控制者下回合的 Taunt Slot。
// 輸入為零；輸出為不可共享的 Go 編寫資料，副作用為零。
func tonorisAbilities() []authoredAbilityDefinition {
	return []authoredAbilityDefinition{
		{
			slot: "ability:zb14m4c8lj:front:on-enter-taunt",
			kind: abilityKindTriggered,
			trigger: &triggerDefinition{
				event: eventKindEnter,
			},
			effects: []authoredEffectDefinition{
				{
					kind: effectKindGrantTaunt,
				},
			},
		},
	}
}

// bulwarkSwordAbilities 宣告 Bulwark Sword 的職業加成與 wield 額外費用 static slot。
// 輸入為零；輸出為不可共享的 Go 編寫資料，副作用為零。
func bulwarkSwordAbilities() []authoredAbilityDefinition {
	return []authoredAbilityDefinition{
		{
			slot: "ability:8kmoi0a5uh:front:class-bonus-wield-cost",
			kind: abilityKindStatic,
			static: &staticEffectDefinition{
				predicate: staticPredicateClassMatchedSelf,
				layer:     effectLayerAbility,
				sublayer:  effectSubLayerNone,
				modifier: continuousModifier{
					PowerDelta:       1,
					ReserveCostDelta: 2,
				},
				duration:       staticDurationWhilePredicate,
				sourcePresence: sourcePresenceRequired,
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
					RemovePride: true,
				},
				duration:       staticDurationWhilePredicate,
				sourcePresence: sourcePresenceRequired,
			},
		},
		{
			slot: "ability:5du8f077ua:front:qualified-human-on-attack",
			kind: abilityKindTriggered,
			trigger: &triggerDefinition{
				event:     eventKindAttack,
				condition: triggerConditionQualifiedHumanAlly,
			},
			effects: []authoredEffectDefinition{
				{
					kind: effectKindChooseCard,
					chooseCard: &chooseCardEffectDefinition{
						selection: cardSelectionSpec{
							Zone: cardZoneHand,
						},
						binding: "discard-card",
						canPass: true,
					},
				},
				{
					kind: effectKindDiscard,
					discard: &discardEffectDefinition{
						binding: "discard-card",
					},
				},
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

// duchessThornesAbilities 宣告 Hindered、Ally Cardistry 觸發與休息放逐折扣三個穩定 Slot。
// 輸入為零；輸出為不共享可變狀態的 Go 編寫資料，副作用為零。
func duchessThornesAbilities() []authoredAbilityDefinition {
	return []authoredAbilityDefinition{
		{
			slot:    "ability:bEXmm4rKOs:front:hindered",
			kind:    abilityKindKeyword,
			keyword: keywordHindered,
		},
		{
			slot: "ability:bEXmm4rKOs:front:ally-cardistry-bonus",
			kind: abilityKindTriggered,
			trigger: &triggerDefinition{
				event: eventKindCardistryActivated,
			},
			effects: []authoredEffectDefinition{
				{
					kind: effectKindModifier,
					modifier: &modifierEffectDefinition{
						duration: modifierDurationEndOfTurn,
						modifier: continuousModifier{
							PowerDelta:     1,
							GrantTrueSight: true,
						},
					},
				},
			},
		},
		{
			slot: "ability:bEXmm4rKOs:front:rest-banish-cardistry-discount",
			kind: abilityKindObject,
			effects: []authoredEffectDefinition{
				{
					kind: effectKindRestSource,
				},
				{
					kind: effectKindBanishSource,
				},
				{
					kind: effectKindCardistryDiscount,
				},
			},
		},
	}
}

// grandCrusadersRingAbilities 宣告 Grand Crusader's Ring 的放逐自身並抽牌 Object slot。
// 輸入為零；輸出為不共享可變狀態的 Go 編寫資料，副作用為零。
func grandCrusadersRingAbilities() []authoredAbilityDefinition {
	return []authoredAbilityDefinition{
		{
			slot: "ability:2gv7DC0KID:front:banish-draw",
			kind: abilityKindObject,
			effects: []authoredEffectDefinition{
				{
					kind: effectKindBanishSource,
				},
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

// waterResonanceBaubleAbilities 宣告 Water Resonance Bauble 在對手控制 Water Champion 時可放逐自身抽牌的 Object slot。
// 輸入為零；輸出為不共享可變狀態的 Go 編寫資料，副作用為零。
func waterResonanceBaubleAbilities() []authoredAbilityDefinition {
	return []authoredAbilityDefinition{
		{
			slot: "ability:dSSRtNnPtw:front:banish-draw-water-champion",
			kind: abilityKindObject,
			objectCondition: &objectActivationCondition{
				opponentChampionElement: "WATER",
			},
			effects: []authoredEffectDefinition{
				{
					kind: effectKindBanishSource,
				},
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

// windResonanceBaubleAbilities 宣告 Wind Resonance Bauble 在對手控制 Wind Champion 時可放逐自身抽牌的 Object slot。
// 輸入為零；輸出為不共享可變狀態的 Go 編寫資料，副作用為零。
func windResonanceBaubleAbilities() []authoredAbilityDefinition {
	return []authoredAbilityDefinition{
		{
			slot: "ability:bHGUNMFLg9:front:banish-draw-wind-champion",
			kind: abilityKindObject,
			objectCondition: &objectActivationCondition{
				opponentChampionElement: "WIND",
			},
			effects: []authoredEffectDefinition{
				{
					kind: effectKindBanishSource,
				},
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

// smokeBombsAbilities 宣告 Smoke Bombs 的放逐、選擇任一 Ally、賦予 stealth 與抽牌 Object slot。
// 輸入為零；輸出為不共享可變狀態的 Go 編寫資料，副作用為零。
func smokeBombsAbilities() []authoredAbilityDefinition {
	return []authoredAbilityDefinition{
		{
			slot: "ability:ScGcOmkoQt:front:banish-stealth-draw",
			kind: abilityKindObject,
			effects: []authoredEffectDefinition{
				{
					kind: effectKindBanishSource,
				},
				{
					kind: effectKindChooseObject,
					chooseObject: &chooseObjectEffectDefinition{
						selector: selectorAllAllies,
					},
				},
				{
					kind: effectKindModifier,
					modifier: &modifierEffectDefinition{
						duration: modifierDurationEndOfTurn,
						modifier: continuousModifier{
							GrantStealth: true,
						},
					},
				},
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

// safeguardAmuletAbilities 宣告 Safeguard Amulet 的放逐自身與本回合非戰鬥傷害防止 Object slot。
// 輸入為零；輸出為不共享可變狀態的 Go 編寫資料，副作用為零。
func safeguardAmuletAbilities() []authoredAbilityDefinition {
	return []authoredAbilityDefinition{
		{
			slot: "ability:yj2rJBREH8:front:banish-prevent-non-combat-damage",
			kind: abilityKindObject,
			effects: []authoredEffectDefinition{
				{
					kind: effectKindBanishSource,
				},
				{
					kind: effectKindPreventNonCombatDamage,
				},
			},
		},
	}
}

// viridianProtectiveTrinketAbilities 宣告控制者回合中對手 Water 卡 activation Reserve tax permission slot。
// 輸入為零；輸出為不可共享的 Go 編寫資料，副作用為零。
func viridianProtectiveTrinketAbilities() []authoredAbilityDefinition {
	return []authoredAbilityDefinition{
		{
			slot: "ability:s3572j3oda:front:opponent-water-activation-tax",
			kind: abilityKindPermission,
			permission: &permissionDefinition{
				candidateElement:     "WATER",
				reserveCostDelta:     2,
				duringControllerTurn: true,
				opponentsOnly:        true,
				sourcePresence:       sourcePresenceRequired,
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
	case arthurYoungHeirCardID:
		return arthurYoungHeirAbilities()
	case noireCardID:
		return noireAbilities()
	case pepperedChefCardID:
		return pepperedChefAbilities()
	case heatedVengeanceCardID:
		return heatedVengeanceAbilities()
	case rougeCardID:
		return rougeAbilities()
	case spiritOfFireCardID:
		return spiritOfFireAbilities()
	case tonorisCardID:
		return tonorisAbilities()
	case bulwarkSwordCardID:
		return bulwarkSwordAbilities()
	case redHareCardID:
		return redHareAbilities()
	case infernalVesselCardID:
		return infernalVesselAbilities()
	case veritaCardID:
		return veritaAbilities()
	case grandCrusadersRingCardID:
		return grandCrusadersRingAbilities()
	case waterResonanceBaubleCardID:
		return waterResonanceBaubleAbilities()
	case windResonanceBaubleCardID:
		return windResonanceBaubleAbilities()
	case smokeBombsCardID:
		return smokeBombsAbilities()
	case safeguardAmuletCardID:
		return safeguardAmuletAbilities()
	case viridianProtectiveTrinketCardID:
		return viridianProtectiveTrinketAbilities()
	case duchessThornesCardID:
		return duchessThornesAbilities()
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
		status := ability.supportStatus
		if status == "" {
			status = supportStatusSupported
		}
		if status != supportStatusSupported {
			return nil, fmt.Errorf("%s: support status %q cannot enter production", context, status)
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
		case abilityKindObject:
			if ability.timing != "" || ability.usage != "" || ability.reduction != "" || ability.baseCost != 0 || ability.classCostReduction != 0 || ability.target != nil || ability.trigger != nil || ability.static != nil || ability.replacement != nil || ability.alternative != nil || ability.deathModifier != nil || ability.cost != nil {
				return nil, fmt.Errorf("%s: Object declaration payload unsupported", context)
			}
			if err := validateObjectActivationCondition(ability.objectCondition); err != nil {
				return nil, fmt.Errorf("%s: Object condition: %w", context, err)
			}
		case abilityKindTriggered:
			if ability.timing != "" || ability.usage != "" || ability.reduction != "" || ability.baseCost != 0 || ability.classCostReduction != 0 || ability.target != nil || ability.cost != nil {
				return nil, fmt.Errorf("%s: triggered declaration payload unsupported", context)
			}
			if ability.trigger == nil || (ability.trigger.event != eventKindWield && ability.trigger.event != eventKindDestroy && ability.trigger.event != eventKindSacrifice && ability.trigger.event != eventKindCardistryActivated && ability.trigger.event != eventKindEnter && ability.trigger.event != eventKindAttack && ability.trigger.event != eventKindSetup) {
				return nil, fmt.Errorf("%s: trigger.event unsupported", context)
			}
			if ability.trigger.condition != "" && ability.trigger.condition != triggerConditionClassMatch && ability.trigger.condition != triggerConditionQualifiedHumanAlly {
				return nil, fmt.Errorf("%s: trigger.condition unsupported %q", context, ability.trigger.condition)
			}
			if ability.trigger.condition != "" && ability.trigger.event != eventKindAttack {
				return nil, fmt.Errorf("%s: trigger condition requires attack event", context)
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
		case abilityKindPermission:
			if ability.timing != "" || ability.usage != "" || ability.reduction != "" || ability.baseCost != 0 || ability.classCostReduction != 0 || ability.target != nil || ability.trigger != nil || ability.static != nil || ability.replacement != nil || ability.alternative != nil || ability.deathModifier != nil || ability.cost != nil || ability.objectCondition != nil || len(ability.effects) != 0 {
				return nil, fmt.Errorf("%s: permission declaration payload unsupported", context)
			}
			if err := validatePermissionDefinition(ability.permission); err != nil {
				return nil, fmt.Errorf("%s: permission: %w", context, err)
			}
		case abilityKindKeyword:
			if ability.keyword != keywordHindered || ability.timing != "" || ability.usage != "" || ability.reduction != "" || ability.baseCost != 0 || ability.classCostReduction != 0 || ability.target != nil || ability.trigger != nil || ability.static != nil || ability.replacement != nil || ability.permission != nil || ability.alternative != nil || ability.deathModifier != nil || ability.cost != nil || ability.objectCondition != nil || len(ability.effects) != 0 {
				return nil, fmt.Errorf("%s: keyword declaration payload unsupported", context)
			}
		default:
			return nil, fmt.Errorf("%s: unknown kind %q", context, ability.kind)
		}
		if ability.kind != abilityKindStatic && ability.kind != abilityKindReplacement && ability.kind != abilityKindPermission && ability.kind != abilityKindKeyword && ability.kind != abilityKindAlly && !(ability.kind == abilityKindTriggered && ability.trigger != nil && (ability.trigger.event == eventKindDestroy || ability.trigger.event == eventKindSacrifice)) && len(ability.effects) == 0 {
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
			keyword:            ability.keyword,
			effects:            make([]compiledEffect, 0, len(ability.effects)),
			supportStatus:      status,
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
		if ability.permission != nil {
			permission := *ability.permission
			result.permission = &permission
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
		if ability.objectCondition != nil {
			condition := *ability.objectCondition
			result.objectCondition = &condition
		}
		bindings := make(map[resolutionBinding]cardSelectionSpec)
		objectBindings := make(map[resolutionBinding]struct{})
		qualifiedAttack := ability.kind == abilityKindTriggered && ability.trigger != nil && ability.trigger.event == eventKindAttack && ability.trigger.condition == triggerConditionQualifiedHumanAlly
		for index, effect := range ability.effects {
			field := fmt.Sprintf("effects[%d]", index)
			switch effect.kind {
			case effectKindDraw:
				setupDraw := ability.kind == abilityKindTriggered && ability.trigger != nil && ability.trigger.event == eventKindSetup
				if (ability.kind != abilityKindCardistry && ability.kind != abilityKindObject && !setupDraw && !qualifiedAttack) || effect.draw == nil || effect.damage != nil {
					return nil, fmt.Errorf("%s: %s.draw payload requires Cardistry, Object, setup or qualified attack trigger", context, field)
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
				if effect.damage.targetBinding == "" && effect.damage.target != expectedTarget {
					return nil, fmt.Errorf("%s: %s.damage.target has wrong reference %q", context, field, effect.damage.target)
				}
				if effect.damage.targetBinding != "" {
					if ability.kind != abilityKindTriggered {
						return nil, fmt.Errorf("%s: %s.damage.targetBinding requires a triggered ability", context, field)
					}
					if effect.damage.target != "" {
						return nil, fmt.Errorf("%s: %s.damage target and targetBinding are mutually exclusive", context, field)
					}
					if _, exists := objectBindings[effect.damage.targetBinding]; !exists {
						return nil, fmt.Errorf("%s: %s.damage.targetBinding %q has no preceding Object choice", context, field, effect.damage.targetBinding)
					}
				}
				if err := validateValueExpression(effect.damage.amount); err != nil {
					return nil, fmt.Errorf("%s: %s.damage.amount: %w", context, field, err)
				}
				result.effects = append(result.effects, compiledEffect{
					kind: effectKindDamage,
					damage: damageEffectDefinition{
						target:        effect.damage.target,
						targetBinding: effect.damage.targetBinding,
						amount:        cloneValueExpression(effect.damage.amount),
					},
				})
			case effectKindChooseCard:
				if (ability.kind != abilityKindCardistry && !qualifiedAttack) || effect.chooseCard == nil || effect.draw != nil || effect.damage != nil || effect.discard != nil || effect.copyAction != nil {
					return nil, fmt.Errorf("%s: %s.choose_card payload requires Cardistry or qualified attack trigger", context, field)
				}
				if effect.chooseCard.canPass != qualifiedAttack {
					return nil, fmt.Errorf("%s: %s.choose_card.can_pass unsupported", context, field)
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
				triggeredEnter := ability.kind == abilityKindTriggered && ability.trigger != nil && ability.trigger.event == eventKindEnter
				supportedSelector := effect.chooseObject != nil && (effect.chooseObject.selector == selectorControlledSuitedAlly || effect.chooseObject.selector == selectorAllAllies || effect.chooseObject.selector == selectorOtherControlledAllies || effect.chooseObject.selector == selectorUnits)
				if (ability.kind != abilityKindCardistry && ability.kind != abilityKindObject && !triggeredEnter) || !supportedSelector {
					return nil, fmt.Errorf("%s: %s.choose_object requires a supported Ally selector", context, field)
				}
				if effect.chooseObject.binding != "" {
					if _, exists := bindings[effect.chooseObject.binding]; exists {
						return nil, fmt.Errorf("%s: %s.choose_object.binding %q is duplicated", context, field, effect.chooseObject.binding)
					}
					if _, exists := objectBindings[effect.chooseObject.binding]; exists {
						return nil, fmt.Errorf("%s: %s.choose_object.binding %q is duplicated", context, field, effect.chooseObject.binding)
					}
					objectBindings[effect.chooseObject.binding] = struct{}{}
				}
				result.effects = append(result.effects, compiledEffect{
					kind:         effectKindChooseObject,
					chooseObject: *effect.chooseObject,
				})
			case effectKindDiscard:
				if (ability.kind != abilityKindCardistry && !qualifiedAttack) || effect.discard == nil || effect.draw != nil || effect.damage != nil || effect.chooseCard != nil || effect.copyAction != nil {
					return nil, fmt.Errorf("%s: %s.discard payload requires Cardistry or qualified attack trigger", context, field)
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
				if (ability.kind != abilityKindCardistry && ability.kind != abilityKindTriggered) || effect.counter == nil || effect.counter.counter != "BUFF" {
					return nil, fmt.Errorf("%s: %s.counter requires a BUFF counter for Cardistry or triggered", context, field)
				}
				if err := validateValueExpression(effect.counter.amount); err != nil {
					return nil, fmt.Errorf("%s: %s.counter.amount: %w", context, field, err)
				}
				if effect.counter.targetBinding != "" {
					if _, exists := objectBindings[effect.counter.targetBinding]; !exists {
						return nil, fmt.Errorf("%s: %s.counter.targetBinding %q has no preceding Object choice", context, field, effect.counter.targetBinding)
					}
				}
				result.effects = append(result.effects, compiledEffect{
					kind:    effectKindCounter,
					counter: *effect.counter,
				})
			case effectKindModifier:
				if !validAbilityModifier(ability, effect.modifier) {
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
			case effectKindRestSource:
				if ability.kind != abilityKindObject || index != 0 {
					return nil, fmt.Errorf("%s: %s.rest_source must be the first Object effect", context, field)
				}
				result.effects = append(result.effects, compiledEffect{
					kind: effectKindRestSource,
				})
			case effectKindBanishSource:
				if ability.kind != abilityKindObject || (index != 0 && (index != 1 || ability.effects[0].kind != effectKindRestSource)) {
					return nil, fmt.Errorf("%s: %s.banish_source must follow optional Object rest source", context, field)
				}
				result.effects = append(result.effects, compiledEffect{
					kind: effectKindBanishSource,
				})
			case effectKindCardistryDiscount:
				if ability.kind != abilityKindObject || index == 0 {
					return nil, fmt.Errorf("%s: %s.cardistry_discount requires Object costs", context, field)
				}
				result.effects = append(result.effects, compiledEffect{
					kind: effectKindCardistryDiscount,
				})
			case effectKindPreventNonCombatDamage:
				if ability.kind != abilityKindObject || index == 0 {
					return nil, fmt.Errorf("%s: %s requires a preceding Object banish source", context, field)
				}
				result.effects = append(result.effects, compiledEffect{
					kind: effectKindPreventNonCombatDamage,
				})
			case effectKindRestObject, effectKindSacrificeAllyPower, effectKindGrantTaunt:
				if ability.kind != abilityKindTriggered || ability.trigger == nil || ability.trigger.event != eventKindEnter {
					return nil, fmt.Errorf("%s: %s.kind requires enter trigger", context, field)
				}
				result.effects = append(result.effects, compiledEffect{
					kind: effect.kind,
				})
			default:
				return nil, fmt.Errorf("%s: %s.kind unknown %q", context, field, effect.kind)
			}
		}
		if ability.kind == abilityKindAction && (len(result.effects) == 0 || result.effects[len(result.effects)-1].kind != effectKindMoveSourceToGraveyard) {
			return nil, fmt.Errorf("%s: Action must end by moving its source to graveyard", context)
		}
		if ability.kind == abilityKindObject && !validObjectEffectSequence(result.effects) {
			return nil, fmt.Errorf("%s: Object must pay optional rest then banish before its resolving effects", context)
		}
		if err := validateAbilityOperations(result.operations()); err != nil {
			return nil, fmt.Errorf("%s: runtime operations: %w", context, err)
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

// validAbilityModifier 限制各能力生命週期可提交的持續效果與到期時點。
// 輸入為能力宣告與修正資料；輸出為是否合法，無副作用。
func validAbilityModifier(ability authoredAbilityDefinition, modifier *modifierEffectDefinition) bool {
	if modifier == nil {
		return false
	}
	switch ability.kind {
	case abilityKindCardistry:
		return modifier.duration == modifierDurationEndOfTurn && !modifier.requiresRetargetSuccess && isCardistryModifier(modifier.modifier)
	case abilityKindAction:
		return modifier.duration == modifierDurationEndOfTurn && isActionModifier(modifier.modifier) && (!modifier.requiresRetargetSuccess || (ability.target != nil && ability.target.kind == selectorRetargetableControlledSuitedAlly))
	case abilityKindObject:
		return modifier.duration == modifierDurationEndOfTurn && !modifier.requiresRetargetSuccess && modifier.modifier.GrantStealth
	case abilityKindTriggered:
		if ability.trigger == nil || modifier.requiresRetargetSuccess {
			return false
		}
		if ability.trigger.event == eventKindEnter {
			return modifier.duration == modifierDurationEndOfNextTurn && isOtherControlledSuitedAllyModifier(modifier.modifier)
		}
		return ability.trigger.event == eventKindCardistryActivated && modifier.duration == modifierDurationEndOfTurn && isCardistryObserverModifier(modifier.modifier)
	default:
		return false
	}
}

// validateObjectActivationCondition 限制目前支援集合的 Object 宣告條件為對手 Water 或 Wind Champion。
// 輸入為可選 Object 宣告條件；輸出為驗證錯誤或 nil，副作用為零。
func validateObjectActivationCondition(condition *objectActivationCondition) error {
	if condition == nil {
		return nil
	}
	if condition.opponentChampionElement != "WATER" && condition.opponentChampionElement != "WIND" {
		return fmt.Errorf("opponent champion element unsupported %q", condition.opponentChampionElement)
	}
	return nil
}

// validObjectEffectSequence 驗證 Object 先支付可選 rest 與必要 banish，再保留至少一個結算效果。
// 輸入為已編譯效果；輸出為是否符合共用 Object 宣告生命週期，副作用為零。
func validObjectEffectSequence(effects []compiledEffect) bool {
	if len(effects) < 2 {
		return false
	}
	banishIndex := 0
	if effects[0].kind == effectKindRestSource {
		banishIndex = 1
	}
	return banishIndex < len(effects)-1 && effects[banishIndex].kind == effectKindBanishSource
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

// validatePermissionDefinition 限制 permission 為場上來源於控制者回合對對手 Water activation 增加 2 Reserve。
// 輸入為 permission 定義；輸出為驗證錯誤或 nil，副作用為零。
func validatePermissionDefinition(definition *permissionDefinition) error {
	if definition == nil || definition.candidateElement != "WATER" || definition.reserveCostDelta != 2 || !definition.duringControllerTurn || !definition.opponentsOnly || definition.sourcePresence != sourcePresenceRequired {
		return fmt.Errorf("unsupported activation tax permission")
	}
	return nil
}

// validateStaticEffectDefinition 限制目前支援集合可使用的 static predicate、layer、duration 與 modifier。
// 輸入為編寫的靜態效果；輸出為驗證錯誤或 nil，副作用為零。
func validateStaticEffectDefinition(effect *staticEffectDefinition) error {
	if effect == nil {
		return fmt.Errorf("definition is required")
	}
	if (effect.layer != effectLayerAbility && effect.layer != effectLayerModifier) || effect.sublayer != effectSubLayerNone || effect.duration != staticDurationWhilePredicate || effect.sourcePresence != sourcePresenceRequired {
		return fmt.Errorf("layer, sublayer, duration, or source presence unsupported")
	}
	switch effect.predicate {
	case staticPredicateSelf:
		if effect.layer != effectLayerAbility {
			return fmt.Errorf("self predicate requires ability layer")
		}
		if !isPrideStaticModifier(effect.modifier) {
			return fmt.Errorf("self predicate requires nonnegative Pride modifier")
		}
	case staticPredicateControlsQualifiedHumanAlly:
		if effect.layer != effectLayerAbility {
			return fmt.Errorf("qualified Human predicate requires ability layer")
		}
		if !isQualifiedHumanStaticModifier(effect.modifier) {
			return fmt.Errorf("qualified Human predicate requires Pride removal")
		}
	case staticPredicateOtherControlledSuitedAlly:
		if effect.layer != effectLayerAbility {
			return fmt.Errorf("other controlled Suited Ally predicate requires ability layer")
		}
		if !isOtherControlledSuitedAllyModifier(effect.modifier) {
			return fmt.Errorf("other controlled Suited Ally predicate requires Immortality")
		}
	case staticPredicateClassMatchedSelf:
		if effect.layer != effectLayerAbility {
			return fmt.Errorf("class matched self predicate requires ability layer")
		}
		if !isClassBonusStaticModifier(effect.modifier) {
			return fmt.Errorf("class matched self predicate requires +1 Power and optional +2 Reserve Cost")
		}
	case staticPredicateRestedSourceOtherAllies:
		if effect.layer != effectLayerModifier {
			return fmt.Errorf("rested source predicate requires modifier layer")
		}
		if !isArthurRestedAlliesModifier(effect.modifier) {
			return fmt.Errorf("rested source predicate requires +1 Power")
		}
	case staticPredicateSelfWithOtherSuitedAlly:
		if effect.layer != effectLayerAbility {
			return fmt.Errorf("other Suited Ally predicate requires ability layer")
		}
		if !isNoireStealthModifier(effect.modifier) {
			return fmt.Errorf("other Suited Ally predicate requires Stealth")
		}
	case staticPredicateSelfChampionDamaged:
		if effect.layer != effectLayerModifier {
			return fmt.Errorf("damaged Champion predicate requires modifier layer")
		}
		if !isHeatedVengeanceModifier(effect.modifier) {
			return fmt.Errorf("damaged Champion predicate requires +3 Power")
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
		!modifier.RemovePride
}

// isCardistryObserverModifier 驗證 Cardistry 觸發效果只提供 +1 Power 與 true sight。
// 輸入為 continuous modifier；輸出為是否符合共用 triggered modifier 形狀，副作用為零。
func isCardistryObserverModifier(modifier continuousModifier) bool {
	return modifier.SetPride == nil &&
		modifier.SetPower == nil &&
		modifier.SetLife == nil &&
		modifier.PowerDelta == 1 &&
		modifier.LifeDelta == 0 &&
		modifier.ReserveCostDelta == 0 &&
		!modifier.GrantImmortality &&
		!modifier.ProhibitRecover &&
		!modifier.SwitchPowerLife &&
		!modifier.GrantStealth &&
		modifier.GrantTrueSight &&
		!modifier.RemovePride
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
		!modifier.RemovePride)
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
		!modifier.RemovePride
}

// isQualifiedHumanStaticModifier 驗證條件式 static 只移除 Pride。
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
		modifier.RemovePride
}

// isOtherControlledSuitedAllyModifier 驗證 Verita 靜態能力只授予 Immortality。
// 輸入為連續效果修正；輸出為是否符合已支援的修正形狀，副作用為零。
func isOtherControlledSuitedAllyModifier(modifier continuousModifier) bool {
	return modifier.SetPride == nil && modifier.SetPower == nil && modifier.SetLife == nil && modifier.PowerDelta == 0 && modifier.LifeDelta == 0 && modifier.ReserveCostDelta == 0 && modifier.GrantImmortality && !modifier.ProhibitRecover && !modifier.SwitchPowerLife && !modifier.GrantStealth && !modifier.GrantTrueSight && !modifier.RemovePride
}

// isClassBonusStaticModifier 驗證 Weapon 職業加成只提供 +1 Power，Bulwark 可同時增加 2 點 wield Reserve 費用。
// 輸入為靜態能力 modifier；輸出為是否符合 Weapon class bonus 形狀，副作用為零。
func isClassBonusStaticModifier(modifier continuousModifier) bool {
	return modifier.SetPride == nil &&
		modifier.SetPower == nil &&
		modifier.SetLife == nil &&
		modifier.PowerDelta == 1 &&
		modifier.LifeDelta == 0 &&
		(modifier.ReserveCostDelta == 0 || modifier.ReserveCostDelta == 2) &&
		!modifier.GrantImmortality &&
		!modifier.ProhibitRecover &&
		!modifier.SwitchPowerLife &&
		!modifier.GrantStealth &&
		!modifier.GrantTrueSight &&
		!modifier.RemovePride
}

// isArthurRestedAlliesModifier 驗證 Arthur 的 static 只向其他受控 Ally 提供 +1 Power。
// 輸入為 continuous modifier；輸出為是否符合 Arthur 的已支援 payload，副作用為零。
func isArthurRestedAlliesModifier(modifier continuousModifier) bool {
	return isCardistryModifier(modifier) && modifier.PowerDelta == 1 && modifier.LifeDelta == 0
}

// isNoireStealthModifier 驗證 Noire 的 static 只授予 Stealth。
// 輸入為 continuous modifier；輸出為是否符合 Noire 的已支援 payload，副作用為零。
func isNoireStealthModifier(modifier continuousModifier) bool {
	return modifier.SetPride == nil && modifier.SetPower == nil && modifier.SetLife == nil && modifier.PowerDelta == 0 && modifier.LifeDelta == 0 && modifier.ReserveCostDelta == 0 && !modifier.GrantImmortality && !modifier.ProhibitRecover && !modifier.SwitchPowerLife && modifier.GrantStealth && !modifier.GrantTrueSight && !modifier.RemovePride
}

// isHeatedVengeanceModifier 驗證 Heated Vengeance 的 static 只提供 +3 Power。
// 輸入為 continuous modifier；輸出為是否符合 Heated Vengeance 的已支援 payload，副作用為零。
func isHeatedVengeanceModifier(modifier continuousModifier) bool {
	return isCardistryModifier(modifier) && modifier.PowerDelta == 3 && modifier.LifeDelta == 0
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
		if value.Constant < 0 || value.Left != nil || value.Right != nil || value.Selector != "" || value.Player != "" || value.Multiplier != 0 {
			return fmt.Errorf("invalid constant payload")
		}
	case valueAdd:
		if value.Left == nil || value.Right == nil || value.Constant != 0 || value.Selector != "" || value.Player != "" || value.Multiplier != 0 {
			return fmt.Errorf("add requires exactly two operands")
		}
		if err := validateValueNode(value.Left, path, depth+1); err != nil {
			return fmt.Errorf("left: %w", err)
		}
		if err := validateValueNode(value.Right, path, depth+1); err != nil {
			return fmt.Errorf("right: %w", err)
		}
	case valueCount, valueSumPrintedReserveCosts, valueDistinctPrintedCosts:
		if value.Selector != selectorControlledSuited || value.Player != referenceController || value.Constant != 0 || value.Left != nil || value.Right != nil || value.Multiplier != 0 {
			return fmt.Errorf("%s requires controlled-suited selector and controller reference", value.Kind)
		}
	case valueThreshold:
		if value.Left == nil || value.Right != nil || value.Constant != 0 || value.Selector != "" || value.Player != "" || value.Multiplier <= 0 {
			return fmt.Errorf("threshold requires one operand and positive multiplier")
		}
		if err := validateValueNode(value.Left, path, depth+1); err != nil {
			return fmt.Errorf("threshold input: %w", err)
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
				TargetBinding:   effect.damage.targetBinding,
				Value:           &amount,
			})
		case effectKindChooseCard:
			operations = append(operations, effectOperation{
				Kind:          effectOperationChoose,
				CardSelection: effect.chooseCard.selection,
				Binding:       effect.chooseCard.binding,
				CanPass:       effect.chooseCard.canPass,
			})
		case effectKindChooseObject:
			operations = append(operations, effectOperation{
				Kind:     effectOperationChoose,
				Selector: effect.chooseObject.selector,
				Binding:  effect.chooseObject.binding,
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
			amount := cloneValueExpression(effect.counter.amount)
			operations = append(operations, effectOperation{
				Kind:          effectOperationCounter,
				TargetBinding: effect.counter.targetBinding,
				Counter:       effect.counter.counter,
				Value:         &amount,
			})
		case effectKindModifier:
			layer := effectLayerModifier
			if effect.modifier.duration == modifierDurationEndOfNextTurn {
				layer = effectLayerAbility
			}
			operations = append(operations, effectOperation{
				Kind:                    effectOperationContinuousModifier,
				Duration:                effect.modifier.duration,
				RequiresRetargetSuccess: effect.modifier.requiresRetargetSuccess,
				ContinuousEffect: continuousEffect{
					Scope:     effectScopeObject,
					Layer:     layer,
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
				Kind: effectOperationMoveSourceToGraveyard,
			})
		case effectKindBanishSource:
			operations = append(operations, effectOperation{
				Kind: effectOperationBanishSource,
			})
		case effectKindRestSource:
			operations = append(operations, effectOperation{
				Kind: effectOperationRestSource,
			})
		case effectKindCardistryDiscount:
			operations = append(operations, effectOperation{
				Kind:   effectOperationCardistryDiscount,
				Amount: 6,
			})
		case effectKindPreventNonCombatDamage:
			operations = append(operations, effectOperation{
				Kind: effectOperationPreventNonCombatDamage,
			})
		case effectKindRestObject:
			operations = append(operations, effectOperation{
				Kind: effectOperationRestObject,
			})
		case effectKindSacrificeAllyPower:
			operations = append(operations, effectOperation{
				Kind:           effectOperationSacrifice,
				TargetBinding:  "sacrifice-target",
				RequiredType:   "ALLY",
				ExcludedObject: "",
			})
			operations = append(operations, effectOperation{
				Kind:     effectOperationContinuousModifier,
				Duration: modifierDurationEndOfTurn,
				ContinuousEffect: continuousEffect{
					Scope:     effectScopeObject,
					Layer:     effectLayerModifier,
					PowerLife: powerLifeModify,
					Modifier: continuousModifier{
						PowerDelta: 2,
					},
				},
			})
		case effectKindGrantTaunt:
			operations = append(operations, effectOperation{
				Kind: effectOperationGrantTaunt,
			})
		}
	}
	return operations
}

// compiledTriggeredAbilities 取得來源卡指定事件的全部已編譯觸發能力。
// 輸入為卡牌實例與事件；輸出依定義順序排列的能力，副作用為零。
func (g *Game) compiledTriggeredAbilities(card cardInstanceID, event eventKind) []compiledAbilityDefinition {
	definition, exists := g.definitions[g.state.Cards[card].Definition]
	if !exists {
		return nil
	}
	abilities := make([]compiledAbilityDefinition, 0)
	for _, ability := range definition.abilities {
		if ability.kind == abilityKindTriggered && ability.trigger != nil && ability.trigger.event == event {
			abilities = append(abilities, ability)
		}
	}
	return abilities
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

// compiledObject 回傳指定場上卡牌已編譯的 Object activated ability。
// 輸入為卡牌實例；輸出為對應能力及是否存在，副作用為零。
func (g *Game) compiledObject(card cardInstanceID) (compiledAbilityDefinition, bool) {
	return g.compiledAbility(card, abilityKindObject)
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
		arthurYoungHeirCardID,
		noireCardID,
		pepperedChefCardID,
		heatedVengeanceCardID,
		rougeCardID,
		spiritOfFireCardID,
		tonorisCardID,
		bulwarkSwordCardID,
		redHareCardID,
		infernalVesselCardID,
		veritaCardID,
		grandCrusadersRingCardID,
		waterResonanceBaubleCardID,
		windResonanceBaubleCardID,
		smokeBombsCardID,
		safeguardAmuletCardID,
		viridianProtectiveTrinketCardID,
		duchessThornesCardID,
	} {
		definition := CardDefinition{
			id: id,
			face: CardFace{
				id: CardFaceID("face:" + string(id) + ":front"),
			},
		}
		abilities, err := compileContentAbilityDefinitions(definition)
		if err != nil {
			return nil, fmt.Errorf("compile replay definition: %w", err)
		}
		definition.abilities = abilities
		definitions[id] = definition
	}
	return definitions, nil
}
