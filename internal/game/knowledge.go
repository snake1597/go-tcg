package game

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"go-tcg/internal/constants"
	"go-tcg/internal/model"
	tcgErrors "go-tcg/internal/tcg_errors"
	"sort"
)

type entityID string

type knowledgeEntity struct {
	Name string `json:"name"`
}

type knowledgeState struct {
	Actions          map[string]map[ViewHandle]constants.ActionKind `json:"actions"`
	Materializations map[string]map[ViewHandle]cardInstanceID       `json:"materializations"`
	Activations      map[string]map[ViewHandle]cardInstanceID       `json:"activations"`
	Attacks          map[string]map[ViewHandle]objectID             `json:"attacks"`
	Wields           map[string]map[ViewHandle]objectID             `json:"wields"`
	Cardistries      map[string]map[ViewHandle]objectID             `json:"cardistries"`
	ObjectAbilities  map[string]map[ViewHandle]objectID             `json:"object_abilities"`
	Cards            map[string]map[entityID]ViewHandle             `json:"cards"`
	Events           map[string][]VisibleEvent                      `json:"events"`
	Choice           *pendingChoice                                 `json:"choice,omitempty"`
	Declaration      *actionDeclaration                             `json:"declaration,omitempty"`
	TriggerOrder     *triggerOrder                                  `json:"trigger_order,omitempty"`
	Attack           *attackDeclaration                             `json:"attack,omitempty"`
	Wield            *wieldDeclaration                              `json:"wield,omitempty"`
	ObjectAbility    *objectAbilityDeclaration                      `json:"object_ability,omitempty"`
	VeritaCost       *veritaAlternativeCostDeclaration              `json:"verita_cost,omitempty"`
}

type pendingChoice struct {
	Actor   *model.Player           `json:"actor"`
	Options map[ViewHandle]entityID `json:"options"`
	CanPass bool                    `json:"can_pass,omitempty"`
}

func (g *Game) initializeKnowledgeState() {
	knowledge := knowledgeState{
		Actions:          make(map[string]map[ViewHandle]constants.ActionKind, len(g.players)),
		Materializations: make(map[string]map[ViewHandle]cardInstanceID, len(g.players)),
		Activations:      make(map[string]map[ViewHandle]cardInstanceID, len(g.players)),
		Attacks:          make(map[string]map[ViewHandle]objectID, len(g.players)),
		Wields:           make(map[string]map[ViewHandle]objectID, len(g.players)),
		Cardistries:      make(map[string]map[ViewHandle]objectID, len(g.players)),
		ObjectAbilities:  make(map[string]map[ViewHandle]objectID, len(g.players)),
		Cards:            make(map[string]map[entityID]ViewHandle, len(g.players)),
		Events:           make(map[string][]VisibleEvent, len(g.players)),
	}
	for _, player := range g.players {
		knowledge.Actions[player.UID] = make(map[ViewHandle]constants.ActionKind)
		knowledge.Materializations[player.UID] = make(map[ViewHandle]cardInstanceID)
		knowledge.Activations[player.UID] = make(map[ViewHandle]cardInstanceID)
		knowledge.Attacks[player.UID] = make(map[ViewHandle]objectID)
		knowledge.Wields[player.UID] = make(map[ViewHandle]objectID)
		knowledge.Cardistries[player.UID] = make(map[ViewHandle]objectID)
		knowledge.ObjectAbilities[player.UID] = make(map[ViewHandle]objectID)
		knowledge.Cards[player.UID] = make(map[entityID]ViewHandle)
		knowledge.Events[player.UID] = []VisibleEvent{}
	}
	g.state.Knowledge = knowledge
	g.refreshLegalActions()
}

