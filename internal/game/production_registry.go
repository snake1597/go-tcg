package game

import "go-tcg/internal/constants"

type faceInventory struct {
	CardID     string
	Abilities  []string
	Mechanisms []MechanismID
}

func productionRegistry() (contentRegistry, error) {
	spec := registrySpec{
		contents: []contentRegistration{
			{
				ID:     ContentID("runtime:copied-action"),
				Status: constants.Unsupported,
				Dependencies: []ContentID{
					ContentID("card:iohZMWh5v5"),
					ContentID("card:gt2zqtgs42"),
					ContentID("card:28bjn8g50v"),
				},
			},
		},
		mechanisms: productionMechanisms(),
		rulings: []rulingRegistration{
			{
				ID:     RulingID("RUL-001"),
				Status: constants.RulingApproved,
			},
			{
				ID:     RulingID("RUL-002"),
				Status: constants.RulingResolved,
			},
			{
				ID:     RulingID("RUL-003"),
				Status: constants.RulingResolved,
			},
		},
	}

	seenCards := make(map[CardID]struct{})
	deck := fixedStandardDeck()
	sections := []DeckSection{
		deck.MainDeck,
		deck.MaterialDeck,
		deck.OutsideGamePool,
	}
	for _, section := range sections {
		for _, entry := range section {
			if _, exists := seenCards[entry.CardID]; exists {
				continue
			}
			seenCards[entry.CardID] = struct{}{}
			spec.cards = append(spec.cards, cardRegistration{
				ID:     entry.CardID,
				Status: constants.Unsupported,
			})
		}
	}

	for _, inventory := range productionFaceInventory() {
		faceID := CardFaceID("face:" + inventory.CardID + ":front")
		spec.faces = append(spec.faces, faceRegistration{
			ID:        faceID,
			CardID:    CardID(inventory.CardID),
			Status:    constants.Unsupported,
			Behaviors: append([]string(nil), inventory.Abilities...),
		})
		for _, key := range inventory.Abilities {
			registration := abilityRegistration{
				ID:         AbilitySlotID("ability:" + inventory.CardID + ":front:" + key),
				FaceID:     faceID,
				Status:     constants.Unsupported,
				Mechanisms: append([]MechanismID(nil), inventory.Mechanisms...),
			}
			if registration.ID == AbilitySlotID("ability:qzv380ujf5:front:cardistry-copy-action") {
				registration.Dependencies = []ContentID{
					ContentID("runtime:copied-action"),
				}
			}
			spec.abilities = append(spec.abilities, registration)
		}
	}

	seenOperations := make(map[OperationID]struct{})
	for _, mechanism := range spec.mechanisms {
		for _, operationID := range mechanism.Operations {
			if _, exists := seenOperations[operationID]; exists {
				continue
			}
			seenOperations[operationID] = struct{}{}
			spec.operations = append(spec.operations, supportRegistration[OperationID]{
				ID:     operationID,
				Status: constants.Unsupported,
			})
		}
	}
	return buildRegistry(spec)
}

