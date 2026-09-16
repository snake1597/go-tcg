package game

import (
	"errors"
	"go-tcg/internal/constants"
	"go-tcg/internal/model"
	"path/filepath"
	"strings"
	"testing"
)

// standardGameConfiguration 建立正式固定牌組測試使用的雙方玩家與 repository 路徑。
// 輸入為無；輸出為可建立單局的 StandardGameConfig，副作用為零。
func standardGameConfiguration() StandardGameConfig {
	repositoryRoot := filepath.Join(
		"..",
		"..",
	)
	return StandardGameConfig{
		Players: [2]*model.Player{
			model.PlayerOne,
			model.PlayerTwo,
		},
		RepositoryRoot: repositoryRoot,
	}
}

func TestProductionSupportSetIsCompleteAndClosed(t *testing.T) {
	registry, err := productionRegistry()
	if err != nil {
		t.Fatalf("productionRegistry() error = %v", err)
	}
	closure, diagnostics := evaluateSupportSet(fixedStandardDeck(), registry)

	if got := len(registry.cards); got != 32 {
		t.Fatalf("registered cards = %d, want 32", got)
	}
	if got := len(registry.faces); got != 32 {
		t.Fatalf("registered faces = %d, want 32", got)
	}
	if got := len(registry.abilities); got != 49 {
		t.Fatalf("registered ability slots = %d, want 49", got)
	}
	if !closure.contents[ContentID("runtime:copied-action")] {
		t.Fatal("Support Set does not recursively include runtime:copied-action")
	}
	if len(diagnostics) != 0 {
		t.Fatalf("complete production Support Set diagnostics = %+v, want none", diagnostics)
	}
}

// TestNewStandardGameRejectsMissingRequiredRegistration 驗證正式建局入口對每種可達支援節點的缺失皆回傳具體 gate 診斷。
// 輸入為移除必要節點的 production registry；輸出為 GateError 內對應節點的診斷，副作用為零。
func TestNewStandardGameRejectsMissingRequiredRegistration(t *testing.T) {
	configuration := standardGameConfiguration()
	tests := []struct {
		name       string
		remove     func(*contentRegistry)
		diagnostic GateDiagnostic
	}{
		{
			name: "content",
			remove: func(registry *contentRegistry) {
				delete(registry.contents, ContentID("runtime:copied-action"))
			},
			diagnostic: GateDiagnostic{
				Kind: constants.GateContent,
				ID:   "runtime:copied-action",
			},
		},
		{
			name: "ability",
			remove: func(registry *contentRegistry) {
				delete(registry.abilities, AbilitySlotID("ability:qzv380ujf5:front:cardistry-copy-action"))
			},
			diagnostic: GateDiagnostic{
				Kind: constants.GateAbility,
				ID:   "ability:qzv380ujf5:front:cardistry-copy-action",
			},
		},
		{
			name: "mechanism",
			remove: func(registry *contentRegistry) {
				delete(registry.mechanisms, MechanismID("MEC-009"))
			},
			diagnostic: GateDiagnostic{
				Kind: constants.GateMechanism,
				ID:   "MEC-009",
			},
		},
		{
			name: "operation",
			remove: func(registry *contentRegistry) {
				delete(registry.operations, OperationID("copy-object"))
			},
			diagnostic: GateDiagnostic{
				Kind: constants.GateOperation,
				ID:   "copy-object",
			},
		},
		{
			name: "ruling",
			remove: func(registry *contentRegistry) {
				delete(registry.rulings, RulingID("RUL-002"))
			},
			diagnostic: GateDiagnostic{
				Kind: constants.GateRuling,
				ID:   "RUL-002",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			registry, err := productionRegistry()
			if err != nil {
				t.Fatalf("productionRegistry() error = %v", err)
			}
			test.remove(&registry)
			game, err := newStandardGameWithRegistry(configuration, registry)
			if game != nil {
				t.Fatal("newStandardGameWithRegistry() returned a game with missing support")
			}
			var gateError *GateError
			if !errors.As(err, &gateError) {
				t.Fatalf("newStandardGameWithRegistry() error = %v, want GateError", err)
			}
			assertDiagnostic(t, gateError.Diagnostics, test.diagnostic)
		})
	}
}

