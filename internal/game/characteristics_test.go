package game

import (
	"testing"

	"go-tcg/internal/model"
)

func TestCharacteristicsApplyArthurRestedAllyModifierWithoutMutation(t *testing.T) {
	game := newActionGame(t)
	arthurCard := findCard(t, game, model.PlayerOne, CardID("rufki4o41y"))
	allyCard := findCard(t, game, model.PlayerOne, CardID("iohZMWh5v5"))
	arthur := game.state.Cards[arthurCard]
	arthur.Definition = arthurYoungHeirCardID
	arthur.Types = []string{"ALLY"}
	game.state.Cards[arthurCard] = arthur
	ally := game.state.Cards[allyCard]
	ally.Types = []string{"ALLY"}
	ally.Power = 2
	game.state.Cards[allyCard] = ally
	game.state.Objects[objectID("arthur")] = fieldObject{
		ID:     objectID("arthur"),
		Card:   arthurCard,
		Owner:  model.PlayerOne,
		Types:  []string{"ALLY"},
		Rested: true,
	}
	allyID := objectID("ally")
	game.state.Objects[allyID] = fieldObject{
		ID:    allyID,
		Card:  allyCard,
		Owner: model.PlayerOne,
		Types: []string{"ALLY"},
	}
	if got := game.characteristicsFor(allyID).Power; got != 3 {
		t.Fatalf("derived ally power = %d, want 3", got)
	}
	if got := game.state.Cards[allyCard].Power; got != 2 {
		t.Fatalf("printed ally power = %d, want unchanged 2", got)
	}
}

func TestCharacteristicsApplyBulwarkClassBonus(t *testing.T) {
	game := newActionGame(t)
	weaponCard := findCard(t, game, model.PlayerOne, CardID("rufki4o41y"))
	weapon := game.state.Cards[weaponCard]
	weapon.Definition = bulwarkSwordCardID
	weapon.Classes = []string{"GUARDIAN"}
	weapon.Power = 2
	game.state.Cards[weaponCard] = weapon
	champion := game.state.Champions[model.PlayerOne.UID]
	championCard := game.state.Cards[champion.Card]
	championCard.Classes = []string{"GUARDIAN"}
	game.state.Cards[champion.Card] = championCard
	weaponID := objectID("bulwark")
	game.state.Objects[weaponID] = fieldObject{
		ID:    weaponID,
		Card:  weaponCard,
		Owner: model.PlayerOne,
		Types: []string{"WEAPON"},
	}
	if got := game.characteristicsFor(weaponID).Power; got != 3 {
		t.Fatalf("derived weapon power = %d, want 3", got)
	}
}

func TestImmortalityPreventsStateBasedDeathUntilExpiry(t *testing.T) {
	game := newActionGame(t)
	card := findCard(t, game, model.PlayerOne, CardID("rufki4o41y"))
	game.state.Cards[card] = withCombatStats(game.state.Cards[card], 1, 1)
	ally := objectID("immortal-ally")
	game.state.Objects[ally] = fieldObject{
		ID:     ally,
		Card:   card,
		Owner:  model.PlayerOne,
		Types:  []string{"ALLY"},
		Damage: 1,
	}
	game.addContinuousEffect(continuousEffect{
		Target:        ally,
		Scope:         effectScopeObject,
		Layer:         effectLayerAbility,
		ExpiresAtTurn: game.state.Scheduler.TurnNumber + 1,
		Modifier: continuousModifier{
			GrantImmortality: true,
		},
	})
	game.resolveCombatStateBased()
	if _, exists := game.state.Objects[ally]; !exists {
		t.Fatal("immortal ally was destroyed")
	}
	game.state.Scheduler.TurnNumber++
	game.resolveCombatStateBased()
	if _, exists := game.state.Objects[ally]; exists {
		t.Fatal("expired immortality kept lethal ally on the Field")
	}
}

func TestArthurImmortalityRestsArthurForTheCurrentTurn(t *testing.T) {
	game := newActionGame(t)
	card := findCard(t, game, model.PlayerOne, CardID("rufki4o41y"))
	arthur := game.state.Cards[card]
	arthur.Definition = arthurYoungHeirCardID
	game.state.Cards[card] = arthur
	id := objectID("arthur")
	game.state.Objects[id] = fieldObject{ID: id, Card: card, Owner: model.PlayerOne}
	game.grantArthurImmortality(id)
	object := game.state.Objects[id]
	if !object.Rested || !game.isImmortal(id) {
		t.Fatalf("Arthur state = %#v, want rested immortality through the owner's next turn", object)
	}
}

func TestCharacteristicsApplyPowerLifeSubLayersThenTimestamp(t *testing.T) {
	game := newActionGame(t)
	card := findCard(t, game, model.PlayerOne, CardID("rufki4o41y"))
	fixture := game.state.Cards[card]
	fixture.Power = 2
	game.state.Cards[card] = fixture
	id := objectID("layered-ally")
	game.state.Objects[id] = fieldObject{
		ID:    id,
		Card:  card,
		Owner: model.PlayerOne,
		Types: []string{"ALLY"},
	}
	setPower := 5
	game.addContinuousEffect(continuousEffect{
		Target:    id,
		Scope:     effectScopeObject,
		Layer:     effectLayerModifier,
		PowerLife: powerLifeModify,
		Modifier: continuousModifier{
			PowerDelta: 2,
		},
	})
	game.addContinuousEffect(continuousEffect{
		Target:    id,
		Scope:     effectScopeObject,
		Layer:     effectLayerModifier,
		PowerLife: powerLifeSet,
		Modifier: continuousModifier{
			SetPower: &setPower,
		},
	})
	if got := game.characteristicsFor(id).Power; got != 7 {
		t.Fatalf("derived layered power = %d, want 7", got)
	}
}

