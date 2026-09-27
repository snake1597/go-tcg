package game

import (
	"strings"
	"testing"

	"go-tcg/internal/constants"
	"go-tcg/internal/model"
)

// TestCompileStraightFlareDefinition 驗證 Action 定義的目標、reference 與數值樹皆在建局前檢查。
// 輸入為有效與被破壞的 Go 編寫資料；輸出為已編譯能力或含欄位脈絡的錯誤，副作用為零。
func TestCompileStraightFlareDefinition(t *testing.T) {
	definition := CardDefinition{
		id: straightFlareCardID,
		face: CardFace{
			id: "face:28bjn8g50v:front",
		},
	}
	authored := straightFlareAbilities()
	compiled, err := compileAbilityDefinitions(definition, authored)
	if err != nil {
		t.Fatalf("compile valid Action: %v", err)
	}
	if len(compiled) != 1 || compiled[0].slot != "ability:28bjn8g50v:front:action-damage" || compiled[0].target.kind != selectorUnits {
		t.Fatalf("compiled Action = %#v", compiled)
	}
	if operations := compiled[0].operations(); len(operations) != 1 || operations[0].TargetReference != referenceDeclaredTarget {
		t.Fatalf("compiled target reference = %#v", operations)
	}
	for _, test := range []struct {
		name   string
		change func(*authoredAbilityDefinition)
		field  string
	}{
		{
			name: "invalid selector",
			change: func(ability *authoredAbilityDefinition) {
				ability.target = &targetSelector{
					kind: "hidden-hand",
				}
			},
			field: "target",
		},
		{
			name: "wrong target reference",
			change: func(ability *authoredAbilityDefinition) {
				ability.effects[0].damage.target = referenceController
			},
			field: "damage.target",
		},
		{
			name: "wrong player reference",
			change: func(ability *authoredAbilityDefinition) {
				ability.effects[0].damage.amount.Right.Player = referenceDeclaredTarget
			},
			field: "damage.amount",
		},
		{
			name: "missing operand",
			change: func(ability *authoredAbilityDefinition) {
				ability.effects[0].damage.amount.Right = nil
			},
			field: "damage.amount",
		},
		{
			name: "cyclic value",
			change: func(ability *authoredAbilityDefinition) {
				value := &ability.effects[0].damage.amount
				value.Right = value
			},
			field: "damage.amount",
		},
	} {
		t.Run(
			test.name,
			func(t *testing.T) {
				entry := straightFlareAbilities()[0]
				test.change(&entry)
				_, err := compileAbilityDefinitions(definition, []authoredAbilityDefinition{entry})
				if err == nil || !strings.Contains(err.Error(), test.field) || !strings.Contains(err.Error(), string(definition.id)) {
					t.Fatalf("compile error = %v, want %s with definition context", err, test.field)
				}
			},
		)
	}
	// 編譯後的運算樹與 selector 不可再由作者資料修改。
	authored[0].target.kind = "hidden-hand"
	authored[0].effects[0].damage.amount.Left.Constant = 99
	if compiled[0].target.kind != selectorUnits || compiled[0].effects[0].damage.amount.Left.Constant != 1 {
		t.Fatal("compiled Action aliases authored data")
	}
}

