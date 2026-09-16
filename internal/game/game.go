package game

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go-tcg/internal/constants"
	"go-tcg/internal/model"
	tcgErrors "go-tcg/internal/tcg_errors"

	"github.com/samber/lo"
)

type Input struct {
	Revision uint64     `json:"revision"`
	Action   ViewHandle `json:"action"`
	Choice   ViewHandle `json:"choice"`
	// Reserve 只接受目前玩家可見的手牌 handles，並由 activation 的 ReserveCost 決定精確張數。
	Reserve        []ViewHandle `json:"reserve,omitempty"`
	FloatingMemory []ViewHandle `json:"floating_memory,omitempty"`
}

type ViewHandle string

type LegalAction struct {
	Handle   ViewHandle           `json:"handle"`
	Kind     constants.ActionKind `json:"kind"`
	CardName string               `json:"card_name,omitempty"`
	// ReserveCost 與 ReserveOptions 定義 activation 必須提交的手牌付款張數與可選 handles。
	ReserveCost    int           `json:"reserve_cost,omitempty"`
	ReserveOptions []VisibleCard `json:"reserve_options,omitempty"`
	// FloatingMemoryRequired 指出本次 Cardistry 付款至少必須使用的 Floating Memory 張數。
	FloatingMemoryRequired int           `json:"floating_memory_required,omitempty"`
	FloatingMemoryOptions  []VisibleCard `json:"floating_memory_options,omitempty"`
	HeuristicRank          int           `json:"heuristic_rank"`
}

// VisibleChoice 提供 PendingChoice 選項的玩家可見描述與啟發式優先級。
// Handle 是提交用的不透明識別；其餘欄位只由既有 PlayerView 可見資料導出，不包含內部 ID。
type VisibleChoice struct {
	Handle        ViewHandle `json:"handle"`
	CardName      string     `json:"card_name,omitempty"`
	HeuristicRank int        `json:"heuristic_rank"`
}

type VisibleChampion struct {
	Owner    *model.Player `json:"owner"`
	CardName string        `json:"card_name"`
	Power    int           `json:"power"`
	Life     int           `json:"life"`
	Damage   int           `json:"damage"`
	Rested   bool          `json:"rested"`
	Taunt    bool          `json:"taunt"`
}

type VisibleCard struct {
	Handle ViewHandle `json:"handle"`
	Name   string     `json:"name"`
}

type VisibleFieldObject struct {
	Owner    *model.Player  `json:"owner"`
	CardName string         `json:"card_name"`
	Types    []string       `json:"types"`
	Rested   bool           `json:"rested"`
	Damage   int            `json:"damage"`
	Counters map[string]int `json:"counters,omitempty"`
}

type VisibleEffectStackItem struct {
	Kind       string        `json:"kind"`
	Controller *model.Player `json:"controller"`
	SourceName string        `json:"source_name"`
}

type VisibleEvent struct {
	Kind     string `json:"kind"`
	CardName string `json:"card_name"`
}

type PendingChoice struct {
	Options []ViewHandle    `json:"options"`
	Choices []VisibleChoice `json:"choices"`
	CanPass bool            `json:"can_pass,omitempty"`
}

type PlayerView struct {
	Revision          uint64                   `json:"revision"`
	Finished          bool                     `json:"finished"`
	Winner            *model.Player            `json:"winner,omitempty"`
	Diagnostic        string                   `json:"diagnostic,omitempty"`
	TurnPlayer        *model.Player            `json:"turn_player,omitempty"`
	TurnNumber        uint64                   `json:"turn_number,omitempty"`
	Phase             Phase                    `json:"phase,omitempty"`
	OpportunityHolder *model.Player            `json:"opportunity_holder,omitempty"`
	DecisionPlayer    *model.Player            `json:"decision_player,omitempty"`
	Champions         []VisibleChampion        `json:"champions,omitempty"`
	Hand              []VisibleCard            `json:"hand"`
	Field             []VisibleFieldObject     `json:"field"`
	EffectsStack      []VisibleEffectStackItem `json:"effects_stack"`
	Cards             []VisibleCard            `json:"cards"`
	VisibleEvents     []VisibleEvent           `json:"visible_events"`
	LegalActions      []LegalAction            `json:"legal_actions"`
	PendingChoice     *PendingChoice           `json:"pending_choice,omitempty"`
}

type Game struct {
	versions Versions
	players  []*model.Player
	state    gameState
	replay   Replay
}