// TestNewStandardGameRejectsMissingCardOrFace 驗證 immutable card data 與 production registry 不一致時，建局前回傳具體錯誤。
// 輸入為缺少必要 card 或 face 的 registry；輸出為包含不一致原因的錯誤，副作用為零。
func TestNewStandardGameRejectsMissingCardOrFace(t *testing.T) {
	configuration := standardGameConfiguration()
	tests := []struct {
		name    string
		remove  func(*contentRegistry)
		message string
	}{
		{
			name: "card",
			remove: func(registry *contentRegistry) {
				delete(registry.cards, CardID("qzv380ujf5"))
			},
			message: "orphaned from the production registry",
		},
		{
			name: "face",
			remove: func(registry *contentRegistry) {
				delete(registry.faces, CardFaceID("face:qzv380ujf5:front"))
			},
			message: "missing from the production registry",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			registry, err := productionRegistry()
			if err != nil {
				t.Fatalf("productionRegistry() error = %v", err)
			}
			test.remove(&registry)
			game, err := newStandardGameWithRegistry(configuration, registry)
			if game != nil {
				t.Fatal("newStandardGameWithRegistry() returned a game with missing production content")
			}
			if err == nil || !strings.Contains(err.Error(), test.message) {
				t.Fatalf("newStandardGameWithRegistry() error = %v, want %q", err, test.message)
			}
		})
	}
}

func TestSupportSetRejectsOrphanProductionContent(t *testing.T) {
	spec := validRegistrySpec()
	spec.cards = append(spec.cards, cardRegistration{
		ID:     CardID("orphan"),
		Status: constants.Supported,
	})
	spec.faces = append(spec.faces, faceRegistration{
		ID:        CardFaceID("face:orphan:front"),
		CardID:    CardID("orphan"),
		Status:    constants.Supported,
		Behaviors: []string{},
	})
	spec.mechanisms = []mechanismRegistration{
		{
			ID:     MechanismID("MEC-orphan"),
			Status: constants.Supported,
		},
	}
	spec.operations = []supportRegistration[OperationID]{
		{
			ID:     OperationID("orphan-operation"),
			Status: constants.Supported,
		},
	}
	spec.rulings = []rulingRegistration{
		{
			ID:     RulingID("RUL-orphan"),
			Status: constants.RulingResolved,
		},
	}
	registry, err := buildRegistry(spec)
	if err != nil {
		t.Fatalf("buildRegistry() error = %v", err)
	}
	deck := DeckManifest{
		Version:         constants.FixedDeckVersion,
		CardDataVersion: constants.FixedCardDataVersion,
		MainDeck: DeckSection{
			deckEntry("card-a", 60),
		},
		MaterialDeck:    DeckSection{},
		OutsideGamePool: DeckSection{},
	}
	_, diagnostics := evaluateSupportSet(deck, registry)
	assertDiagnostic(t, diagnostics, GateDiagnostic{
		Kind: constants.GateRegistry,
		ID:   "card:orphan",
	})
	assertDiagnostic(t, diagnostics, GateDiagnostic{
		Kind: constants.GateRegistry,
		ID:   "MEC-orphan",
	})
	assertDiagnostic(t, diagnostics, GateDiagnostic{
		Kind: constants.GateRegistry,
		ID:   "orphan-operation",
	})
	assertDiagnostic(t, diagnostics, GateDiagnostic{
		Kind: constants.GateRegistry,
		ID:   "RUL-orphan",
	})
}

func TestSupportSetReportsUnresolvedRuling(t *testing.T) {
	spec := validRegistrySpec()
	spec.abilities[0].Mechanisms = []MechanismID{
		MechanismID("MEC-test"),
	}
	spec.mechanisms = []mechanismRegistration{
		{
			ID:     MechanismID("MEC-test"),
			Status: constants.Supported,
			Rulings: []RulingID{
				RulingID("RUL-test"),
			},
		},
	}
	spec.rulings = []rulingRegistration{
		{
			ID:     RulingID("RUL-test"),
			Status: constants.RulingPending,
		},
	}
	registry, err := buildRegistry(spec)
	if err != nil {
		t.Fatalf("buildRegistry() error = %v", err)
	}
	deck := DeckManifest{
		Version:         constants.FixedDeckVersion,
		CardDataVersion: constants.FixedCardDataVersion,
		MainDeck: DeckSection{
			deckEntry("card-a", 60),
		},
		MaterialDeck:    DeckSection{},
		OutsideGamePool: DeckSection{},
	}
	_, diagnostics := evaluateSupportSet(deck, registry)
	assertDiagnostic(t, diagnostics, GateDiagnostic{
		Kind: constants.GateRuling,
		ID:   "RUL-test",
	})
}

