package game

import (
	"fmt"
	"sort"

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

// passOpportunity 僅接受穩定排程下的行動機會持有者 pass。
// 所有玩家連續 pass 後只結算堆疊頂的一項；一般行動結算後重新授予回合玩家行動機會。
// Materialize 結算完畢或在空堆疊上連續 pass 時推進階段，結算導致對局結束時直接返回。
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
		if len(g.state.EffectsStack) == 0 && scheduler.Phase == PhaseMaterialize {
			g.advanceAfterOpportunity()
			return nil
		}
		g.grantOpportunity(scheduler.TurnPlayer)
		return nil
	}
	g.advanceAfterOpportunity()
	return nil
}

// advanceAfterOpportunity 在行動機會結束後推進階段，再執行標準排程。
// 離開回憶階段時才回收 Memory；回合結束時換玩家、增加回合數並清除 Cardistry 折扣。
// 僅允許從具行動機會的階段進入，其他階段呼叫會 panic。
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

// runStandardScheduler 自動執行喚醒與抽牌等階段，直到需要玩家輸入或對局結束。
// 首輪各玩家略過物質化與回憶階段，只有第 1 回合略過回合抽牌。
// 物質化階段等待玩家選牌；回憶、主階段與結束階段授予行動機會後暫停。
// 未知階段屬內部排程錯誤，會 panic。
func (g *Game) runStandardScheduler() {
	for !g.state.Finished {
		scheduler := &g.state.Scheduler
		switch scheduler.Phase {
		case PhaseWakeUp:
			g.wakeUpObjects(scheduler.TurnPlayer)
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

// wakeUpObjects 同時喚醒回合玩家控制的所有 rested Champion 與 Field objects，並以穩定順序記錄事件。
// 輸入為回合玩家；輸出為零值，副作用為清除合格 object 的 Rested 並追加一個公開 simultaneous event batch。
func (g *Game) wakeUpObjects(player *model.Player) {
	wokenCapacity := len(g.state.Objects) + 1
	woken := make([]cardInstanceID, 0, wokenCapacity)
	champion, championExists := g.state.Champions[player.UID]
	if championExists && champion.Rested {
		champion.Rested = false
		g.state.Champions[player.UID] = champion
		woken = append(woken, champion.Card)
	}
	objectCapacity := len(g.state.Objects)
	objects := make([]objectID, 0, objectCapacity)
	for id, object := range g.state.Objects {
		if samePlayer(object.Owner, player) && object.Rested {
			objects = append(objects, id)
		}
	}
	sort.Slice(
		objects,
		func(first, second int) bool {
			return objects[first] < objects[second]
		},
	)
	for _, id := range objects {
		object := g.state.Objects[id]
		object.Rested = false
		g.state.Objects[id] = object
		woken = append(woken, object.Card)
	}
	if len(woken) == 0 {
		return
	}
	wokenCount := len(woken)
	batch := eventBatch{
		Player:       player,
		Cause:        "turn:wake-up",
		Simultaneous: wokenCount > 1,
		Events:       make([]gameEvent, 0, wokenCount),
	}
	for _, card := range woken {
		g.state.NextEvent++
		batch.Events = append(
			batch.Events,
			gameEvent{
				Sequence: g.state.NextEvent,
				Kind:     "wake-up",
				Card:     card,
			},
		)
		for _, viewer := range g.players {
			cardEntity := entityID(card)
			g.recordVisibleEvent(viewer, "wake-up", cardEntity)
		}
	}
	g.state.Events = append(g.state.Events, batch)
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
	return g.drawCardsWithDeckOut(
		player,
		1,
		"turn:draw",
	)
}

// drawCardsWithDeckOut 將指定張數由主牌組頂移入手牌，並在牌組抽空時結束對局。
// 此行為僅適用於標準回合與開局抽牌；能力抽牌使用 drawCards，且不會造成敗北。
func (g *Game) drawCardsWithDeckOut(player *model.Player, count int, cause string) bool {
	batch := eventBatch{
		Player: player,
		Cause:  cause,
		Events: make([]gameEvent, 0, count),
	}
	for range count {
		zones := g.state.Zones[player.UID]
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
		g.state.Zones[player.UID] = zones
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
