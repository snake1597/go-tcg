package game

import (
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

type GateKind string

const (
	GateAbility   GateKind = "ability-slot"
	GateContent   GateKind = "content"
	GateMechanism GateKind = "mechanism"
	GateOperation GateKind = "operation"
	GateRegistry  GateKind = "registry"
	GateRuling    GateKind = "ruling"
)

type GateDiagnostic struct {
	Kind   GateKind
	ID     string
	Reason string
}

type GateError struct {
	Diagnostics []GateDiagnostic
}

func (gateError *GateError) Error() string {
	if len(gateError.Diagnostics) == 0 {
		return "standard game support gate failed"
	}
	parts := make([]string, 0, len(gateError.Diagnostics))
	for _, diagnostic := range gateError.Diagnostics {
		parts = append(parts, fmt.Sprintf("%s %s: %s", diagnostic.Kind, diagnostic.ID, diagnostic.Reason))
	}
	return "standard game support gate failed: " + strings.Join(parts, "; ")
}

type StandardGameConfig struct {
	Players        [2]PlayerID
	RepositoryRoot string
	Seed           uint64
}

type supportClosure struct {
	cards      map[CardID]bool
	faces      map[CardFaceID]bool
	abilities  map[AbilitySlotID]bool
	contents   map[ContentID]bool
	mechanisms map[MechanismID]bool
	operations map[OperationID]bool
	rulings    map[RulingID]bool
}

func NewStandardGame(configuration StandardGameConfig) (*Game, error) {
	if configuration.Players[0] == "" || configuration.Players[1] == "" || configuration.Players[0] == configuration.Players[1] {
		return nil, errors.New("standard game requires two distinct players")
	}
	definitions, err := loadCardDefinitions(
		filepath.Join(configuration.RepositoryRoot, "card"),
		filepath.Join(configuration.RepositoryRoot, "card-data-manifest.json"),
	)
	if err != nil {
		return nil, fmt.Errorf("load fixed card data: %w", err)
	}
	firstDeck := fixedStandardDeck()
	secondDeck := fixedStandardDeck()
	if err := validateFixedStandardDeck(firstDeck, definitions); err != nil {
		return nil, fmt.Errorf("validate first player deck: %w", err)
	}
	if err := validateFixedStandardDeck(secondDeck, definitions); err != nil {
		return nil, fmt.Errorf("validate second player deck: %w", err)
	}
	if err := validateMirroredDecks(firstDeck, secondDeck); err != nil {
		return nil, err
	}
	registry, err := productionRegistry()
	if err != nil {
		return nil, fmt.Errorf("build production registry: %w", err)
	}
	if err := validateDefinitionsAgainstRegistry(definitions, registry); err != nil {
		return nil, err
	}
	_, diagnostics := evaluateSupportSet(firstDeck, registry)
	if len(diagnostics) > 0 {
		return nil, &GateError{
			Diagnostics: diagnostics,
		}
	}

	players := []PlayerID{
		configuration.Players[0],
		configuration.Players[1],
	}
	return &Game{
		players: players,
		state: gameState{
			Revision: 1,
		},
		replay: Replay{
			EngineVersion:   "grand-archive-v1",
			RulesVersion:    "602c917f2f8fd4df7198429a72eb596bf7f647c6",
			CardDataVersion: FixedCardDataVersion,
			DeckVersion:     FixedDeckVersion,
			PRNGVersion:     "not-initialized",
			Seed:            configuration.Seed,
		},
	}, nil
}

func validateDefinitionsAgainstRegistry(definitions map[CardID]CardDefinition, registry contentRegistry) error {
	for cardID, definition := range definitions {
		if _, exists := registry.cards[cardID]; !exists {
			return fmt.Errorf("card definition %q is orphaned from the production registry", cardID)
		}
		if _, exists := registry.faces[definition.Face().ID()]; !exists {
			return fmt.Errorf("CardFace definition %q is missing from the production registry", definition.Face().ID())
		}
	}
	for cardID := range registry.cards {
		if _, exists := definitions[cardID]; !exists {
			return fmt.Errorf("production registry card %q has no card definition", cardID)
		}
	}
	return nil
}

func evaluateSupportSet(deck DeckManifest, registry contentRegistry) (supportClosure, []GateDiagnostic) {
	closure := supportClosure{
		cards:      make(map[CardID]bool),
		faces:      make(map[CardFaceID]bool),
		abilities:  make(map[AbilitySlotID]bool),
		contents:   make(map[ContentID]bool),
		mechanisms: make(map[MechanismID]bool),
		operations: make(map[OperationID]bool),
		rulings:    make(map[RulingID]bool),
	}
	diagnosticSet := make(map[string]GateDiagnostic)
	addDiagnostic := func(kind GateKind, id, reason string) {
		key := string(kind) + "\x00" + id
		diagnosticSet[key] = GateDiagnostic{
			Kind:   kind,
			ID:     id,
			Reason: reason,
		}
	}

	var visitContent func(ContentID)
	var visitCard func(CardID)
	var visitFace func(CardFaceID)
	var visitAbility func(AbilitySlotID)
	var visitMechanism func(MechanismID)
	var visitOperation func(OperationID)
	var visitRuling func(RulingID)

	visitRuling = func(id RulingID) {
		if closure.rulings[id] {
			return
		}
		closure.rulings[id] = true
		registration, exists := registry.rulings[id]
		if !exists {
			addDiagnostic(GateRuling, string(id), "missing from registry")
			return
		}
		if registration.Status == rulingPending {
			addDiagnostic(GateRuling, string(id), "unresolved")
		}
	}
	visitOperation = func(id OperationID) {
		if closure.operations[id] {
			return
		}
		closure.operations[id] = true
		registration, exists := registry.operations[id]
		if !exists {
			addDiagnostic(GateOperation, string(id), "missing from registry")
			return
		}
		if registration.Status == Unsupported {
			addDiagnostic(GateOperation, string(id), "unsupported")
		}
	}
	visitMechanism = func(id MechanismID) {
		if closure.mechanisms[id] {
			return
		}
		closure.mechanisms[id] = true
		registration, exists := registry.mechanisms[id]
		if !exists {
			addDiagnostic(GateMechanism, string(id), "missing from registry")
			return
		}
		if registration.Status == Unsupported {
			addDiagnostic(GateMechanism, string(id), "unsupported")
		}
		for _, operationID := range registration.Operations {
			visitOperation(operationID)
		}
		for _, rulingID := range registration.Rulings {
			visitRuling(rulingID)
		}
	}
	visitAbility = func(id AbilitySlotID) {
		if closure.abilities[id] {
			return
		}
		closure.abilities[id] = true
		registration, exists := registry.abilities[id]
		if !exists {
			addDiagnostic(GateAbility, string(id), "missing from registry")
			return
		}
		if registration.Status == Unsupported {
			addDiagnostic(GateAbility, string(id), "unsupported")
		}
		for _, mechanismID := range registration.Mechanisms {
			visitMechanism(mechanismID)
		}
		for _, operationID := range registration.Operations {
			visitOperation(operationID)
		}
		for _, dependencyID := range registration.Dependencies {
			visitContent(dependencyID)
		}
		for _, rulingID := range registration.Rulings {
			visitRuling(rulingID)
		}
	}
	visitFace = func(id CardFaceID) {
		if closure.faces[id] {
			return
		}
		closure.faces[id] = true
		registration, exists := registry.faces[id]
		if !exists {
			addDiagnostic(GateContent, string(id), "missing from registry")
			return
		}
		if registration.Status == Unsupported {
			addDiagnostic(GateContent, string(id), "unsupported")
		}
		for abilityID, ability := range registry.abilities {
			if ability.FaceID == id {
				visitAbility(abilityID)
			}
		}
	}
	visitCard = func(id CardID) {
		if closure.cards[id] {
			return
		}
		closure.cards[id] = true
		registration, exists := registry.cards[id]
		contentID := "card:" + string(id)
		if !exists {
			addDiagnostic(GateContent, contentID, "missing from registry")
			return
		}
		if registration.Status == Unsupported {
			addDiagnostic(GateContent, contentID, "unsupported")
		}
		for faceID, face := range registry.faces {
			if face.CardID == id {
				visitFace(faceID)
			}
		}
		for _, dependencyID := range registration.Dependencies {
			visitContent(dependencyID)
		}
	}
	visitContent = func(id ContentID) {
		if strings.HasPrefix(string(id), "card:") {
			visitCard(CardID(strings.TrimPrefix(string(id), "card:")))
			return
		}
		if closure.contents[id] {
			return
		}
		closure.contents[id] = true
		registration, exists := registry.contents[id]
		if !exists {
			addDiagnostic(GateContent, string(id), "missing from registry")
			return
		}
		if registration.Status == Unsupported {
			addDiagnostic(GateContent, string(id), "unsupported")
		}
		for _, dependencyID := range registration.Dependencies {
			visitContent(dependencyID)
		}
	}

	sections := []DeckSection{
		deck.MainDeck,
		deck.MaterialDeck,
		deck.OutsideGamePool,
	}
	for _, section := range sections {
		for _, entry := range section {
			visitCard(entry.CardID)
			visitFace(entry.FaceID)
		}
	}
	addOrphanDiagnostics(closure, registry, addDiagnostic)

	diagnostics := make([]GateDiagnostic, 0, len(diagnosticSet))
	for _, diagnostic := range diagnosticSet {
		diagnostics = append(diagnostics, diagnostic)
	}
	sort.Slice(diagnostics, func(first, second int) bool {
		return compareGateDiagnostic(diagnostics[first], diagnostics[second]) < 0
	})
	return closure, diagnostics
}

func addOrphanDiagnostics(closure supportClosure, registry contentRegistry, add func(GateKind, string, string)) {
	for id := range registry.cards {
		if !closure.cards[id] {
			add(GateRegistry, "card:"+string(id), "orphaned formal content")
		}
	}
	for id := range registry.faces {
		if !closure.faces[id] {
			add(GateRegistry, string(id), "orphaned formal content")
		}
	}
	for id := range registry.abilities {
		if !closure.abilities[id] {
			add(GateRegistry, string(id), "orphaned formal content")
		}
	}
	for id := range registry.contents {
		if !closure.contents[id] {
			add(GateRegistry, string(id), "orphaned formal content")
		}
	}
}

func compareGateDiagnostic(first, second GateDiagnostic) int {
	if first.Kind < second.Kind {
		return -1
	}
	if first.Kind > second.Kind {
		return 1
	}
	return strings.Compare(first.ID, second.ID)
}