// TestStraightFlareRejectsIllegalTargetAndPreservesPayment 驗證公開選項、非法提交與目標失效後的費用。
// 輸入為 Player View handles；輸出為拒絕時相同 hash 與失效後的公開資訊，副作用為提交一場測試對局。
func TestStraightFlareRejectsIllegalTargetAndPreservesPayment(t *testing.T) {
	game := newActionGameWithSource(t, straightFlareCardID)
	player := model.PlayerOne
	opponent := model.PlayerTwo
	// 只在初始夾具加入可被回應擊倒的 Ally，並讓對手持有 Fast Action。
	zones := game.state.Zones[player.UID]
	var ally cardInstanceID
	for index, card := range zones.MainDeck {
		if game.state.Cards[card].Definition == redHareCardID {
			ally = card
			zones.MainDeck = removeCardAt(zones.MainDeck, index)
			break
		}
	}
	if ally == "" {
		t.Fatal("fixture has no Red Hare")
	}
	game.state.Zones[player.UID] = zones
	game.state.Objects["ally:flare-target"] = fieldObject{
		ID:    "ally:flare-target",
		Card:  ally,
		Owner: player,
		Types: []string{"ALLY"},
	}
	allyObject := game.state.Objects["ally:flare-target"]
	allyObject.Damage = game.characteristicsFor(allyObject.ID).Life - 2
	game.state.Objects[allyObject.ID] = allyObject
	zones = game.state.Zones[opponent.UID]
	var fiery cardInstanceID
	for index, card := range zones.MainDeck {
		if game.state.Cards[card].Definition == fieryInterferenceCardID {
			fiery = card
			zones.MainDeck = removeCardAt(zones.MainDeck, index)
			zones.Hand = append(zones.Hand, card)
			break
		}
	}
	if fiery == "" {
		t.Fatal("fixture has no Fiery Interference")
	}
	for len(zones.Memory) < game.state.Cards[fiery].ReserveCost {
		payment := zones.Hand[0]
		if payment == fiery {
			payment = zones.Hand[1]
		}
		zones.Hand = removeCardAt(zones.Hand, cardIndex(zones.Hand, payment))
		zones.Memory = append(zones.Memory, payment)
	}
	game.state.Zones[opponent.UID] = zones
	game.advanceKnowledgeRevision()
	game.captureReplayInitialState()
	view, err := game.PlayerView(player)
	if err != nil {
		t.Fatalf("PlayerView: %v", err)
	}
	action := actionByCardName(t, view, "Straight Flare")
	handBefore := len(view.Hand)
	paidHandles := reserveHandles(action)
	if action.AbilitySlot != "ability:28bjn8g50v:front:action-damage" {
		t.Fatalf("LegalAction.AbilitySlot = %q", action.AbilitySlot)
	}
	if err := game.Submit(player, Input{
		Revision: view.Revision,
		Action:   action.Handle,
		Reserve:  paidHandles,
	}); err != nil {
		t.Fatalf("declare Straight Flare: %v", err)
	}
	view, err = game.PlayerView(player)
	if err != nil {
		t.Fatalf("PlayerView after declaration: %v", err)
	}
	if view.PendingChoice == nil || len(view.PendingChoice.Options) != 3 {
		t.Fatalf("target options = %#v, want two public champions and one Ally", view.PendingChoice)
	}
	before := game.StateHash()
	if err := game.Submit(player, Input{
		Revision: view.Revision,
		Choice:   "forged-hidden-target",
	}); err == nil {
		t.Fatal("forged target accepted")
	}
	if game.StateHash() != before {
		t.Fatal("rejected target changed state")
	}
	selectVisibleTargetByName(t, game, player, "Red Hare, Unrivaled Stallion")
	submitActionKind(t, game, player, constants.ActionPass)
	view, err = game.PlayerView(opponent)
	if err != nil {
		t.Fatalf("opponent PlayerView: %v", err)
	}
	response := actionByCardName(t, view, "Fiery Interference")
	if err := game.Submit(opponent, Input{
		Revision: view.Revision,
		Action:   response.Handle,
		Reserve:  reserveHandles(response),
	}); err != nil {
		t.Fatalf("respond with Fiery Interference: %v", err)
	}
	selectVisibleTargetByName(t, game, opponent, "Red Hare, Unrivaled Stallion")
	for step := 0; step < 4; step++ {
		view, err = game.PlayerView(player)
		if err != nil {
			t.Fatalf("PlayerView during resolution: %v", err)
		}
		if len(view.EffectsStack) == 0 {
			break
		}
		if view.OpportunityHolder == nil {
			t.Fatal("stack is nonempty without opportunity holder")
		}
		submitActionKind(t, game, view.OpportunityHolder, constants.ActionPass)
	}
	view, err = game.PlayerView(player)
	if err != nil {
		t.Fatalf("PlayerView after fizzle: %v", err)
	}
	if len(view.EffectsStack) != 0 || len(view.Hand) > handBefore-action.ReserveCost {
		t.Fatalf("fizzle outcome: stack=%#v hand=%d, before=%d", view.EffectsStack, len(view.Hand), handBefore)
	}
	for _, card := range view.Hand {
		for _, paid := range paidHandles {
			if card.Handle == paid {
				t.Fatalf("paid reserve card %q returned to hand", paid)
			}
		}
	}
	for _, object := range view.Field {
		if object.CardName == "Red Hare, Unrivaled Stallion" {
			t.Fatalf("response did not remove the declared target: %#v", object)
		}
	}
	if err := game.Replay().Verify(); err != nil {
		t.Fatalf("Replay.Verify after fizzle: %v", err)
	}
}