func TestSupportSetReportsEveryMissingRequirement(t *testing.T) {
	spec := validRegistrySpec()
	spec.faces[0].Behaviors = []string{
		"on-enter",
		"missing-mechanism",
		"missing-operation-ruling",
	}
	spec.abilities = append(spec.abilities,
		abilityRegistration{
			ID: AbilitySlotID(
				"ability:card-a:front:missing-mechanism",
			),
			FaceID: CardFaceID("face:card-a:front"),
			Status: constants.Supported,
			Mechanisms: []MechanismID{
				MechanismID("MEC-missing"),
			},
		},
		abilityRegistration{
			ID: AbilitySlotID(
				"ability:card-a:front:missing-operation-ruling",
			),
			FaceID: CardFaceID("face:card-a:front"),
			Status: constants.Supported,
			Mechanisms: []MechanismID{
				MechanismID("MEC-present"),
			},
		},
	)
	spec.mechanisms = []mechanismRegistration{
		{
			ID:     MechanismID("MEC-missing"),
			Status: constants.Supported,
		},
		{
			ID:     MechanismID("MEC-present"),
			Status: constants.Supported,
			Operations: []OperationID{
				OperationID("test-operation"),
			},
			Rulings: []RulingID{
				RulingID("RUL-test"),
			},
		},
	}
	spec.operations = []supportRegistration[OperationID]{
		{
			ID:     OperationID("test-operation"),
			Status: constants.Supported,
		},
	}
	spec.rulings = []rulingRegistration{
		{
			ID:     RulingID("RUL-test"),
			Status: constants.RulingPending,
		},
	}
	registry, err := buildRegistry(spec)
	if err != nil {
		t.Fatalf("buildRegistry() error = %v", err)
	}
	delete(
		registry.abilities,
		AbilitySlotID("ability:card-a:front:on-enter"),
	)
	delete(
		registry.mechanisms,
		MechanismID("MEC-missing"),
	)
	delete(
		registry.operations,
		OperationID("test-operation"),
	)
	delete(
		registry.rulings,
		RulingID("RUL-test"),
	)

	deck := DeckManifest{
		Version:         constants.FixedDeckVersion,
		CardDataVersion: constants.FixedCardDataVersion,
		MainDeck: DeckSection{
			deckEntry(
				"card-a",
				60,
			),
		},
		MaterialDeck:    DeckSection{},
		OutsideGamePool: DeckSection{},
	}
	_, diagnostics := evaluateSupportSet(deck, registry)
	assertDiagnostic(t, diagnostics, GateDiagnostic{
		Kind: constants.GateAbility,
		ID:   "ability:card-a:front:on-enter",
	})
	assertDiagnostic(t, diagnostics, GateDiagnostic{
		Kind: constants.GateMechanism,
		ID:   "MEC-missing",
	})
	assertDiagnostic(t, diagnostics, GateDiagnostic{
		Kind: constants.GateOperation,
		ID:   "test-operation",
	})
	assertDiagnostic(t, diagnostics, GateDiagnostic{
		Kind: constants.GateRuling,
		ID:   "RUL-test",
	})
}

// TestNewStandardGameCreatesCompleteFixedDeckGame 驗證正式入口在 Support Set 通過後建立完整開局狀態。
// 輸入為固定牌組設定；輸出為包含雙方 zones 的單局，並確認不會回傳 gate 錯誤。
func TestNewStandardGameCreatesCompleteFixedDeckGame(t *testing.T) {
	repositoryRoot := filepath.Join("..", "..")
	configuration := StandardGameConfig{
		Players: [2]*model.Player{
			model.PlayerOne,
			model.PlayerTwo,
		},
		RepositoryRoot: repositoryRoot,
	}
	game, err := NewStandardGame(configuration)
	if err != nil {
		t.Fatalf("NewStandardGame() error = %v", err)
	}
	if game == nil {
		t.Fatal("NewStandardGame() returned nil game")
	}
	for _, player := range configuration.Players {
		zones, exists := game.state.Zones[player.UID]
		if !exists {
			t.Fatalf("NewStandardGame() has no zones for player %q", player.UID)
		}
		if len(zones.Hand) != 7 {
			t.Fatalf("opening hand for player %q = %d, want 7", player.UID, len(zones.Hand))
		}
	}
}

func TestDefinitionRegistryValidationRejectsUnknownCardFace(t *testing.T) {
	repositoryRoot := filepath.Join(
		"..",
		"..",
	)
	definitions, err := loadCardDefinitions(
		filepath.Join(
			repositoryRoot,
			"card",
		),
		filepath.Join(
			repositoryRoot,
			"card-data-manifest.json",
		),
	)
	if err != nil {
		t.Fatalf("loadCardDefinitions() error = %v", err)
	}
	registry, err := productionRegistry()
	if err != nil {
		t.Fatalf("productionRegistry() error = %v", err)
	}
	registry.faces[CardFaceID("face:GjM8b5fxqj:back")] = faceRegistration{
		ID:        CardFaceID("face:GjM8b5fxqj:back"),
		CardID:    CardID("GjM8b5fxqj"),
		Status:    constants.Unsupported,
		Behaviors: []string{},
	}
	if err := validateDefinitionsAgainstRegistry(definitions, registry); err == nil {
		t.Fatal("validateDefinitionsAgainstRegistry() accepted a CardFace absent from immutable card data")
	}
}

func assertDiagnostic(t *testing.T, diagnostics []GateDiagnostic, want GateDiagnostic) {
	t.Helper()
	for _, diagnostic := range diagnostics {
		if diagnostic.Kind == want.Kind && diagnostic.ID == want.ID {
			return
		}
	}
	t.Fatalf("diagnostics do not contain %+v: %+v", want, diagnostics)
}
