package game

import (
	"fmt"
	"go-tcg/internal/constants"
	"go-tcg/internal/model"
	tcgErrors "go-tcg/internal/tcg_errors"

	"github.com/samber/lo"
)

type Game struct {
	versions Versions
	players  []*model.Player
	state    gameState
	replay   Replay
}

// NewGame 建立引擎的空白狀態並固定亂數種子與版本。
// 玩家牌組、開局事件與標準回合排程由 setup 層配置。
func NewGame(config StandardGameConfig) *Game {
	versions := currentVersions()
	game := &Game{
		versions: versions,
		players:  config.Players,
		state: gameState{
			Revision:           1,
			Entities:           make(map[entityID]knowledgeEntity),
			Cards:              make(map[cardInstanceID]cardInstance),
			Zones:              make(map[string]playerZones),
			Champions:          make(map[string]championObject),
			Objects:            make(map[objectID]fieldObject),
			CardistryUsed:      make(map[objectID]bool),
			CardistryDiscounts: make(map[string]int),
			Events:             []eventBatch{},
			PRNG: prngState{
				Seed: config.Seed,
			},
		},
		replay: Replay{
			FormatVersion: constants.ReplayFormatVersion,
			Versions:      versions,
			InitialSeed:   config.Seed,
		},
	}
	game.initializeKnowledgeState()
	game.captureReplayInitialState()
	return game
}

func currentVersions() Versions {
	return Versions{
		Engine:   "grand-archive-v1",
		Rules:    "602c917f2f8fd4df7198429a72eb596bf7f647c6",
		CardData: "card-data-v3",
		Deck:     "standard-fire-v2",
		PRNG:     "splitmix64-v1",
	}
}

// Submit 以目前 revision、action kind 的 payload 契約與玩家專屬 handle 驗證輸入，再交給對應行動或選擇流程。
// 等待選擇時只接受該選擇、投降，以及可略過能力的 pass；Reserve 與 MemoryPayment 皆由 action 的宣告規格驗證。
// 成功接受輸入後記錄 replay；驗證或子流程失敗時不記錄此步。
// 此入口沒有統一回滾機制，子流程須自行維持失敗時的狀態契約。
func (g *Game) Submit(player *model.Player, input Input) error {
	if !g.hasPlayer(player) {
		return fmt.Errorf("%w %q", tcgErrors.ErrUnknownPlayer, player)
	}
	if g.state.Finished {
		return fmt.Errorf("%w", tcgErrors.ErrGameFinished)
	}
	if input.Revision != g.state.Revision {
		return fmt.Errorf("%w: got %d, current %d", tcgErrors.ErrStaleRevision, input.Revision, g.state.Revision)
	}
	if err := g.validateInputPayload(player, input); err != nil {
		return fmt.Errorf("validate input payload: %w", err)
	}
	if input.Choice != "" {
		err := g.submitChoice(
			player,
			input,
		)
		if err != nil {
			return fmt.Errorf("submit choice: %w", err)
		}
		g.recordReplayStep(
			player,
			input,
		)
		return nil
	}
	kind, exists := g.state.Knowledge.Actions[player.UID][input.Action]
	if g.state.Knowledge.VeritaCost != nil && g.state.Knowledge.Choice != nil && g.state.Knowledge.Choice.CanPass && exists && kind == constants.ActionPass {
		g.state.Knowledge.VeritaCost = nil
		g.state.Knowledge.Choice = nil
		g.advanceKnowledgeRevision()
		g.recordReplayStep(player, input)
		return nil
	}
	if g.state.AbilityChoice != nil && g.state.AbilityChoice.CanPass && exists && kind == constants.ActionPass {
		if g.state.AbilityChoice.Instance.RuntimeCopy {
			g.destroyRuntimeCopy(g.state.AbilityChoice.Instance.Source)
		}
		g.state.AbilityChoice = nil
		g.state.Knowledge.Choice = nil
		g.advanceKnowledgeRevision()
		g.recordReplayStep(player, input)
		return nil
	}
	if len(input.MemoryPayment) > 0 {
		ability, abilityExists := g.state.Knowledge.Abilities[player.UID][input.Action]
		_, materializationExists := g.state.Knowledge.Materializations[player.UID][input.Action]
		if (!abilityExists || ability.Kind != activatedAbilityCardistry) && !materializationExists {
			return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, input.Action)
		}
	}
	if g.state.Knowledge.Choice != nil && (!exists || kind != constants.ActionConcede) {
		return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, input.Action)
	}
	if !exists {
		card, materializeExists := g.state.Knowledge.Materializations[player.UID][input.Action]
		if materializeExists {
			if err := g.materialize(
				player,
				card,
				input.MemoryPayment,
			); err != nil {
				return fmt.Errorf("materialize champion: %w", err)
			}
		} else {
			card, activateExists := g.state.Knowledge.Activations[player.UID][input.Action]
			if activateExists {
				if containsString(g.state.Cards[card].Types, "ALLY") {
					if err := g.commitAllyActivation(player, card, input.Reserve); err != nil {
						return fmt.Errorf("commit Ally activation: %w", err)
					}
				} else if err := g.beginActionDeclaration(player, card, input.Reserve); err != nil {
					return fmt.Errorf("begin action declaration: %w", err)
				}
			} else {
				attacker, attackExists := g.state.Knowledge.Attacks[player.UID][input.Action]
				if attackExists {
					if err := g.beginAttack(player, attacker); err != nil {
						return fmt.Errorf("begin attack: %w", err)
					}
				} else {
					ability, abilityExists := g.state.Knowledge.Abilities[player.UID][input.Action]
					if abilityExists {
						switch ability.Kind {
						case activatedAbilityCardistry:
							if err := g.activateCardistry(player, ability.Source, input.MemoryPayment); err != nil {
								return fmt.Errorf("activate Cardistry: %w", err)
							}
						case activatedAbilityObject:
							if err := g.beginObjectAbility(player, ability.Source); err != nil {
								return fmt.Errorf("begin object ability: %w", err)
							}
						default:
							return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, input.Action)
						}
					} else {
						weapon, wieldExists := g.state.Knowledge.Wields[player.UID][input.Action]
						if !wieldExists {
							return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, input.Action)
						}
						if err := g.beginWield(player, weapon); err != nil {
							return fmt.Errorf("begin wield: %w", err)
						}
					}
				}
			}
		}
		g.advanceKnowledgeRevision()
		g.recordReplayStep(
			player,
			input,
		)
		return nil
	}
	switch kind {
	case constants.ActionConcede:
		g.state.Finished = true
		g.state.Winner = g.otherPlayer(player)
		g.state.Scheduler = schedulerFrame{
			Kind: schedulerFinished,
		}
	case constants.ActionPass:
		if err := g.passOpportunity(player); err != nil {
			return fmt.Errorf("pass opportunity: %w", err)
		}
	case constants.ActionSkipMaterialize:
		if err := g.skipMaterialize(player); err != nil {
			return fmt.Errorf("skip materialize: %w", err)
		}
	default:
		return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, input.Action)
	}
	g.advanceKnowledgeRevision()
	g.recordReplayStep(
		player,
		input,
	)
	return nil
}

