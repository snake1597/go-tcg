package game

import (
	"fmt"
	"go-tcg/internal/constants"
	tcgErrors "go-tcg/internal/tcg_errors"
)

type Phase string

const (
	PhaseWakeUp       Phase = "wake_up"
	PhaseMaterialize  Phase = "materialize"
	PhaseRecollection Phase = "recollection"
	PhaseDraw         Phase = "draw"
	PhaseMain         Phase = "main"
	PhaseEnd          Phase = "end"
)

func (g *Game) passOpportunity(player constants.PlayerID) error {
	scheduler := &g.state.Scheduler
	if scheduler.Kind != schedulerStable || scheduler.OpportunityHolder != player {
		return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, player)
	}
	scheduler.ConsecutivePasses++
	if scheduler.ConsecutivePasses < len(g.players) {
		scheduler.OpportunityHolder = g.nextPlayer(player)
		return nil
	}
	scheduler.OpportunityHolder = ""
	scheduler.ConsecutivePasses = 0
	if len(g.state.EffectsStack) > 0 {
		g.resolveTopEffectStack()
		if len(g.state.EffectsStack) > 0 {
			g.grantOpportunity(scheduler.TurnPlayer)
			return nil
		}
	}
	g.advanceAfterOpportunity()
	return nil
}

func (g *Game) advanceAfterOpportunity() {
	scheduler := &g.state.Scheduler
	switch scheduler.Phase {
	case PhaseMaterialize:
		scheduler.Phase = PhaseRecollection
	case PhaseRecollection:
		scheduler.Phase = PhaseDraw
	case PhaseMain:
		scheduler.Phase = PhaseEnd
	case PhaseEnd:
		scheduler.TurnPlayer = g.nextPlayer(scheduler.TurnPlayer)
		scheduler.TurnNumber++
		scheduler.Phase = PhaseWakeUp
	default:
		panic(fmt.Sprintf("cannot advance after opportunity in phase %q", scheduler.Phase))
	}
	g.runStandardScheduler()
}

func (g *Game) skipMaterialize(player constants.PlayerID) error {
	scheduler := &g.state.Scheduler
	if scheduler.Kind != schedulerStable || scheduler.Phase != PhaseMaterialize || scheduler.TurnPlayer != player {
		return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, player)
	}
	scheduler.Phase = PhaseRecollection
	g.runStandardScheduler()
	return nil
}

func (g *Game) runStandardScheduler() {
	for !g.state.Finished {
		scheduler := &g.state.Scheduler
		switch scheduler.Phase {
		case PhaseWakeUp:
			g.expireTimedChampionEffects()
			scheduler.Phase = PhaseMaterialize
		case PhaseMaterialize:
			if g.isFirstTurn() {
				scheduler.Phase = PhaseRecollection
				continue
			}
			return
		case PhaseRecollection:
			if g.isFirstTurn() {
				scheduler.Phase = PhaseDraw
				continue
			}
			g.grantOpportunity(scheduler.TurnPlayer)
			return
		case PhaseDraw:
			if scheduler.TurnNumber != 1 && !g.drawTurnCard(scheduler.TurnPlayer) {
				return
			}
			scheduler.Phase = PhaseMain
		case PhaseMain, PhaseEnd:
			g.grantOpportunity(scheduler.TurnPlayer)
			return
		default:
			panic(fmt.Sprintf("unknown standard phase %q", scheduler.Phase))
		}
	}
}

func (g *Game) grantOpportunity(player constants.PlayerID) {
	g.state.Scheduler.OpportunityHolder = player
	g.state.Scheduler.ConsecutivePasses = 0
}

func (g *Game) isFirstTurn() bool {
	return g.state.Scheduler.TurnNumber <= uint64(len(g.players))
}

func (g *Game) nextPlayer(player constants.PlayerID) constants.PlayerID {
	for index, candidate := range g.players {
		if candidate != player {
			continue
		}
		return g.players[(index+1)%len(g.players)]
	}
	panic(fmt.Sprintf("unknown player %q", player))
}

func (g *Game) drawTurnCard(player constants.PlayerID) bool {
	zones := g.state.Zones[player]
	if len(zones.MainDeck) == 0 {
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
	g.state.Events = append(
		g.state.Events,
		eventBatch{
			Player: player,
			Cause:  "turn:draw",
			Events: []gameEvent{
				{
					Sequence: g.state.NextEvent,
					Kind:     "draw",
					Card:     card,
				},
			},
		},
	)
	g.recordVisibleEvent(player, "draw", entityID(card))
	return true
}
