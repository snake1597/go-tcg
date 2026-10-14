package game

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// compiledSupportAbility 是由單一 compiled Ability Definition 衍生的 production 支援資料。
// 輸入來自已驗證的 definition；輸出保留其 CardFace、Slot、機制、runtime operation、可執行內容相依及卡面 ruling，副作用為零。
type compiledSupportAbility struct {
	Face         CardFaceID
	Mechanism    abilityDefinitionKind
	Operations   []effectOperationKind
	Dependencies []CardID
	Rulings      []*Rule
}

// compiledSupportSet 是固定牌組可達內容的衍生 Support Set。
// 它只保存已通過 compiler 的 CardFace 與 Ability Slot metadata，不保存另一份手寫 registry。
type compiledSupportSet struct {
	faces     map[CardFaceID]CardID
	abilities map[AbilitySlotID]compiledSupportAbility
}

// deriveSupportSet 從固定牌組可達的已編譯 Card Definitions 產生並驗證 Support Set。
// 輸入為牌組與已完成 content loading 的 definitions；輸出為完整衍生集合或任一缺漏／未知 dependency 的錯誤，副作用為零。
func deriveSupportSet(deck DeckManifest, definitions map[CardID]CardDefinition) (compiledSupportSet, error) {
	if err := validateDeckReferences(deck, definitions); err != nil {
		return compiledSupportSet{}, fmt.Errorf("validate support deck references: %w", err)
	}
	support := compiledSupportSet{
		faces:     make(map[CardFaceID]CardID),
		abilities: make(map[AbilitySlotID]compiledSupportAbility),
	}
	for _, cardID := range reachableDeckCards(deck) {
		definition := definitions[cardID]
		if cardHasRulesText(definition.card) && len(definition.abilities) == 0 {
			return compiledSupportSet{}, fmt.Errorf("Card Definition %q has rules text but no compiled Ability Slots", cardID)
		}
		if _, exists := support.faces[definition.face.id]; exists {
			return compiledSupportSet{}, fmt.Errorf("duplicate CardFace %q", definition.face.id)
		}
		support.faces[definition.face.id] = cardID
		for _, ability := range definition.abilities {
			if ability.supportStatus != supportStatusSupported {
				return compiledSupportSet{}, fmt.Errorf("Ability Slot %q has non-production support status %q", ability.slot, ability.supportStatus)
			}
			if _, exists := support.abilities[ability.slot]; exists {
				return compiledSupportSet{}, fmt.Errorf("duplicate Ability Slot %q", ability.slot)
			}
			metadata, err := deriveAbilitySupport(definition, ability, definitions)
			if err != nil {
				return compiledSupportSet{}, fmt.Errorf("derive Ability Slot %q: %w", ability.slot, err)
			}
			support.abilities[ability.slot] = metadata
		}
	}
	if err := validateSupportDependencyCycles(support); err != nil {
		return compiledSupportSet{}, err
	}
	return support, nil
}

// cardHasRulesText 判斷卡面是否宣告需由可執行定義承接的規則文字。
// 輸入為 immutable Card data；輸出為 effect_raw 是否含非空白文字，副作用為零。
func cardHasRulesText(card Card) bool {
	return card.EffectRaw != nil && strings.TrimSpace(*card.EffectRaw) != ""
}

// reachableDeckCards 回傳三個牌組區段中每種 Card Definition 一次，並以 Card ID 排序。
// 輸入為固定牌組；輸出為可達 Card ID，副作用為零。
func reachableDeckCards(deck DeckManifest) []CardID {
	seen := make(map[CardID]struct{})
	for _, section := range []DeckSection{
		deck.MainDeck,
		deck.MaterialDeck,
		deck.OutsideGamePool,
	} {
		for _, entry := range section {
			seen[entry.CardID] = struct{}{}
		}
	}
	cards := make([]CardID, 0, len(seen))
	for cardID := range seen {
		cards = append(cards, cardID)
	}
	sort.Slice(cards, func(first, second int) bool {
		return cards[first] < cards[second]
	})
	return cards
}

// deriveAbilitySupport 將一個 compiled ability 的語意投影為 Support Set 所需 metadata。
// 輸入為來源 definition、已驗證 ability 與完整 definitions；輸出為 metadata 或未知可執行 dependency 的錯誤，副作用為零。
func deriveAbilitySupport(definition CardDefinition, ability compiledAbilityDefinition, definitions map[CardID]CardDefinition) (compiledSupportAbility, error) {
	metadata := compiledSupportAbility{
		Face:      definition.face.id,
		Mechanism: ability.kind,
		Rulings:   append([]*Rule(nil), definition.card.Rule...),
	}
	for index, ruling := range metadata.Rulings {
		if ruling == nil || strings.TrimSpace(ruling.Description) == "" || strings.TrimSpace(ruling.DateAdded) == "" {
			return compiledSupportAbility{}, fmt.Errorf("ruling reference %d is incomplete", index)
		}
	}
	operations := ability.operations()
	if len(operations) != len(ability.effects) {
		return compiledSupportAbility{}, fmt.Errorf("compiled effects contain an unknown or unmapped primitive")
	}
	for _, operation := range operations {
		if !isSupportedOperation(operation.Kind) {
			return compiledSupportAbility{}, fmt.Errorf("unknown runtime operation %q", operation.Kind)
		}
		metadata.Operations = append(metadata.Operations, operation.Kind)
		if operation.Kind != effectOperationCopyAction {
			continue
		}
		dependencies, err := copyActionDependencies(operation.CardSelection, definitions)
		if err != nil {
			return compiledSupportAbility{}, err
		}
		for _, dependency := range dependencies {
			metadata.Dependencies = append(metadata.Dependencies, dependency)
		}
	}
	return metadata, nil
}