// refreshLegalActions 清空並重建各玩家的行動 handle，反映目前行動機會、階段與待選狀態。
// 等待選擇時不提供一般行動；投降與可略過能力的 pass 另行處理。
// 遊戲結束時清空所有行動，但既有卡牌追蹤 handle 不由此函式重建。
func (g *Game) refreshLegalActions() {
	for _, player := range g.players {
		actions := g.state.Knowledge.Actions[player.UID]
		materializations := g.state.Knowledge.Materializations[player.UID]
		activations := g.state.Knowledge.Activations[player.UID]
		attacks := g.state.Knowledge.Attacks[player.UID]
		wields := g.state.Knowledge.Wields[player.UID]
		cardistries := g.state.Knowledge.Cardistries[player.UID]
		objectAbilities := g.state.Knowledge.ObjectAbilities[player.UID]
		clear(actions)
		clear(materializations)
		clear(activations)
		clear(attacks)
		clear(wields)
		clear(cardistries)
		clear(objectAbilities)
		if g.state.Finished {
			continue
		}
		handle := g.newViewHandle(
			player,
			"action:concede",
		)
		actions[handle] = constants.ActionConcede
		if g.state.Knowledge.Choice == nil {
			switch {
			case samePlayer(player, g.state.Scheduler.OpportunityHolder):
				handle := g.newViewHandle(
					player,
					"action:pass",
				)
				actions[handle] = constants.ActionPass
				for _, card := range g.legalActionCards(player) {
					handle := g.newViewHandle(
						player,
						"action:activate:"+string(card),
					)
					activations[handle] = card
				}
				for _, attacker := range g.legalAttackers(player) {
					handle := g.newViewHandle(player, "action:attack:"+string(attacker))
					attacks[handle] = attacker
				}
				for _, weapon := range g.legalWeapons(player) {
					if !g.canWield(player, weapon) {
						continue
					}
					handle := g.newViewHandle(player, "action:wield:"+string(weapon))
					wields[handle] = weapon
				}
				for _, source := range g.legalCardistries(player) {
					handle := g.newViewHandle(player, "action:cardistry:"+string(source))
					cardistries[handle] = source
				}
				for source, object := range g.state.Objects {
					definition := g.state.Cards[object.Card].Definition
					if samePlayer(object.Owner, player) && ((definition == duchessThornesCardID && !object.Rested) || definition == smokeBombsCardID || definition == safeguardAmuletCardID) {
						handle := g.newViewHandle(player, "action:ability:"+string(source))
						objectAbilities[handle] = source
					}
				}
			case samePlayer(player, g.state.Scheduler.TurnPlayer) && g.state.Scheduler.Phase == PhaseMaterialize:
				for _, card := range g.legalChampionMaterializations(player) {
					handle := g.newViewHandle(
						player,
						"action:materialize:"+string(card),
					)
					materializations[handle] = card
				}
				handle := g.newViewHandle(
					player,
					"action:skip-materialize",
				)
				actions[handle] = constants.ActionSkipMaterialize
			}
		}
		if g.state.Knowledge.Choice != nil && ((g.state.AbilityChoice != nil && g.state.AbilityChoice.CanPass && samePlayer(g.state.AbilityChoice.Instance.Controller, player)) || (g.state.Knowledge.VeritaCost != nil && g.state.Knowledge.Choice.CanPass && samePlayer(g.state.Knowledge.VeritaCost.Controller, player))) {
			handle := g.newViewHandle(
				player,
				"action:pass",
			)
			actions[handle] = constants.ActionPass
		}
	}
}

func (g *Game) hasAction(player *model.Player, kind constants.ActionKind) bool {
	for _, candidate := range g.state.Knowledge.Actions[player.UID] {
		if candidate == kind {
			return true
		}
	}
	return false
}

func (g *Game) legalActions(player *model.Player) []LegalAction {
	actions := g.state.Knowledge.Actions[player.UID]
	legalActions := make([]LegalAction, 0, len(actions))
	for handle, kind := range actions {
		legalActions = append(
			legalActions,
			LegalAction{
				Handle: handle,
				Kind:   kind,
			},
		)
	}
	for handle, card := range g.state.Knowledge.Materializations[player.UID] {
		legalActions = append(
			legalActions,
			LegalAction{
				Handle:   handle,
				Kind:     constants.ActionMaterialize,
				CardName: g.state.Entities[entityID(card)].Name,
			},
		)
	}
	for handle, card := range g.state.Knowledge.Activations[player.UID] {
		legalActions = append(
			legalActions,
			LegalAction{
				Handle:   handle,
				Kind:     constants.ActionActivate,
				CardName: g.state.Entities[entityID(card)].Name,
			},
		)
	}
	for handle := range g.state.Knowledge.Attacks[player.UID] {
		legalActions = append(legalActions, LegalAction{
			Handle: handle,
			Kind:   constants.ActionAttack,
		})
	}
	for handle, weapon := range g.state.Knowledge.Wields[player.UID] {
		legalActions = append(legalActions, LegalAction{
			Handle:   handle,
			Kind:     constants.ActionWield,
			CardName: g.state.Entities[entityID(g.state.Objects[weapon].Card)].Name,
		})
	}
	for handle, source := range g.state.Knowledge.Cardistries[player.UID] {
		legalActions = append(
			legalActions,
			LegalAction{
				Handle:   handle,
				Kind:     constants.ActionActivate,
				CardName: g.state.Entities[entityID(g.state.Objects[source].Card)].Name,
			},
		)
	}
	for handle, source := range g.state.Knowledge.ObjectAbilities[player.UID] {
		legalActions = append(legalActions, LegalAction{Handle: handle, Kind: constants.ActionActivate, CardName: g.state.Entities[entityID(g.state.Objects[source].Card)].Name})
	}
	sort.Slice(
		legalActions,
		func(first, second int) bool {
			return legalActions[first].Kind < legalActions[second].Kind
		},
	)
	return legalActions
}

