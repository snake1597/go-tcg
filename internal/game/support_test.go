package game

import (
	"errors"
	"path/filepath"
	"slices"
	"testing"
)

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
	if len(diagnostics) == 0 {
		t.Fatal("unsupported production content unexpectedly passed the gate")
	}
	if !slices.IsSortedFunc(diagnostics, compareGateDiagnostic) {
		t.Fatalf("gate diagnostics are not sorted: %+v", diagnostics)
	}
	assertDiagnostic(t, diagnostics, GateDiagnostic{
		Kind: GateAbility,
		ID:   "ability:qzv380ujf5:front:cardistry-copy-action",
	})
	assertDiagnostic(t, diagnostics, GateDiagnostic{
		Kind: GateContent,
		ID:   "runtime:copied-action",
	})
	assertDiagnostic(t, diagnostics, GateDiagnostic{
		Kind: GateMechanism,
		ID:   "MEC-009",
	})
	assertDiagnostic(t, diagnostics, GateDiagnostic{
		Kind: GateOperation,
		ID:   "copy-object",
	})
	for _, diagnostic := range diagnostics {
		if diagnostic.Kind == GateRuling {
			t.Fatalf("resolved or approved ruling was reported as missing: %+v", diagnostic)
		}
	}
}

func TestSupportSetRejectsOrphanProductionContent(t *testing.T) {
	spec := validRegistrySpec()
	spec.cards = append(spec.cards, cardRegistration{
		ID:     CardID("orphan"),
		Status: Supported,
	})
	spec.faces = append(spec.faces, faceRegistration{
		ID:        CardFaceID("face:orphan:front"),
		CardID:    CardID("orphan"),
		Status:    Supported,
		Behaviors: []string{},
	})
	registry, err := buildRegistry(spec)
	if err != nil {
		t.Fatalf("buildRegistry() error = %v", err)
	}
	deck := DeckManifest{
		Version:         FixedDeckVersion,
		CardDataVersion: FixedCardDataVersion,
		MainDeck: DeckSection{
			deckEntry("card-a", 60),
		},
		MaterialDeck:    DeckSection{},
		OutsideGamePool: DeckSection{},
	}
	_, diagnostics := evaluateSupportSet(deck, registry)
	assertDiagnostic(t, diagnostics, GateDiagnostic{
		Kind: GateRegistry,
		ID:   "card:orphan",
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
			Status: Supported,
			Rulings: []RulingID{
				RulingID("RUL-test"),
			},
		},
	}
	spec.rulings = []rulingRegistration{
		{
			ID:     RulingID("RUL-test"),
			Status: rulingPending,
		},
	}
	registry, err := buildRegistry(spec)
	if err != nil {
		t.Fatalf("buildRegistry() error = %v", err)
	}
	deck := DeckManifest{
		Version:         FixedDeckVersion,
		CardDataVersion: FixedCardDataVersion,
		MainDeck: DeckSection{
			deckEntry("card-a", 60),
		},
		MaterialDeck:    DeckSection{},
		OutsideGamePool: DeckSection{},
	}
	_, diagnostics := evaluateSupportSet(deck, registry)
	assertDiagnostic(t, diagnostics, GateDiagnostic{
		Kind: GateRuling,
		ID:   "RUL-test",
	})
}

func TestNewStandardGameReturnsCompleteGateFailure(t *testing.T) {
	repositoryRoot := filepath.Join("..", "..")
	configuration := StandardGameConfig{
		Players: [2]PlayerID{
			PlayerID("player-1"),
			PlayerID("player-2"),
		},
		RepositoryRoot: repositoryRoot,
	}
	game, err := NewStandardGame(configuration)
	if game != nil {
		t.Fatal("NewStandardGame() returned a game with unsupported content")
	}
	var gateError *GateError
	if !errors.As(err, &gateError) {
		t.Fatalf("NewStandardGame() error = %v, want GateError", err)
	}
	if len(gateError.Diagnostics) < 49 {
		t.Fatalf("gate returned only %d diagnostics, want a complete list", len(gateError.Diagnostics))
	}
	assertDiagnostic(t, gateError.Diagnostics, GateDiagnostic{
		Kind: GateAbility,
		ID:   "ability:LMyKyVC2O9:front:on-enter-draw-seven",
	})
	assertDiagnostic(t, gateError.Diagnostics, GateDiagnostic{
		Kind: GateMechanism,
		ID:   "MEC-001",
	})
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
