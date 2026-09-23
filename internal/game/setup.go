package game

import (
	"fmt"
	"go-tcg/internal/constants"
	"go-tcg/internal/model"
)

const spiritOfFireOnEnterCause = "ability:LMyKyVC2O9:front:on-enter-draw-seven"

// newStandardSetup 使用已驗證的雙方牌組建立卡牌實例，洗牌並處理 Spirit of Fire 開局抽牌。
// 開局未導致對局結束時才啟動標準排程；最後更新 revision 並保存 replay 初始快照。
// configuration、definitions 與牌組的一致性由呼叫端先行驗證。
func newStandardSetup(
	configuration StandardGameConfig,
	definitions map[CardID]CardDefinition,
	firstDeck DeckManifest,
	secondDeck DeckManifest,
) (*Game, error) {
	game := NewGame(configuration)

	decks := []DeckManifest{
		firstDeck,
		secondDeck,
	}
	for index, player := range game.players {
		err := game.addPlayerDeck(player, decks[index], definitions)
		if err != nil {
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
			Phase:      constants.PhaseWakeUp,
			TurnNumber: 1,
		}
		game.runStandardScheduler()
	}
	game.advanceKnowledgeRevision()
	game.captureReplayInitialState()
	return game, nil
}

func (g *Game) addPlayerDeck(player *model.Player, deck DeckManifest, definitions map[CardID]CardDefinition) error {
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
				if _, exists := g.state.Champions[player.UID]; exists {
					return fmt.Errorf("player %q has multiple starting Champions", player)
				}
				playerName := player.UID
				championID := objectID("champion:" + playerName)
				champion := championObject{
					ID:    championID,
					Card:  card,
					Owner: player,
				}
				g.state.Champions[player.UID] = champion
				g.grantCardTracking(player, entityID(card))
				for _, opponent := range g.players {
					if !samePlayer(opponent, player) {
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
	g.state.Zones[player.UID] = zones
	return nil
}

func (g *Game) shuffleMainDeck(player *model.Player) {
	zones := g.state.Zones[player.UID]
	for index := len(zones.MainDeck) - 1; index > 0; index-- {
		swapIndex := int(g.nextRandom() % uint64(index+1))
		zones.MainDeck[index], zones.MainDeck[swapIndex] = zones.MainDeck[swapIndex], zones.MainDeck[index]
	}
	g.state.Zones[player.UID] = zones
}

// nextRandom 使用種子與遞增 cursor 產生固定亂數序列，每次呼叫都推進 cursor。
// 重播必須保留種子、cursor 與呼叫順序。
func (g *Game) nextRandom() uint64 {
	g.state.PRNG.Cursor++
	value := g.state.PRNG.Seed + 0x9e3779b97f4a7c15*g.state.PRNG.Cursor
	value = (value ^ (value >> 30)) * 0xbf58476d1ce4e5b9
	value = (value ^ (value >> 27)) * 0x94d049bb133111eb
	return value ^ (value >> 31)
}

func (g *Game) resolveSpiritOfFireOnEnter(player *model.Player) bool {
	return g.drawCardsWithDeckOut(
		player,
		7,
		spiritOfFireOnEnterCause,
	)
}
