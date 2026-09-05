package game

import (
	"fmt"
	"go-tcg/internal/constants"
	"regexp"
	"strings"
)

type ContentID string

type MechanismID string

type OperationID string

type RulingID string

type AbilityHandler func()

type cardRegistration struct {
	ID           CardID
	Status       constants.SupportStatus
	Dependencies []ContentID
}

type faceRegistration struct {
	ID        CardFaceID
	CardID    CardID
	Status    constants.SupportStatus
	Behaviors []string
}

type abilityRegistration struct {
	ID           AbilitySlotID
	FaceID       CardFaceID
	Status       constants.SupportStatus
	Handler      AbilityHandler
	Mechanisms   []MechanismID
	Operations   []OperationID
	Dependencies []ContentID
	Rulings      []RulingID
}

type contentRegistration struct {
	ID           ContentID
	Status       constants.SupportStatus
	Dependencies []ContentID
}

type supportRegistration[T ~string] struct {
	ID     T
	Status constants.SupportStatus
}

type mechanismRegistration struct {
	ID         MechanismID
	Status     constants.SupportStatus
	Operations []OperationID
	Rulings    []RulingID
}

type rulingRegistration struct {
	ID     RulingID
	Status constants.RulingStatus
}

type registrySpec struct {
	cards      []cardRegistration
	faces      []faceRegistration
	abilities  []abilityRegistration
	contents   []contentRegistration
	mechanisms []mechanismRegistration
	operations []supportRegistration[OperationID]
	rulings    []rulingRegistration
}

type contentRegistry struct {
	cards      map[CardID]cardRegistration
	faces      map[CardFaceID]faceRegistration
	abilities  map[AbilitySlotID]abilityRegistration
	contents   map[ContentID]contentRegistration
	mechanisms map[MechanismID]mechanismRegistration
	operations map[OperationID]supportRegistration[OperationID]
	rulings    map[RulingID]rulingRegistration
}

var semanticKeyPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func buildRegistry(spec registrySpec) (contentRegistry, error) {
	registry := contentRegistry{
		cards:      make(map[CardID]cardRegistration, len(spec.cards)),
		faces:      make(map[CardFaceID]faceRegistration, len(spec.faces)),
		abilities:  make(map[AbilitySlotID]abilityRegistration, len(spec.abilities)),
		contents:   make(map[ContentID]contentRegistration, len(spec.contents)),
		mechanisms: make(map[MechanismID]mechanismRegistration, len(spec.mechanisms)),
		operations: make(map[OperationID]supportRegistration[OperationID], len(spec.operations)),
		rulings:    make(map[RulingID]rulingRegistration, len(spec.rulings)),
	}

	for _, registration := range spec.cards {
		if registration.ID == "" {
			return contentRegistry{}, fmt.Errorf("invalid Card ID %q", registration.ID)
		}
		if err := validateSupportStatus(registration.Status); err != nil {
			return contentRegistry{}, fmt.Errorf("card %q: %w", registration.ID, err)
		}
		if _, exists := registry.cards[registration.ID]; exists {
			return contentRegistry{}, fmt.Errorf("duplicate Card ID %q", registration.ID)
		}
		registry.cards[registration.ID] = registration
	}

	for _, registration := range spec.faces {
		if _, exists := registry.cards[registration.CardID]; !exists {
			return contentRegistry{}, fmt.Errorf("CardFace %q has unknown Card parent %q", registration.ID, registration.CardID)
		}
		if !validFaceID(registration.ID, registration.CardID) {
			return contentRegistry{}, fmt.Errorf("invalid CardFace ID %q", registration.ID)
		}
		if err := validateSupportStatus(registration.Status); err != nil {
			return contentRegistry{}, fmt.Errorf("face %q: %w", registration.ID, err)
		}
		if _, exists := registry.faces[registration.ID]; exists {
			return contentRegistry{}, fmt.Errorf("duplicate CardFace ID %q", registration.ID)
		}
		if err := validateBehaviorKeys(registration); err != nil {
			return contentRegistry{}, err
		}
		registry.faces[registration.ID] = registration
	}

	for _, registration := range spec.abilities {
		if _, exists := registry.faces[registration.FaceID]; !exists {
			return contentRegistry{}, fmt.Errorf("Ability Slot %q has unknown CardFace parent %q", registration.ID, registration.FaceID)
		}
		if !validAbilityID(registration.ID, registration.FaceID) {
			return contentRegistry{}, fmt.Errorf("invalid Ability Slot ID %q", registration.ID)
		}
		if err := validateSupportStatus(registration.Status); err != nil {
			return contentRegistry{}, fmt.Errorf("ability %q: %w", registration.ID, err)
		}
		if registration.Status == constants.Supported && registration.Handler == nil {
			return contentRegistry{}, fmt.Errorf("supported Ability Slot %q has no handler", registration.ID)
		}
		if _, exists := registry.abilities[registration.ID]; exists {
			return contentRegistry{}, fmt.Errorf("duplicate Ability Slot ID %q", registration.ID)
		}
		registry.abilities[registration.ID] = registration
	}

	if err := validateBehaviorSlots(registry); err != nil {
		return contentRegistry{}, err
	}
	if err := addSupportRegistrations(&registry, spec); err != nil {
		return contentRegistry{}, err
	}
	if err := validateRegistryReferences(registry); err != nil {
		return contentRegistry{}, err
	}
	return registry, nil
}

