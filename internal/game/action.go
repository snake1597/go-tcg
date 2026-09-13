package game

import (
	"fmt"
	"go-tcg/internal/model"
	tcgErrors "go-tcg/internal/tcg_errors"
	"sort"
)

const (
	blazingThrowCardID      CardID = "iohZMWh5v5"
	fieryInterferenceCardID CardID = "gt2zqtgs42"
	straightFlareCardID     CardID = "28bjn8g50v"
)

type declarationStage string

const (
	declarationTarget declarationStage = "target"
	declarationWeapon declarationStage = "weapon"
)

type actionDeclaration struct {
	Controller *model.Player    `json:"controller"`
	Source     cardInstanceID   `json:"source"`
	Target     objectID         `json:"target,omitempty"`
	Stage      declarationStage `json:"stage"`
}

func (g *Game) legalActionCards(player *model.Player) []cardInstanceID {
	if g.state.Knowledge.Declaration != nil {
		return nil
	}
	zones := g.state.Zones[player.UID]
	cards := make([]cardInstanceID, 0, len(zones.Hand))
	for _, card := range zones.Hand {
		if g.canActivateAction(player, card) {
			cards = append(cards, card)
		}
	}
	return cards
}

func (g *Game) canActivateAction(player *model.Player, card cardInstanceID) bool {
	scheduler := g.state.Scheduler
	candidate, exists := g.state.Cards[card]
	if !exists || !samePlayer(candidate.Owner, player) || (!containsString(candidate.Types, "ACTION") && candidate.Definition != veritaCardID) {
		return false
	}
	canUseAlternativeCost := candidate.Definition == veritaCardID && len(g.veritaAlternativeCostCards(player)) > 0
	if !samePlayer(scheduler.OpportunityHolder, player) || (!canUseAlternativeCost && len(g.state.Zones[player.UID].Memory) < g.actionReserveCost(player, card)) {
		return false
	}
	if !candidate.Fast && (!samePlayer(scheduler.TurnPlayer, player) || scheduler.Phase != PhaseMain || len(g.state.EffectsStack) != 0) {
		return false
	}
	if candidate.Definition == blazingThrowCardID && len(g.legalWeapons(player)) == 0 {
		return false
	}
	if candidate.Definition == veritaCardID {
		return true
	}
	if candidate.Definition == trumpSetCardID {
		return g.hasActiveCombat() && len(g.controlledSuitedAllies(player)) > 0
	}
	return candidate.Definition == blazingThrowCardID || candidate.Definition == fieryInterferenceCardID || candidate.Definition == straightFlareCardID
}

func (g *Game) beginActionDeclaration(player *model.Player, card cardInstanceID) error {
	if g.state.Cards[card].Definition == veritaCardID {
		return g.commitAllyActivation(player, card)
	}
	targets := g.legalTargets()
	if g.state.Cards[card].Definition == trumpSetCardID {
		targets = g.controlledSuitedAllies(player)
	}
	if !g.canActivateAction(player, card) || len(targets) == 0 {
		return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, card)
	}
	g.state.Knowledge.Declaration = &actionDeclaration{
		Controller: player,
		Source:     card,
		Stage:      declarationTarget,
	}
	g.setDeclarationChoice(player, targets)
	return nil
}

func (g *Game) commitAllyActivation(player *model.Player, card cardInstanceID) error {
	if !g.canActivateAction(player, card) {
		return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, card)
	}
	if g.state.Cards[card].Definition == veritaCardID {
		if alternativeCost := g.veritaAlternativeCostCards(player); len(alternativeCost) > 0 {
			if err := g.payVeritaAlternativeCost(player, alternativeCost); err != nil {
				return err
			}
			zones := g.state.Zones[player.UID]
			index := cardIndex(zones.Hand, card)
			if index < 0 {
				return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, card)
			}
			zones.Hand = removeCardAt(zones.Hand, index)
			g.state.Zones[player.UID] = zones
			g.putAllyOnField(player, card)
			return nil
		}
	}
	zones := g.state.Zones[player.UID]
	index := cardIndex(zones.Hand, card)
	if index < 0 {
		return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, card)
	}
	zones.Hand = removeCardAt(zones.Hand, index)
	for payment := 0; payment < g.actionReserveCost(player, card); payment++ {
		memoryIndex := int(g.nextRandom() % uint64(len(zones.Memory)))
		zones.Banishment = append(zones.Banishment, zones.Memory[memoryIndex])
		zones.Memory = removeCardAt(zones.Memory, memoryIndex)
	}
	g.state.Zones[player.UID] = zones
	g.putAllyOnField(player, card)
	return nil
}

