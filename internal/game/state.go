package game

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"go-tcg/internal/constants"
	"go-tcg/internal/model"
)

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