func TestCharacteristicsResolveDependencyBeforeTimestampAndBreakLoopsByTimestamp(t *testing.T) {
	game := newActionGame(t)
	card := findCard(t, game, model.PlayerOne, CardID("rufki4o41y"))
	id := objectID("dependency-ally")
	game.state.Objects[id] = fieldObject{ID: id, Card: card, Owner: model.PlayerOne, Types: []string{"ALLY"}}
	first := continuousEffect{ID: 2, Target: id, Scope: effectScopeObject, Layer: effectLayerModifier, PowerLife: powerLifeModify, Timestamp: 2, DependsOn: []uint64{1}, Modifier: continuousModifier{PowerDelta: 2}}
	second := continuousEffect{ID: 1, Target: id, Scope: effectScopeObject, Layer: effectLayerModifier, PowerLife: powerLifeModify, Timestamp: 1, DependsOn: []uint64{2}, Modifier: continuousModifier{PowerDelta: 1}}
	game.state.ContinuousEffects = []continuousEffect{first, second}
	effects := game.orderedEffectsFor(id)
	if len(effects) != 2 || effects[0].ID != 1 || effects[1].ID != 2 {
		t.Fatalf("dependency loop order = %#v, want timestamp order", effects)
	}
}

func TestArthurImmortalityExpiresAtBeginningOfOwnersNextTurn(t *testing.T) {
	game := newActionGame(t)
	card := findCard(t, game, model.PlayerOne, CardID("rufki4o41y"))
	arthur := game.state.Cards[card]
	arthur.Definition = arthurYoungHeirCardID
	game.state.Cards[card] = arthur
	id := objectID("arthur")
	game.state.Objects[id] = fieldObject{
		ID:    id,
		Card:  card,
		Owner: model.PlayerOne,
	}
	game.grantArthurImmortality(id)
	game.state.Scheduler.TurnNumber++
	game.expireTimedChampionEffects()
	if !game.isImmortal(id) {
		t.Fatal("Arthur lost immortality during the opponent's turn")
	}
	game.state.Scheduler.TurnNumber++
	game.expireTimedChampionEffects()
	if game.isImmortal(id) {
		t.Fatal("Arthur retained immortality at the beginning of the owner's next turn")
	}
}

func TestZeroImmortalityDurationIsNotActive(t *testing.T) {
	game := newTestGame(1)
	id := objectID("ordinary")
	game.state.Objects[id] = fieldObject{}
	if game.isImmortal(id) {
		t.Fatal("zero immortality duration was active")
	}
}

func TestPlayerViewUsesDerivedChampionCharacteristics(t *testing.T) {
	game := newActionGame(t)
	champion := game.state.Champions[model.PlayerOne.UID]
	card := game.state.Cards[champion.Card]
	card.Power = 7
	card.Life = 11
	game.state.Cards[champion.Card] = card
	view, err := game.PlayerView(model.PlayerOne)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	for _, visible := range view.Champions {
		if !samePlayer(visible.Owner, model.PlayerOne) {
			continue
		}
		if visible.Power != 7 || visible.Life != 11 {
			t.Fatalf("visible characteristics = %#v, want power 7 life 11", visible)
		}
		return
	}
	t.Fatal("player one's Champion was not visible")
}

func TestBulwarkWieldRequiresAndPaysTwoReserve(t *testing.T) {
	game := newActionGame(t)
	weaponCard := findCard(t, game, model.PlayerOne, CardID("rufki4o41y"))
	weapon := game.state.Cards[weaponCard]
	weapon.Definition = bulwarkSwordCardID
	game.state.Cards[weaponCard] = weapon
	weaponID := objectID("bulwark")
	game.state.Objects[weaponID] = fieldObject{
		ID:    weaponID,
		Card:  weaponCard,
		Owner: model.PlayerOne,
		Types: []string{"WEAPON"},
	}
	zones := game.state.Zones[model.PlayerOne.UID]
	zones.Memory = zones.Hand[:1]
	game.state.Zones[model.PlayerOne.UID] = zones
	if game.canWield(model.PlayerOne, weaponID) {
		t.Fatal("Bulwark was legal with fewer than two reserve")
	}
	zones.Memory = append(zones.Memory, zones.Hand[1])
	game.state.Zones[model.PlayerOne.UID] = zones
	if !game.canWield(model.PlayerOne, weaponID) {
		t.Fatal("Bulwark was not legal with two reserve")
	}
	game.payWieldReserve(model.PlayerOne, weaponID)
	zones = game.state.Zones[model.PlayerOne.UID]
	if len(zones.Memory) != 0 || len(zones.Banishment) < 2 {
		t.Fatalf("wield payment zones = %#v, want two banished reserve", zones)
	}
}
