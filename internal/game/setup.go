package game

import (
	"fmt"
	"go-tcg/internal/constants"
)

const spiritOfFireOnEnterCause = "ability:LMyKyVC2O9:front:on-enter-draw-seven"

type cardInstanceID string

type objectID string

type cardInstance struct {
	ID         cardInstanceID     `json:"id"`
	Owner      constants.PlayerID `json:"owner"`
	Definition CardID             `json:"definition"`
	Face       CardFaceID         `json:"face"`
}

type playerZones struct {
	MainDeck        []cardInstanceID `json:"main_deck"`
	Hand            []cardInstanceID `json:"hand"`
	MaterialDeck    []cardInstanceID `json:"material_deck"`
	OutsideGamePool []cardInstanceID `json:"outside_game_pool"`
}

type championObject struct {
	ID    objectID           `json:"id"`
	Card  cardInstanceID     `json:"card"`
	Owner constants.PlayerID `json:"owner"`
}

type schedulerKind string

const (
	schedulerStable   schedulerKind = "stable"
	schedulerFinished schedulerKind = "finished"
)

type schedulerFrame struct {
	Kind              schedulerKind      `json:"kind"`
	TurnPlayer        constants.PlayerID `json:"turn_player"`
	Phase             Phase              `json:"phase,omitempty"`
	OpportunityHolder constants.PlayerID `json:"opportunity_holder,omitempty"`
	ConsecutivePasses int                `json:"consecutive_passes"`
	TurnNumber        uint64             `json:"turn_number"`
}

type gameEvent struct {
	Sequence uint64         `json:"sequence"`
	Kind     string         `json:"kind"`
	Card     cardInstanceID `json:"card"`
}

type eventBatch struct {
	Player constants.PlayerID `json:"player"`
	Cause  string             `json:"cause"`
	Events []gameEvent        `json:"events"`
}

// NewStandardSetup builds the deterministic opening state for the fixed deck.
// It deliberately omits the complete Support Set gate, which remains enforced
// by NewStandardGame until every reachable card behavior is supported.
func NewStandardSetup(configuration StandardGameConfig) (*Game, error) {
	decks, err := loadValidatedStandardDecks(configuration)
	if err != nil {
		return nil, err
	}
	return newStandardSetup(
		configuration,
		decks.Definitions,
		decks.First,
		decks.Second,
	)
}

func newStandardSetup(
	configuration StandardGameConfig,
	definitions map[CardID]CardDefinition,
	firstDeck DeckManifest,
	secondDeck DeckManifest,
) (*Game, error) {
	game := NewGame(configuration.Seed)
	game.players = []constants.PlayerID{
		configuration.Players[0],
		configuration.Players[1],
	}
	game.state.NextHandle = 0
	game.initializeKnowledgeState()
	game.state.Zones = make(map[constants.PlayerID]playerZones, len(game.players))
	game.state.Champions = make(map[constants.PlayerID]championObject, len(game.players))

	decks := []DeckManifest{
		firstDeck,
		secondDeck,
	}
	for index, player := range game.players {
		if err := game.addPlayerDeck(
			player,
			decks[index],
			definitions,
		); err != nil {
			return nil, err
		}
		game.shuffleMainDeck(player)
	}
	for _, player := range game.players {
		if !game.resolveSpiritOfFireOnEnter(player) {
			break
		}
	}
	if !game.state.Finished {
		game.state.Scheduler = schedulerFrame{
			Kind:       schedulerStable,
			TurnPlayer: game.players[0],
			Phase:      PhaseWakeUp,
			TurnNumber: 1,
		}
		game.runStandardScheduler()
	}
	game.advanceKnowledgeRevision()
	return game, nil
}