func validateSupportStatus(status constants.SupportStatus) error {
	if status != constants.Supported && status != constants.Unsupported {
		return fmt.Errorf("invalid support status %q", status)
	}
	return nil
}

func validFaceID(id CardFaceID, cardID CardID) bool {
	prefix := "face:" + string(cardID) + ":"
	if !strings.HasPrefix(string(id), prefix) {
		return false
	}
	return semanticKeyPattern.MatchString(strings.TrimPrefix(string(id), prefix))
}

func validAbilityID(id AbilitySlotID, faceID CardFaceID) bool {
	faceParts := strings.Split(string(faceID), ":")
	if len(faceParts) != 3 {
		return false
	}
	prefix := "ability:" + faceParts[1] + ":" + faceParts[2] + ":"
	if !strings.HasPrefix(string(id), prefix) {
		return false
	}
	return semanticKeyPattern.MatchString(strings.TrimPrefix(string(id), prefix))
}

func validateBehaviorKeys(face faceRegistration) error {
	seen := make(map[string]struct{}, len(face.Behaviors))
	for _, behavior := range face.Behaviors {
		if !semanticKeyPattern.MatchString(behavior) {
			return fmt.Errorf("CardFace %q has invalid behavior key %q", face.ID, behavior)
		}
		if _, exists := seen[behavior]; exists {
			return fmt.Errorf("CardFace %q repeats behavior key %q", face.ID, behavior)
		}
		seen[behavior] = struct{}{}
	}
	return nil
}

func validateBehaviorSlots(registry contentRegistry) error {
	for _, face := range registry.faces {
		for _, behavior := range face.Behaviors {
			id := abilitySlotID(
				face.ID,
				behavior,
			)
			if _, exists := registry.abilities[id]; !exists {
				return fmt.Errorf("CardFace %q behavior %q has no Ability Slot", face.ID, behavior)
			}
		}
	}
	for _, ability := range registry.abilities {
		face := registry.faces[ability.FaceID]
		prefix := string(
			abilitySlotID(
				face.ID,
				"",
			),
		)
		behavior := strings.TrimPrefix(string(ability.ID), prefix)
		if !contains(face.Behaviors, behavior) {
			return fmt.Errorf("Ability Slot %q has no rules-bearing behavior", ability.ID)
		}
	}
	return nil
}