func (g *Game) submitActionDeclarationChoice(player *model.Player, subject entityID) error {
	declaration := g.state.Knowledge.Declaration
	if declaration == nil || !samePlayer(declaration.Controller, player) {
		return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, subject)
	}
	switch declaration.Stage {
	case declarationTarget:
		target := objectID(subject)
		if !g.isLegalTarget(target) || (g.state.Cards[declaration.Source].Definition == trumpSetCardID && !containsObject(g.controlledSuitedAllies(player), target)) {
			return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, subject)
		}
		declaration.Target = target
		if g.state.Cards[declaration.Source].Definition == blazingThrowCardID {
			declaration.Stage = declarationWeapon
			weapons := g.legalWeapons(player)
			g.setDeclarationChoice(player, weapons)
			return nil
		}
		return g.commitActionDeclaration()
	case declarationWeapon:
		weapon := objectID(subject)
		if !g.isLegalWeapon(player, weapon) {
			return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, subject)
		}
		return g.commitActionDeclarationWithWeapon(weapon)
	default:
		return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, subject)
	}
}

func (g *Game) setDeclarationChoice(player *model.Player, subjects []objectID) {
	options := make(map[ViewHandle]entityID, len(subjects))
	for _, subject := range subjects {
		handle := g.newViewHandle(player, "choice:"+string(subject))
		options[handle] = entityID(subject)
	}
	g.state.Knowledge.Choice = &pendingChoice{
		Actor:   player,
		Options: options,
	}
}

func (g *Game) commitActionDeclarationWithWeapon(weapon objectID) error {
	declaration := g.state.Knowledge.Declaration
	if declaration == nil || !g.isLegalWeapon(declaration.Controller, weapon) || !g.canCommitActionDeclaration(declaration) {
		return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, weapon)
	}
	object := g.state.Objects[weapon]
	delete(g.state.Objects, weapon)
	zones := g.state.Zones[object.Owner.UID]
	zones.Graveyard = append(zones.Graveyard, object.Card)
	g.state.Zones[object.Owner.UID] = zones
	return g.commitActionDeclaration()
}

func (g *Game) commitActionDeclaration() error {
	declaration := g.state.Knowledge.Declaration
	if declaration == nil || !g.canCommitActionDeclaration(declaration) {
		return fmt.Errorf("%w", tcgErrors.ErrInvalidViewHandle)
	}
	zones := g.state.Zones[declaration.Controller.UID]
	sourceIndex := cardIndex(zones.Hand, declaration.Source)
	if sourceIndex < 0 {
		return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, declaration.Source)
	}
	zones.Hand = removeCardAt(zones.Hand, sourceIndex)
	for paymentCount := 0; paymentCount < g.actionReserveCost(declaration.Controller, declaration.Source); paymentCount++ {
		memoryIndex := int(g.nextRandom() % uint64(len(zones.Memory)))
		payment := zones.Memory[memoryIndex]
		zones.Memory = removeCardAt(zones.Memory, memoryIndex)
		zones.Banishment = append(zones.Banishment, payment)
	}
	g.state.Zones[declaration.Controller.UID] = zones
	g.pushAbility(g.actionAbilityInstance(declaration))
	g.state.Knowledge.Choice = nil
	g.state.Knowledge.Declaration = nil
	g.grantOpportunity(declaration.Controller)
	return nil
}

func (g *Game) canCommitActionDeclaration(declaration *actionDeclaration) bool {
	if !g.isLegalTarget(declaration.Target) {
		return false
	}
	candidate, exists := g.state.Cards[declaration.Source]
	if !exists || !samePlayer(candidate.Owner, declaration.Controller) || !containsString(candidate.Types, "ACTION") {
		return false
	}
	zones := g.state.Zones[declaration.Controller.UID]
	return cardIndex(zones.Hand, declaration.Source) >= 0 && len(zones.Memory) >= g.actionReserveCost(declaration.Controller, declaration.Source)
}

func (g *Game) actionReserveCost(player *model.Player, card cardInstanceID) int {
	cost := g.characteristicsForCard(card).ReserveCost
	if g.state.Cards[card].Definition == trumpSetCardID && g.championHasClass(player, g.state.Cards[card].Classes) && cost > 0 {
		return cost - 1
	}
	return cost
}

func (g *Game) legalTargets() []objectID {
	targets := make([]objectID, 0, len(g.state.Champions)+len(g.state.Objects))
	for _, player := range g.players {
		champion, exists := g.state.Champions[player.UID]
		if exists {
			targets = append(targets, champion.ID)
		}
	}
	for id, object := range g.state.Objects {
		if containsString(object.Types, "ALLY") {
			targets = append(targets, id)
		}
	}
	sort.Slice(
		targets,
		func(first, second int) bool {
			return targets[first] < targets[second]
		},
	)
	return targets
}

