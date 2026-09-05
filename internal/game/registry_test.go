package game

import (
	"strings"
	"testing"
)

func TestRegistryAcceptsTypedHierarchicalContent(t *testing.T) {
	spec := validRegistrySpec()
	registry, err := buildRegistry(spec)
	if err != nil {
		t.Fatalf("buildRegistry() error = %v", err)
	}
	if got := len(registry.abilities); got != 1 {
		t.Fatalf("ability count = %d, want 1", got)
	}
}

func TestRegistryRejectsInvalidContent(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*registrySpec)
		message string
	}{
		{
			name: "illegal face key",
			mutate: func(spec *registrySpec) {
				spec.faces[0].ID = CardFaceID("face:card-a:Front")
			},
			message: "invalid CardFace ID",
		},
		{
			name: "unknown face parent",
			mutate: func(spec *registrySpec) {
				spec.faces[0].CardID = CardID("missing")
			},
			message: "unknown Card parent",
		},
		{
			name: "duplicate ability slot",
			mutate: func(spec *registrySpec) {
				spec.abilities = append(spec.abilities, spec.abilities[0])
			},
			message: "duplicate Ability Slot ID",
		},
		{
			name: "missing behavior slot",
			mutate: func(spec *registrySpec) {
				spec.abilities = nil
			},
			message: "has no Ability Slot",
		},
		{
			name: "unsupported status only",
			mutate: func(spec *registrySpec) {
				spec.abilities[0].Status = SupportStatus("partial")
			},
			message: "invalid support status",
		},
		{
			name: "supported slot without handler",
			mutate: func(spec *registrySpec) {
				spec.abilities[0].Status = Supported
				spec.abilities[0].Handler = nil
			},
			message: "has no handler",
		},
		{
			name: "unknown mechanism reference",
			mutate: func(spec *registrySpec) {
				spec.abilities[0].Mechanisms = []MechanismID{
					MechanismID("MEC-missing"),
				}
			},
			message: "unknown Mechanism",
		},
		{
			name: "unknown operation reference",
			mutate: func(spec *registrySpec) {
				spec.abilities[0].Operations = []OperationID{
					OperationID("missing-operation"),
				}
			},
			message: "unknown Operation",
		},
		{
			name: "unknown dependency reference",
			mutate: func(spec *registrySpec) {
				spec.abilities[0].Dependencies = []ContentID{
					ContentID("missing-content"),
				}
			},
			message: "unknown Content dependency",
		},
		{
			name: "unknown ruling reference",
			mutate: func(spec *registrySpec) {
				spec.abilities[0].Rulings = []RulingID{
					RulingID("RUL-missing"),
				}
			},
			message: "unknown Ruling",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			spec := validRegistrySpec()
			test.mutate(&spec)
			_, err := buildRegistry(spec)
			if err == nil || !strings.Contains(err.Error(), test.message) {
				t.Fatalf("buildRegistry() error = %v, want %q", err, test.message)
			}
		})
	}
}

func validRegistrySpec() registrySpec {
	return registrySpec{
		cards: []cardRegistration{
			{
				ID:     CardID("card-a"),
				Status: Supported,
			},
		},
		faces: []faceRegistration{
			{
				ID:        CardFaceID("face:card-a:front"),
				CardID:    CardID("card-a"),
				Status:    Supported,
				Behaviors: []string{"on-enter"},
			},
		},
		abilities: []abilityRegistration{
			{
				ID:     AbilitySlotID("ability:card-a:front:on-enter"),
				FaceID: CardFaceID("face:card-a:front"),
				Status: Supported,
				Handler: func() {
				},
			},
		},
	}
}