func (g *Game) visibleChampions(_ *model.Player) []VisibleChampion {
	champions := make([]VisibleChampion, 0, len(g.state.Champions))
	for _, owner := range g.players {
		champion, exists := g.state.Champions[owner.UID]
		if !exists {
			continue
		}
		champions = append(
			champions,
			VisibleChampion{
				Owner:    owner,
				CardName: g.state.Entities[entityID(champion.Card)].Name,
				Power:    g.characteristicsFor(champion.ID).Power,
				Life:     g.characteristicsFor(champion.ID).Life,
				Rested:   champion.Rested,
				Taunt:    champion.TauntUntilTurn > g.state.Scheduler.TurnNumber,
			},
		)
	}
	return champions
}

// grantCardTracking 為玩家首次追蹤到的卡牌配置 handle，後續追蹤保留同一 handle。
// 是否應讓玩家追蹤此牌由呼叫端決定，此函式不檢查牌所在區域或可見性。
func (g *Game) grantCardTracking(player *model.Player, card entityID) {
	if _, exists := g.state.Knowledge.Cards[player.UID][card]; exists {
		return
	}
	handle := g.newViewHandle(
		player,
		"card:"+string(card),
	)
	g.state.Knowledge.Cards[player.UID][card] = handle
}

func (g *Game) revokeCardTracking(player *model.Player, card entityID) {
	handle, exists := g.state.Knowledge.Cards[player.UID][card]
	if !exists {
		return
	}
	delete(g.state.Knowledge.Cards[player.UID], card)
	choice := g.state.Knowledge.Choice
	if choice == nil || !samePlayer(choice.Actor, player) {
		return
	}
	delete(choice.Options, handle)
	if len(choice.Options) == 0 {
		g.state.Knowledge.Choice = nil
	}
}

// visibleCards 回傳玩家已取得追蹤 handle 的卡牌，而非直接列出區域內的所有牌。
// 結果按 handle 排序；隱藏或公開卡牌時，呼叫端須同步維護追蹤映射。
func (g *Game) visibleCards(player *model.Player) []VisibleCard {
	cards := g.state.Knowledge.Cards[player.UID]
	visibleCards := make([]VisibleCard, 0, len(cards))
	for card, handle := range cards {
		visibleCards = append(
			visibleCards,
			VisibleCard{
				Handle: handle,
				Name:   g.state.Entities[card].Name,
			},
		)
	}
	sort.Slice(
		visibleCards,
		func(first, second int) bool {
			return visibleCards[first].Handle < visibleCards[second].Handle
		},
	)
	return visibleCards
}

func (g *Game) recordVisibleEvent(player *model.Player, kind string, card entityID) {
	event := VisibleEvent{
		Kind:     kind,
		CardName: g.state.Entities[card].Name,
	}
	g.state.Knowledge.Events[player.UID] = append(
		g.state.Knowledge.Events[player.UID],
		event,
	)
}

func (g *Game) visibleEvents(player *model.Player) []VisibleEvent {
	return append(
		[]VisibleEvent(nil),
		g.state.Knowledge.Events[player.UID]...,
	)
}