// copyActionDependencies 從 copy_action 的 card selection 條件找出會被建立為 runtime copy 的固定卡牌。
// 輸入為已編譯選牌條件與已載入 definitions；輸出為依 Card ID 排序的可執行內容相依，副作用為零。
func copyActionDependencies(selection cardSelectionSpec, definitions map[CardID]CardDefinition) ([]CardID, error) {
	dependencies := make([]CardID, 0)
	for cardID, definition := range definitions {
		if !cardMatchesSelection(definition.card, selection) {
			continue
		}
		if _, exists := firstAbilityOfKind(definition.abilities, abilityKindAction); !exists {
			return nil, fmt.Errorf("matching Action %q has no compiled Action Ability Slot", cardID)
		}
		dependencies = append(dependencies, cardID)
	}
	sort.Slice(dependencies, func(first, second int) bool {
		return dependencies[first] < dependencies[second]
	})
	return dependencies, nil
}

// isSupportedOperation 限制 Support Set 僅接受 runtime 已實作且可重播的 typed operations。
// 輸入為 operation kind；輸出為是否在 production primitive 集合中，副作用為零。
func isSupportedOperation(operation effectOperationKind) bool {
	switch operation {
	case effectOperationChoose,
		effectOperationMoveSourceToGraveyard,
		effectOperationRestSource,
		effectOperationBanishSource,
		effectOperationCardistryDiscount,
		effectOperationPreventNonCombatDamage,
		effectOperationDraw,
		effectOperationDrawToMemory,
		effectOperationCounter,
		effectOperationDamage,
		effectOperationContinuousModifier,
		effectOperationDiscard,
		effectOperationDeploy,
		effectOperationPutAllyOnField,
		effectOperationRestGrantImmortality,
		effectOperationSuitedThresholdCounter,
		effectOperationSuitedThresholdDamage,
		effectOperationCopyAction,
		effectOperationSacrificeAllyPower,
		effectOperationRetargetAttack,
		effectOperationGrantTaunt:
		return true
	default:
		return false
	}
}

// validateSupportDependencyCycles 驗證 Ability content dependencies 不會形成遞迴建立鏈。
// 輸入為已衍生 Support Set；輸出為第一個循環錯誤或 nil，副作用為零。
func validateSupportDependencyCycles(support compiledSupportSet) error {
	dependencies := make(map[CardID][]CardID)
	for _, ability := range support.abilities {
		source := support.faces[ability.Face]
		dependencies[source] = append(dependencies[source], ability.Dependencies...)
	}
	visiting := make(map[CardID]bool)
	visited := make(map[CardID]bool)
	var visit func(CardID) error
	visit = func(cardID CardID) error {
		if visiting[cardID] {
			return fmt.Errorf("cyclic content dependency at Card Definition %q", cardID)
		}
		if visited[cardID] {
			return nil
		}
		visiting[cardID] = true
		for _, dependency := range dependencies[cardID] {
			if err := visit(dependency); err != nil {
				return err
			}
		}
		visiting[cardID] = false
		visited[cardID] = true
		return nil
	}
	for cardID := range dependencies {
		if err := visit(cardID); err != nil {
			return err
		}
	}
	return nil
}

// cardMatchesSelection 以 immutable Card data 驗證 card selection 的 type、subtype、element 與 reserve cost 條件。
// 輸入為候選卡面與選牌條件；輸出為是否符合單卡條件，副作用為零。
func cardMatchesSelection(card Card, selection cardSelectionSpec) bool {
	if selection.RequiredType != "" && !containsString(card.Types, selection.RequiredType) {
		return false
	}
	if selection.RequiredSubtype != "" && !containsString(card.Subtypes, selection.RequiredSubtype) {
		return false
	}
	if selection.RequiredElement != "" && !containsString(card.Elements, selection.RequiredElement) {
		return false
	}
	if len(selection.RequiredElements) > 0 && !containsAnyString(card.Elements, selection.RequiredElements) {
		return false
	}
	if selection.MaximumReserveCost == 0 {
		return true
	}
	if card.Cost == nil || card.Cost.Type != "reserve" {
		return false
	}
	reserveCost, err := strconv.Atoi(card.Cost.Value)
	return err == nil && reserveCost <= selection.MaximumReserveCost
}

// firstAbilityOfKind 回傳 definition 中第一個指定 kind 的 compiled ability。
// 輸入為能力集合與目標 kind；輸出為對應 ability 與是否存在，副作用為零。
func firstAbilityOfKind(abilities []compiledAbilityDefinition, kind abilityDefinitionKind) (compiledAbilityDefinition, bool) {
	for _, ability := range abilities {
		if ability.kind == kind {
			return ability, true
		}
	}
	return compiledAbilityDefinition{}, false
}

// containsOperation 判定衍生 metadata 是否包含指定 runtime operation。
// 輸入為 operation 集合與目標 operation；輸出為是否存在，副作用為零。
func containsOperation(operations []effectOperationKind, want effectOperationKind) bool {
	for _, operation := range operations {
		if operation == want {
			return true
		}
	}
	return false
}

// containsSupportDependency 判定衍生 metadata 是否包含指定 Card Definition dependency。
// 輸入為 dependency 集合與目標 Card ID；輸出為是否存在，副作用為零。
func containsSupportDependency(dependencies []CardID, want CardID) bool {
	for _, dependency := range dependencies {
		if dependency == want {
			return true
		}
	}
	return false
}
