package game

import (
	"fmt"
	"go-tcg/internal/model"
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

func (g *Game) passOpportunity(player *model.Player) error {
	scheduler := &g.state.Scheduler
	if scheduler.Kind != schedulerStable || !samePlayer(scheduler.OpportunityHolder, player) {
		return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, player)
	}
	scheduler.ConsecutivePasses++
	if scheduler.ConsecutivePasses < len(g.players) {
		scheduler.OpportunityHolder = g.nextPlayer(player)
		return nil
	}
	scheduler.OpportunityHolder = nil
	scheduler.ConsecutivePasses = 0
	if len(g.state.EffectsStack) > 0 {
		g.resolveTopEffectStack()
		if g.state.Finished {
			return nil
		}
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
		g.recollectMemory(scheduler.TurnPlayer)
		scheduler.Phase = PhaseDraw
	case PhaseMain:
		scheduler.Phase = PhaseEnd
	case PhaseEnd:
		scheduler.TurnPlayer = g.nextPlayer(scheduler.TurnPlayer)
		scheduler.TurnNumber++
		clear(g.state.CardistryDiscounts)
		scheduler.Phase = PhaseWakeUp
	default:
		panic(fmt.Sprintf("cannot advance after opportunity in phase %q", scheduler.Phase))
	}
	g.runStandardScheduler()
}

func (g *Game) skipMaterialize(player *model.Player) error {
	scheduler := &g.state.Scheduler
	if scheduler.Kind != schedulerStable || scheduler.Phase != PhaseMaterialize || !samePlayer(scheduler.TurnPlayer, player) {
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
			g.wakeUpChampion(scheduler.TurnPlayer)
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

func (g *Game) wakeUpChampion(player *model.Player) {
	champion, exists := g.state.Champions[player.UID]
	if !exists || !champion.Rested {
		return
	}
	champion.Rested = false
	g.state.Champions[player.UID] = champion
	g.recordPublicEvent(
		player,
		"turn:wake-up",
		"wake-up",
		champion.Card,
	)
}

func (g *Game) recollectMemory(player *model.Player) {
	zones := g.state.Zones[player.UID]
	if len(zones.Memory) == 0 {
		return
	}
	batch := eventBatch{
		Player: player,
		Cause:  "turn:recollection",
		Events: make([]gameEvent, 0, len(zones.Memory)),
	}
	for _, card := range zones.Memory {
		zones.Hand = append(zones.Hand, card)
		g.grantCardTracking(player, entityID(card))
		g.state.NextEvent++
		batch.Events = append(
			batch.Events,
			gameEvent{
				Sequence: g.state.NextEvent,
				Kind:     "recollect",
				Card:     card,
			},
		)
		g.recordVisibleEvent(player, "recollect", entityID(card))
	}
	zones.Memory = nil
	g.state.Zones[player.UID] = zones
	g.state.Events = append(g.state.Events, batch)
}

func (g *Game) grantOpportunity(player *model.Player) {
	g.state.Scheduler.OpportunityHolder = player
	g.state.Scheduler.ConsecutivePasses = 0
}

func (g *Game) isFirstTurn() bool {
	return g.state.Scheduler.TurnNumber <= uint64(len(g.players))
}

func (g *Game) nextPlayer(player *model.Player) *model.Player {
	for index, candidate := range g.players {
		if !samePlayer(candidate, player) {
			continue
		}
		return g.players[(index+1)%len(g.players)]
	}
	panic(fmt.Sprintf("unknown player %q", player.UID))
}

func (g *Game) drawTurnCard(player *model.Player) bool {
	zones := g.state.Zones[player.UID]
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
	g.state.Zones[player.UID] = zones
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