func productionFaceInventory() []faceInventory {
	return []faceInventory{
		face("GjM8b5fxqj", []string{"on-enter-rest-immortality", "rested-allies-power"}, "MEC-004", "MEC-005", "MEC-011"),
		face("iohZMWh5v5", []string{"sacrifice-weapon-cost", "deal-four"}, "MEC-003", "MEC-006", "MEC-010"),
		face("8kmoi0a5uh", []string{"class-bonus-power", "wield-payment"}, "MEC-005", "MEC-007"),
		face("qzv380ujf5", []string{"kindle-six", "cardistry-copy-action"}, "MEC-003", "MEC-008", "MEC-009", "MEC-011"),
		face("gt2zqtgs42", []string{"deal-two-recover-lock"}, "MEC-003", "MEC-006", "MEC-007"),
		face("i9hf5lhl5f", []string{"cardistry-power-five", "floating-memory"}, "MEC-005", "MEC-008", "MEC-011"),
		face("xgax8bbjqj", []string{"cardistry-memory-deploy"}, "MEC-008", "MEC-010"),
		face("8bolq2y5qp", []string{"cardistry-memory-draw"}, "MEC-008", "MEC-010"),
		face("2gv7DC0KID", []string{"divine-relic", "banish-draw"}, "MEC-003", "MEC-010", "MEC-011"),
		face("td460e8ig0", []string{"damaged-champion-power", "on-attack-self-damage"}, "MEC-004", "MEC-005", "MEC-007"),
		face("chsbalegbs", []string{"class-bonus-power", "on-wield-self-damage"}, "MEC-004", "MEC-005", "MEC-007"),
		face("vgWgu1DUYv", []string{"replace-recover"}, "MEC-006"),
		face("wbjc9t8ycp", []string{"suited-stealth", "on-enter-suited-counters"}, "MEC-004", "MEC-005", "MEC-007", "MEC-010", "MEC-011"),
		face("lcy0lw1veb", []string{"on-enter-sacrifice-power"}, "MEC-004", "MEC-005", "MEC-010"),
		face("5du8f077ua", []string{"pride-three", "unique-human-removes-pride", "granted-on-attack-filter"}, "MEC-004", "MEC-005", "MEC-007", "MEC-010", "MEC-011"),
		face("h68dr63eo5", []string{"on-enter-suited-threshold-damage"}, "MEC-004", "MEC-006", "MEC-010"),
		face("yj2rJBREH8", []string{"banish-prevent-noncombat"}, "MEC-003", "MEC-006", "MEC-010"),
		face("ScGcOmkoQt", []string{"banish-stealth-draw"}, "MEC-003", "MEC-005", "MEC-007", "MEC-010"),
		face("LMyKyVC2O9", []string{"on-enter-draw-seven"}, "MEC-001", "MEC-004", "MEC-010"),
		face("28bjn8g50v", []string{"suited-count-damage"}, "MEC-003", "MEC-007", "MEC-008"),
		face("bEXmm4rKOs", []string{"hindered", "cardistry-trigger-power-true-sight", "cardistry-discount"}, "MEC-003", "MEC-004", "MEC-005", "MEC-008", "MEC-010", "MEC-011"),
		face("1db8hz4prm", []string{"cardistry-draw-discard"}, "MEC-008", "MEC-010"),
		face("o09csnorqv", []string{"cardistry-life-two"}, "MEC-005", "MEC-008"),
		face("zb14m4c8lj", []string{"on-enter-taunt"}, "MEC-002", "MEC-004", "MEC-005", "MEC-007", "MEC-011"),
		face("w7g91ru45w", []string{"class-bonus-discount", "retarget-attack-buff"}, "MEC-003", "MEC-005", "MEC-007"),
		face("rufki4o41y", []string{"cardistry-power-two"}, "MEC-005", "MEC-008"),
		face("e8ygl32jef", []string{"fast-activation", "cardistry-buff-counter"}, "MEC-003", "MEC-005", "MEC-008", "MEC-010"),
		face("4qc47amgpp", []string{"suited-alternative-cost", "suited-immortality", "on-death-suited-power"}, "MEC-003", "MEC-004", "MEC-005", "MEC-010", "MEC-011"),
		face("s3572j3oda", []string{"opponent-water-tax"}, "MEC-005"),
		face("dSSRtNnPtw", []string{"conditional-banish-draw"}, "MEC-003", "MEC-010"),
		face("bHGUNMFLg9", []string{"conditional-banish-draw"}, "MEC-003", "MEC-010"),
		face("0mf1ug6yfi", []string{"cardistry-draw"}, "MEC-008", "MEC-010"),
	}
}

func face(cardID string, abilities []string, mechanisms ...string) faceInventory {
	mechanismIDs := make([]MechanismID, 0, len(mechanisms))
	for _, mechanism := range mechanisms {
		mechanismIDs = append(mechanismIDs, MechanismID(mechanism))
	}
	return faceInventory{
		CardID:     cardID,
		Abilities:  abilities,
		Mechanisms: mechanismIDs,
	}
}

func productionMechanisms() []mechanismRegistration {
	return []mechanismRegistration{
		mechanism("MEC-001", []string{"shuffle", "draw", "deckout"}, "RUL-003"),
		mechanism("MEC-002", []string{"materialize", "lineage", "payment"}, "RUL-003"),
		mechanism("MEC-003", []string{"declare", "choose", "target", "pay", "rollback", "resolve", "fizzle"}, "RUL-001", "RUL-002"),
		mechanism("MEC-004", []string{"buffer-trigger", "order-trigger", "grant-ability", "expire", "last-known-information"}, "RUL-001"),
		mechanism("MEC-005", []string{"characteristic-query", "layer-modifier", "cost-modifier", "permission-prohibition"}),
		mechanism("MEC-006", []string{"prevent-damage", "replace-recover", "prohibit-recover", "expire"}),
		mechanism("MEC-007", []string{"attack-declare", "wield", "taunt", "stealth", "true-sight", "retarget", "damage"}),
		mechanism("MEC-008", []string{"distinct-cost-query", "discount", "rest", "banish", "once-per-instance", "activation-event"}, "RUL-001"),
		mechanism("MEC-009", []string{"copy-object", "source-face", "free-optional-activation", "stack-identity"}, "RUL-002"),
		mechanism("MEC-010", []string{"buff-counter", "draw", "draw-to-memory", "discard", "put", "banish", "sacrifice"}),
		mechanism("MEC-011", []string{"keyword-permission-restriction", "deck-validation"}),
	}
}

func mechanism(id string, operations []string, rulings ...string) mechanismRegistration {
	operationIDs := make([]OperationID, 0, len(operations))
	for _, operation := range operations {
		operationIDs = append(operationIDs, OperationID(operation))
	}
	rulingIDs := make([]RulingID, 0, len(rulings))
	for _, ruling := range rulings {
		rulingIDs = append(rulingIDs, RulingID(ruling))
	}
	return mechanismRegistration{
		ID:         MechanismID(id),
		Status:     constants.Unsupported,
		Operations: operationIDs,
		Rulings:    rulingIDs,
	}
}