type gameState struct {
	Revision           uint64
	Finished           bool
	Winner             *model.Player
	Diagnostic         string
	PRNG               prngState
	Knowledge          knowledgeState
	Entities           map[entityID]knowledgeEntity
	Cards              map[cardInstanceID]cardInstance
	Zones              map[string]playerZones
	Champions          map[string]championObject
	Objects            map[objectID]fieldObject
	EffectSources      []cardInstanceID
	EffectsStack       []effectStackItem
	ContinuousEffects  []continuousEffect
	AbilityChoice      *abilityChoice
	ReplacementEffects []replacementEffect
	ReplacementChoice  *replacementChoice
	CardistryUsed      map[objectID]bool
	CardistryDiscounts map[string]int
	Scheduler          schedulerFrame
	Events             []eventBatch
	NextHandle         uint64
	NextEvent          uint64
	NextEffect         uint64
	NextAbility        uint64
	NextObject         uint64
}

type prngState struct {
	Seed   uint64 `json:"seed"`
	Cursor uint64 `json:"cursor"`
}

type canonicalState struct {
	SchemaVersion      int                             `json:"schema_version"`
	Versions           Versions                        `json:"versions"`
	Players            []*model.Player                 `json:"players"`
	Revision           uint64                          `json:"revision"`
	Finished           bool                            `json:"finished"`
	Winner             *model.Player                   `json:"winner"`
	Diagnostic         string                          `json:"diagnostic,omitempty"`
	PRNG               prngState                       `json:"prng"`
	Knowledge          knowledgeState                  `json:"knowledge"`
	Entities           map[entityID]knowledgeEntity    `json:"entities"`
	Cards              map[cardInstanceID]cardInstance `json:"cards"`
	Zones              map[string]playerZones          `json:"zones"`
	Champions          map[string]championObject       `json:"champions"`
	Objects            map[objectID]fieldObject        `json:"objects"`
	EffectSources      []cardInstanceID                `json:"effect_sources,omitempty"`
	EffectsStack       []effectStackItem               `json:"effects_stack,omitempty"`
	ContinuousEffects  []continuousEffect              `json:"continuous_effects,omitempty"`
	AbilityChoice      *abilityChoice                  `json:"ability_choice,omitempty"`
	ReplacementEffects []replacementEffect             `json:"replacement_effects,omitempty"`
	ReplacementChoice  *replacementChoice              `json:"replacement_choice,omitempty"`
	CardistryUsed      map[objectID]bool               `json:"cardistry_used,omitempty"`
	CardistryDiscounts map[string]int                  `json:"cardistry_discounts,omitempty"`
	Scheduler          schedulerFrame                  `json:"scheduler"`
	Events             []eventBatch                    `json:"events"`
	NextHandle         uint64                          `json:"next_handle"`
	NextEvent          uint64                          `json:"next_event"`
	NextEffect         uint64                          `json:"next_effect"`
	NextAbility        uint64                          `json:"next_ability"`
	NextObject         uint64                          `json:"next_object"`
}