func (g *Game) setPendingCardChoice(player *model.Player, card entityID) {
	handle, exists := g.state.Knowledge.Cards[player.UID][card]
	if !exists {
		return
	}
	g.state.Knowledge.Choice = &pendingChoice{
		Actor: player,
		Options: map[ViewHandle]entityID{
			handle: card,
		},
	}
}

func (g *Game) pendingChoice(player *model.Player) *PendingChoice {
	choice := g.state.Knowledge.Choice
	if choice == nil || !samePlayer(choice.Actor, player) {
		return nil
	}
	options := make([]ViewHandle, 0, len(choice.Options))
	for handle := range choice.Options {
		options = append(options, handle)
	}
	sort.Slice(
		options,
		func(first, second int) bool {
			return options[first] < options[second]
		},
	)
	return &PendingChoice{
		Options: options,
		CanPass: choice.CanPass,
	}
}

// submitChoice 先驗證選擇者與待選 handle，再依目前宣告、觸發排序或能力狀態分派。
// 成功處理後更新 revision 與合法行動；replay 由外層 Submit 記錄。
func (g *Game) submitChoice(player *model.Player, input Input) error {
	choice := g.state.Knowledge.Choice
	if choice == nil || !samePlayer(choice.Actor, player) {
		return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, input.Choice)
	}
	subject, exists := choice.Options[input.Choice]
	if !exists {
		return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, input.Choice)
	}
	if g.state.Knowledge.Declaration != nil {
		if err := g.submitActionDeclarationChoice(player, subject); err != nil {
			return err
		}
		g.advanceKnowledgeRevision()
		return nil
	}
	if g.state.Knowledge.TriggerOrder != nil {
		if err := g.submitTriggerOrderChoice(player, input.Choice); err != nil {
			return err
		}
		g.advanceKnowledgeRevision()
		return nil
	}
	if g.state.Knowledge.Attack != nil {
		if err := g.submitAttackChoice(player, subject); err != nil {
			return err
		}
		g.advanceKnowledgeRevision()
		return nil
	}
	if g.state.Knowledge.Wield != nil {
		if err := g.submitWieldChoice(player, subject); err != nil {
			return err
		}
		g.advanceKnowledgeRevision()
		return nil
	}
	if g.state.Knowledge.ObjectAbility != nil {
		if err := g.submitObjectAbilityChoice(player, objectID(subject)); err != nil {
			return err
		}
		g.advanceKnowledgeRevision()
		return nil
	}
	if g.state.Knowledge.VeritaCost != nil {
		if err := g.submitVeritaAlternativeCostChoice(
			player,
			cardInstanceID(subject),
		); err != nil {
			return err
		}
		g.advanceKnowledgeRevision()
		return nil
	}
	if g.state.ReplacementChoice != nil {
		if err := g.submitReplacementChoice(player, objectID(subject)); err != nil {
			return err
		}
		g.advanceKnowledgeRevision()
		return nil
	}
	if g.state.AbilityChoice != nil {
		continuation := g.state.AbilityChoice
		continuation.Instance.Target = objectID(subject)
		continuation.Instance.Operations = continuation.Operations
		g.state.AbilityChoice = nil
		g.state.Knowledge.Choice = nil
		g.pushAbility(continuation.Instance)
		g.grantOpportunity(player)
		g.advanceKnowledgeRevision()
		return nil
	}
	g.state.Knowledge.Choice = nil
	g.advanceKnowledgeRevision()
	return nil
}

// advanceKnowledgeRevision 增加 revision 並重建合法行動，使舊視圖的輸入失效。
// 此函式不記錄 replay，也不重新配置既有卡牌追蹤 handle。
func (g *Game) advanceKnowledgeRevision() {
	g.state.Revision++
	g.refreshLegalActions()
}

// newViewHandle 以種子、遞增序號、玩家與 subject 產生可重播的 handle，並推進 NextHandle。
// handle 的玩家歸屬與有效性由知識映射驗證，不以雜湊本身作為授權依據。
func (g *Game) newViewHandle(player *model.Player, subject string) ViewHandle {
	g.state.NextHandle++
	value := fmt.Sprintf(
		"view-handle-v1:%d:%d:%s:%s",
		g.state.PRNG.Seed,
		g.state.NextHandle,
		player,
		subject,
	)
	valueBytes := []byte(value)
	sum := sha256.Sum256(valueBytes)
	encoded := hex.EncodeToString(sum[:])
	return ViewHandle(encoded)
}