func addSupportRegistrations(registry *contentRegistry, spec registrySpec) error {
	for _, registration := range spec.contents {
		if err := validateSupportStatus(registration.Status); err != nil {
			return fmt.Errorf("content %q: %w", registration.ID, err)
		}
		if _, exists := registry.contents[registration.ID]; exists {
			return fmt.Errorf("duplicate Content ID %q", registration.ID)
		}
		registry.contents[registration.ID] = registration
	}
	for _, registration := range spec.mechanisms {
		if err := validateSupportStatus(registration.Status); err != nil {
			return fmt.Errorf("mechanism %q: %w", registration.ID, err)
		}
		if _, exists := registry.mechanisms[registration.ID]; exists {
			return fmt.Errorf("duplicate Mechanism ID %q", registration.ID)
		}
		registry.mechanisms[registration.ID] = registration
	}
	for _, registration := range spec.operations {
		if err := validateSupportStatus(registration.Status); err != nil {
			return fmt.Errorf("operation %q: %w", registration.ID, err)
		}
		if _, exists := registry.operations[registration.ID]; exists {
			return fmt.Errorf("duplicate Operation ID %q", registration.ID)
		}
		registry.operations[registration.ID] = registration
	}
	for _, registration := range spec.rulings {
		if registration.Status != constants.RulingResolved && registration.Status != constants.RulingApproved && registration.Status != constants.RulingPending {
			return fmt.Errorf("ruling %q has invalid status %q", registration.ID, registration.Status)
		}
		if _, exists := registry.rulings[registration.ID]; exists {
			return fmt.Errorf("duplicate Ruling ID %q", registration.ID)
		}
		registry.rulings[registration.ID] = registration
	}
	return nil
}

func validateRegistryReferences(registry contentRegistry) error {
	for _, registration := range registry.cards {
		if err := validateContentDependencies(registration.Dependencies, registry); err != nil {
			return err
		}
	}
	for _, registration := range registry.contents {
		if err := validateContentDependencies(registration.Dependencies, registry); err != nil {
			return err
		}
	}
	for _, registration := range registry.abilities {
		if err := validateAbilityReferences(registration, registry); err != nil {
			return err
		}
	}
	for _, registration := range registry.mechanisms {
		for _, operationID := range registration.Operations {
			if _, exists := registry.operations[operationID]; !exists {
				return fmt.Errorf("Mechanism %q has unknown Operation %q", registration.ID, operationID)
			}
		}
		for _, rulingID := range registration.Rulings {
			if _, exists := registry.rulings[rulingID]; !exists {
				return fmt.Errorf("Mechanism %q has unknown Ruling %q", registration.ID, rulingID)
			}
		}
	}
	return nil
}

func validateAbilityReferences(registration abilityRegistration, registry contentRegistry) error {
	for _, mechanismID := range registration.Mechanisms {
		if _, exists := registry.mechanisms[mechanismID]; !exists {
			return fmt.Errorf("Ability Slot %q has unknown Mechanism %q", registration.ID, mechanismID)
		}
	}
	for _, operationID := range registration.Operations {
		if _, exists := registry.operations[operationID]; !exists {
			return fmt.Errorf("Ability Slot %q has unknown Operation %q", registration.ID, operationID)
		}
	}
	if err := validateContentDependencies(registration.Dependencies, registry); err != nil {
		return fmt.Errorf("Ability Slot %q: %w", registration.ID, err)
	}
	for _, rulingID := range registration.Rulings {
		if _, exists := registry.rulings[rulingID]; !exists {
			return fmt.Errorf("Ability Slot %q has unknown Ruling %q", registration.ID, rulingID)
		}
	}
	return nil
}

func validateContentDependencies(dependencies []ContentID, registry contentRegistry) error {
	for _, dependencyID := range dependencies {
		dependency := string(dependencyID)
		if strings.HasPrefix(
			dependency,
			"card:",
		) {
			cardIDValue := strings.TrimPrefix(
				dependency,
				"card:",
			)
			cardID := CardID(cardIDValue)
			if _, exists := registry.cards[cardID]; !exists {
				return fmt.Errorf("unknown Content dependency %q", dependencyID)
			}
			continue
		}
		if _, exists := registry.contents[dependencyID]; !exists {
			return fmt.Errorf("unknown Content dependency %q", dependencyID)
		}
	}
	return nil
}