// selectVisibleTargetByName 只使用 Player View 的待選 handle 提交指定公開卡名。
// 輸入為測試對局、玩家與公開卡名；輸出為提交結果，副作用是完成一次正式目標選擇。
func selectVisibleTargetByName(t *testing.T, game *Game, player *model.Player, name string) {
	t.Helper()
	view, err := game.PlayerView(player)
	if err != nil {
		t.Fatalf("PlayerView for target: %v", err)
	}
	if view.PendingChoice == nil {
		t.Fatal("expected target choice")
	}
	for _, choice := range view.PendingChoice.Choices {
		if choice.CardName == name {
			if err := game.Submit(player, Input{
				Revision: view.Revision,
				Choice:   choice.Handle,
			}); err != nil {
				t.Fatalf("submit target %q: %v", name, err)
			}
			return
		}
	}
	t.Fatalf("target %q not in %#v", name, view.PendingChoice.Choices)
}

// TestStraightFlareUsesReplacementPipeline 驗證已編譯數值效果仍經過正式 replacement 與 replay 路徑。
// 輸入為 Amulet 與 Straight Flare 的 Player View handles；輸出為被防止的公開傷害與可重播結果，副作用為提交兩項行動。
func TestStraightFlareUsesReplacementPipeline(t *testing.T) {
	game := newActionGameWithSource(t, straightFlareCardID)
	player := model.PlayerOne
	_ = replacementObject(t, game, player, safeguardAmuletCardID, "straight-flare-amulet")
	game.advanceKnowledgeRevision()
	game.captureReplayInitialState()
	view, err := game.PlayerView(player)
	if err != nil {
		t.Fatalf("PlayerView: %v", err)
	}
	amulet := actionByCardName(t, view, "Safeguard Amulet")
	if err := game.Submit(player, Input{
		Revision: view.Revision,
		Action:   amulet.Handle,
		Reserve:  reserveHandles(amulet),
	}); err != nil {
		t.Fatalf("activate Amulet: %v", err)
	}
	view, err = game.PlayerView(player)
	if err != nil {
		t.Fatalf("PlayerView after Amulet: %v", err)
	}
	flare := actionByCardName(t, view, "Straight Flare")
	if err := game.Submit(player, Input{
		Revision: view.Revision,
		Action:   flare.Handle,
		Reserve:  reserveHandles(flare),
	}); err != nil {
		t.Fatalf("declare Straight Flare: %v", err)
	}
	selectPendingChoiceSubject(t, game, player, entityID("champion:"+player.UID))
	view, err = game.PlayerView(player)
	if err != nil {
		t.Fatalf("PlayerView on stack: %v", err)
	}
	if len(view.EffectsStack) != 1 || view.EffectsStack[0].AbilitySlot != "ability:28bjn8g50v:front:action-damage" {
		t.Fatalf("EffectsStack = %#v, want compiled ability slot", view.EffectsStack)
	}
	passOpportunityRound(t, game, player)
	view, err = game.PlayerView(player)
	if err != nil {
		t.Fatalf("PlayerView after resolution: %v", err)
	}
	for _, champion := range view.Champions {
		if champion.Owner == player && champion.Damage != 0 {
			t.Fatalf("Amulet did not prevent Straight Flare damage: %#v", champion)
		}
	}
	if err := game.Replay().Verify(); err != nil {
		t.Fatalf("Replay.Verify: %v", err)
	}
}