func (g *Game) addPlayerDeck(player constants.PlayerID, deck DeckManifest, definitions map[CardID]CardDefinition) error {
	zones := playerZones{}
	for _, entry := range deck.MainDeck {
		for count := 0; count < entry.Count; count++ {
			card := g.newCardInstance(player, entry, definitions)
			zones.MainDeck = append(zones.MainDeck, card)
		}
	}
	for _, entry := range deck.MaterialDeck {
		for count := 0; count < entry.Count; count++ {
			card := g.newCardInstance(player, entry, definitions)
			definition := definitions[entry.CardID]
			if definition.Face().HasType("CHAMPION") && definition.Face().Level() == 0 {
				if _, exists := g.state.Champions[player]; exists {
					return fmt.Errorf("player %q has multiple starting Champions", player)
				}
				playerName := string(player)
				championID := objectID("champion:" + playerName)
				champion := championObject{
					ID:    championID,
					Card:  card,
					Owner: player,
				}
				g.state.Champions[player] = champion
				g.grantCardTracking(player, entityID(card))
				for _, opponent := range g.players {
					if opponent != player {
						g.grantCardTracking(opponent, entityID(card))
					}
				}
				continue
			}
			zones.MaterialDeck = append(zones.MaterialDeck, card)
		}
	}
	for _, entry := range deck.OutsideGamePool {
		for count := 0; count < entry.Count; count++ {
			card := g.newCardInstance(player, entry, definitions)
			zones.OutsideGamePool = append(zones.OutsideGamePool, card)
		}
	}
	g.state.Zones[player] = zones
	return nil
}

func (g *Game) newCardInstance(player constants.PlayerID, entry DeckEntry, definitions map[CardID]CardDefinition) cardInstanceID {
	cardCount := len(g.state.Cards) + 1
	identifierText := fmt.Sprintf(
		"card:%s:%d",
		player,
		cardCount,
	)
	identifier := cardInstanceID(identifierText)
	g.state.Cards[identifier] = cardInstance{
		ID:         identifier,
		Owner:      player,
		Definition: entry.CardID,
		Face:       entry.FaceID,
	}
	g.state.Entities[entityID(identifier)] = knowledgeEntity{
		Name: definitions[entry.CardID].Name(),
	}
	return identifier
}

func (g *Game) shuffleMainDeck(player constants.PlayerID) {
	zones := g.state.Zones[player]
	for index := len(zones.MainDeck) - 1; index > 0; index-- {
		swapIndex := int(g.nextRandom() % uint64(index+1))
		zones.MainDeck[index], zones.MainDeck[swapIndex] = zones.MainDeck[swapIndex], zones.MainDeck[index]
	}
	g.state.Zones[player] = zones
}

func (g *Game) nextRandom() uint64 {
	g.state.PRNG.Cursor++
	value := g.state.PRNG.Seed + 0x9e3779b97f4a7c15*g.state.PRNG.Cursor
	value = (value ^ (value >> 30)) * 0xbf58476d1ce4e5b9
	value = (value ^ (value >> 27)) * 0x94d049bb133111eb
	return value ^ (value >> 31)
}

func (g *Game) resolveSpiritOfFireOnEnter(player constants.PlayerID) bool {
	batch := eventBatch{
		Player: player,
		Cause:  spiritOfFireOnEnterCause,
		Events: make([]gameEvent, 0, 7),
	}
	for draw := 0; draw < 7; draw++ {
		zones := g.state.Zones[player]
		if len(zones.MainDeck) == 0 {
			if len(batch.Events) > 0 {
				g.state.Events = append(g.state.Events, batch)
			}
			g.state.Finished = true
			g.state.Winner = g.otherPlayer(player)
			g.state.Scheduler = schedulerFrame{
				Kind: schedulerFinished,
			}
			return false
		}
		card := zones.MainDeck[len(zones.MainDeck)-1]
		zones.MainDeck = zones.MainDeck[:len(zones.MainDeck)-1]
		zones.Hand = append(zones.Hand, card)
		g.state.Zones[player] = zones
		g.grantCardTracking(player, entityID(card))
		g.state.NextEvent++
		batch.Events = append(batch.Events, gameEvent{
			Sequence: g.state.NextEvent,
			Kind:     "draw",
			Card:     card,
		})
		g.recordVisibleEvent(player, "draw", entityID(card))
	}
	g.state.Events = append(g.state.Events, batch)
	return true
}