func (g *Game) isLegalTarget(target objectID) bool {
	for _, champion := range g.state.Champions {
		if champion.ID == target {
			return true
		}
	}
	object, exists := g.state.Objects[target]
	return exists && containsString(object.Types, "ALLY")
}

func (g *Game) legalWeapons(player *model.Player) []objectID {
	weapons := make([]objectID, 0, len(g.state.Objects))
	for id := range g.state.Objects {
		if g.isLegalWeapon(player, id) {
			weapons = append(weapons, id)
		}
	}
	sort.Slice(
		weapons,
		func(first, second int) bool {
			return weapons[first] < weapons[second]
		},
	)
	return weapons
}

func (g *Game) isLegalWeapon(player *model.Player, id objectID) bool {
	object, exists := g.state.Objects[id]
	return exists && samePlayer(object.Owner, player) && containsString(object.Types, "WEAPON")
}

func (g *Game) actionAbilityInstance(declaration *actionDeclaration) abilityInstance {
	operations := []effectOperation{}
	source := g.state.Cards[declaration.Source]
	switch source.Definition {
	case blazingThrowCardID:
		operations = append(operations, effectOperation{
			Kind:   effectOperationDamage,
			Amount: 4,
		})
	case fieryInterferenceCardID:
		operations = append(operations, effectOperation{
			Kind:   effectOperationDamage,
			Amount: 2,
		})
		operations = append(operations, effectOperation{
			Kind: effectOperationContinuousModifier,
			ContinuousEffect: continuousEffect{
				Scope:         effectScopeObject,
				Layer:         effectLayerAbility,
				ExpiresAtTurn: g.state.Scheduler.TurnNumber + 1,
				Modifier: continuousModifier{
					ProhibitRecover: true,
				},
			},
		})
	case straightFlareCardID:
		operations = append(operations, effectOperation{
			Kind:                      effectOperationDamage,
			Amount:                    1,
			DistinctSuitedCostsDamage: true,
		})
	case trumpSetCardID:
		operations = append(operations, effectOperation{Kind: effectOperationRetargetAttack})
	}
	operations = append(operations, effectOperation{
		Kind:                  effectOperationMove,
		MoveSourceToGraveyard: true,
	})
	return g.newAbilityInstance(
		declaration.Controller,
		declaration.Source,
		declaration.Target,
		operations,
	)
}

func (g *Game) hasActiveCombat() bool {
	for _, item := range g.state.EffectsStack {
		if item.Kind == effectStackCombat {
			return true
		}
	}
	return false
}

func (g *Game) removeEffectSource(source cardInstanceID) {
	index := cardIndex(g.state.EffectSources, source)
	if index >= 0 {
		g.state.EffectSources = removeCardAt(g.state.EffectSources, index)
	}
}

func (g *Game) putInGraveyard(card cardInstanceID) {
	owner := g.state.Cards[card].Owner
	zones := g.state.Zones[owner.UID]
	zones.Graveyard = append(zones.Graveyard, card)
	g.state.Zones[owner.UID] = zones
}

func (g *Game) damageUnit(target objectID, amount int) {
	for playerID, champion := range g.state.Champions {
		if champion.ID == target {
			champion.Damage += amount
			champion.DamageTurn = g.state.Scheduler.TurnNumber
			g.state.Champions[playerID] = champion
			return
		}
	}
	object, exists := g.state.Objects[target]
	if exists {
		object.Damage += amount
		g.state.Objects[target] = object
	}
}

func (g *Game) isChampion(target objectID) bool {
	for _, champion := range g.state.Champions {
		if champion.ID == target {
			return true
		}
	}
	return false
}

func (g *Game) recoverChampion(target objectID, amount int) bool {
	for playerID, champion := range g.state.Champions {
		if champion.ID != target {
			continue
		}
		if g.characteristicsFor(target).RecoverProhibited {
			return false
		}
		champion.Damage -= amount
		if champion.Damage < 0 {
			champion.Damage = 0
		}
		g.state.Champions[playerID] = champion
		return true
	}
	return false
}

func (g *Game) distinctSuitedPrintedReserveCosts(player *model.Player) int {
	costs := make(map[int]struct{})
	for _, object := range g.state.Objects {
		if !samePlayer(object.Owner, player) || !g.cardHasSubtype(object.Card, "SUITED") {
			continue
		}
		costs[g.printedReserveCost(object.Card)] = struct{}{}
	}
	return len(costs)
}

func (g *Game) cardHasSubtype(card cardInstanceID, want string) bool {
	return containsString(g.state.Cards[card].Subtypes, want)
}

func (g *Game) printedReserveCost(card cardInstanceID) int {
	return g.state.Cards[card].ReserveCost
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