// validateInputPayload 依 action handle 所屬種類限制 Input 可攜帶的欄位，並驗證 activation 的 Reserve 張數。
// 輸入為提交玩家與尚未執行的 Input；輸出為 payload 契約錯誤或 nil，無副作用且不檢查 action 的其餘規則合法性。
func (g *Game) validateInputPayload(player *model.Player, input Input) error {
	if input.Choice != "" {
		if input.Action != "" || len(input.Reserve) > 0 || len(input.MemoryPayment) > 0 {
			return fmt.Errorf("%w: unused input payload for choice", tcgErrors.ErrInvalidViewHandle)
		}
		return nil
	}
	if input.Action == "" {
		return fmt.Errorf("%w: missing action or choice", tcgErrors.ErrInvalidViewHandle)
	}
	if card, exists := g.state.Knowledge.Activations[player.UID][input.Action]; exists {
		if len(input.MemoryPayment) > 0 || len(input.Reserve) != g.activationReserveCost(player, card) {
			return fmt.Errorf("%w: unused input payload for activation", tcgErrors.ErrInvalidViewHandle)
		}
		return nil
	}
	if ability, exists := g.state.Knowledge.Abilities[player.UID][input.Action]; exists {
		if len(input.Reserve) > 0 || (ability.Kind != activatedAbilityCardistry && len(input.MemoryPayment) > 0) {
			return fmt.Errorf("%w: unused input payload for ability", tcgErrors.ErrInvalidViewHandle)
		}
		return nil
	}
	_, actionExists := g.state.Knowledge.Actions[player.UID][input.Action]
	_, materializationExists := g.state.Knowledge.Materializations[player.UID][input.Action]
	_, attackExists := g.state.Knowledge.Attacks[player.UID][input.Action]
	_, wieldExists := g.state.Knowledge.Wields[player.UID][input.Action]
	if materializationExists {
		if len(input.Reserve) > 0 {
			return fmt.Errorf("%w: unused input payload for materialization", tcgErrors.ErrInvalidViewHandle)
		}
		return nil
	}
	if actionExists || attackExists || wieldExists {
		if len(input.Reserve) > 0 || len(input.MemoryPayment) > 0 {
			return fmt.Errorf("%w: unused input payload for action", tcgErrors.ErrInvalidViewHandle)
		}
	}
	return nil
}

func (g *Game) recordReplayStep(player *model.Player, input Input) {
	g.replay.Steps = append(
		g.replay.Steps,
		ReplayStep{
			Player: player,
			Input:  input,
		},
	)
	g.replay.Steps[len(g.replay.Steps)-1].StateHash = g.StateHash()
}

// Replay 複製初始狀態與頂層玩家、步驟切片後回傳。
// 玩家指標與步驟內的參考資料仍為淺拷貝，呼叫端不應修改其內容。
func (g *Game) Replay() Replay {
	replay := g.replay
	if replay.InitialState != nil {
		state := cloneGameState(*replay.InitialState)
		replay.InitialState = &state
	}
	replay.InitialPlayers = append(
		[]*model.Player(nil),
		replay.InitialPlayers...,
	)
	replay.Steps = append(
		[]ReplayStep(nil),
		g.replay.Steps...,
	)
	return replay
}

func (g *Game) hasPlayer(player *model.Player) bool {
	return lo.ContainsBy(
		g.players,
		func(candidate *model.Player) bool {
			return samePlayer(candidate, player)
		},
	)
}

func (g *Game) otherPlayer(player *model.Player) *model.Player {
	for _, candidate := range g.players {
		if !samePlayer(candidate, player) {
			return candidate
		}
	}
	return nil
}

func samePlayer(first, second *model.Player) bool {
	if first == nil || second == nil {
		return first == second
	}
	return first.UID == second.UID
}