// NewGame 建立引擎的空白狀態並固定亂數種子與版本。
// 玩家牌組、開局事件與標準回合排程由 setup 層配置。
func NewGame(seed uint64) *Game {
	versions := currentVersions()
	game := &Game{
		versions: versions,
		players: []*model.Player{
			model.PlayerOne,
			model.PlayerTwo,
		},
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
				Seed: seed,
			},
		},
		replay: Replay{
			FormatVersion: constants.ReplayFormatVersion,
			Versions:      versions,
			InitialSeed:   seed,
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
// 等待選擇時只接受該選擇、投降，以及可略過能力的 pass；Reserve 只供手牌 activation，Floating Memory 只供 Cardistry 付款。
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
	if len(input.FloatingMemory) > 0 {
		if _, exists := g.state.Knowledge.Cardistries[player.UID][input.Action]; !exists {
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
					source, cardistryExists := g.state.Knowledge.Cardistries[player.UID][input.Action]
					if cardistryExists {
						if err := g.activateCardistry(
							player,
							source,
							input.FloatingMemory,
						); err != nil {
							return fmt.Errorf("activate Cardistry: %w", err)
						}
					} else {
						object, objectAbilityExists := g.state.Knowledge.ObjectAbilities[player.UID][input.Action]
						if objectAbilityExists {
							if err := g.beginObjectAbility(player, object); err != nil {
								return fmt.Errorf("begin object ability: %w", err)
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
		if input.Action != "" || len(input.Reserve) > 0 || len(input.FloatingMemory) > 0 {
			return fmt.Errorf("%w: unused input payload for choice", tcgErrors.ErrInvalidViewHandle)
		}
		return nil
	}
	if input.Action == "" {
		return fmt.Errorf("%w: missing action or choice", tcgErrors.ErrInvalidViewHandle)
	}
	if card, exists := g.state.Knowledge.Activations[player.UID][input.Action]; exists {
		if len(input.FloatingMemory) > 0 || len(input.Reserve) != g.activationReserveCost(player, card) {
			return fmt.Errorf("%w: unused input payload for activation", tcgErrors.ErrInvalidViewHandle)
		}
		return nil
	}
	if _, exists := g.state.Knowledge.Cardistries[player.UID][input.Action]; exists {
		if len(input.Reserve) > 0 {
			return fmt.Errorf("%w: unused input payload for Cardistry", tcgErrors.ErrInvalidViewHandle)
		}
		return nil
	}
	_, actionExists := g.state.Knowledge.Actions[player.UID][input.Action]
	_, materializationExists := g.state.Knowledge.Materializations[player.UID][input.Action]
	_, attackExists := g.state.Knowledge.Attacks[player.UID][input.Action]
	_, wieldExists := g.state.Knowledge.Wields[player.UID][input.Action]
	_, abilityExists := g.state.Knowledge.ObjectAbilities[player.UID][input.Action]
	if actionExists || materializationExists || attackExists || wieldExists || abilityExists {
		if len(input.Reserve) > 0 || len(input.FloatingMemory) > 0 {
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

// PlayerView 投影指定玩家可見的牌、事件、合法行動與待選項目。
// 提交行動須使用此視圖的 revision 與 handle；未知玩家回傳 ErrUnknownPlayer。
func (g *Game) PlayerView(player *model.Player) (PlayerView, error) {
	if !g.hasPlayer(player) {
		return PlayerView{}, fmt.Errorf("%w %q", tcgErrors.ErrUnknownPlayer, player)
	}
	return PlayerView{
		Revision:          g.state.Revision,
		Finished:          g.state.Finished,
		Winner:            g.state.Winner,
		Diagnostic:        g.state.Diagnostic,
		TurnPlayer:        g.state.Scheduler.TurnPlayer,
		TurnNumber:        g.state.Scheduler.TurnNumber,
		Phase:             g.state.Scheduler.Phase,
		OpportunityHolder: g.state.Scheduler.OpportunityHolder,
		DecisionPlayer:    g.decisionPlayer(),
		Champions: g.visibleChampions(
			player,
		),
		Hand: g.visibleHand(
			player,
		),
		Field:        g.visibleField(),
		EffectsStack: g.visibleEffectsStack(),
		Cards: g.visibleCards(
			player,
		),
		VisibleEvents: g.visibleEvents(
			player,
		),
		LegalActions: g.legalActions(
			player,
		),
		PendingChoice: g.pendingChoice(
			player,
		),
	}, nil
}

// decisionPlayer 回傳目前依 scheduler 或 PendingChoice 應取得輸入的玩家。
// 輸入為目前單局狀態；輸出為待決玩家或 nil，無副作用且不透露任何私人卡牌資訊。
func (g *Game) decisionPlayer() *model.Player {
	if g.state.Knowledge.Choice != nil {
		return g.state.Knowledge.Choice.Actor
	}
	if g.state.Finished {
		return nil
	}
	if g.state.Scheduler.OpportunityHolder != nil {
		return g.state.Scheduler.OpportunityHolder
	}
	if g.state.Scheduler.Phase == PhaseMaterialize {
		return g.state.Scheduler.TurnPlayer
	}
	return nil
}

// StateHash 對 canonicalState 明列的版本、亂數進度、知識與對局欄位計算 SHA-256，供 replay 逐步比對。
// JSON 編碼失敗代表內部狀態無法序列化，會 panic。
func (g *Game) StateHash() string {
	canonical := canonicalState{
		SchemaVersion:      constants.CanonicalStateSchemaVersion,
		Versions:           g.versions,
		Players:            g.players,
		Revision:           g.state.Revision,
		Finished:           g.state.Finished,
		Winner:             g.state.Winner,
		Diagnostic:         g.state.Diagnostic,
		PRNG:               g.state.PRNG,
		Knowledge:          g.state.Knowledge,
		Entities:           g.state.Entities,
		Cards:              g.state.Cards,
		Zones:              g.state.Zones,
		Champions:          g.state.Champions,
		Objects:            g.state.Objects,
		EffectSources:      g.state.EffectSources,
		EffectsStack:       g.state.EffectsStack,
		ContinuousEffects:  g.state.ContinuousEffects,
		AbilityChoice:      g.state.AbilityChoice,
		ReplacementEffects: g.state.ReplacementEffects,
		ReplacementChoice:  g.state.ReplacementChoice,
		Scheduler:          g.state.Scheduler,
		Events:             g.state.Events,
		NextHandle:         g.state.NextHandle,
		NextEvent:          g.state.NextEvent,
	}
	state, err := json.Marshal(canonical)
	if err != nil {
		panic(err)
	}
	sum := sha256.Sum256(state)
	return hex.EncodeToString(sum[:])
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
