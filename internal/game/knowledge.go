package game

import (
	"fmt"
	"go-tcg/internal/constants"
	"go-tcg/internal/model"
	tcgErrors "go-tcg/internal/tcg_errors"
	"maps"
	"sort"
	"strconv"
	"strings"
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
			constants.ViewHandleSubjectActionConcede,
		)
		actions[handle] = constants.ActionConcede
		if g.state.Knowledge.Choice == nil {
			switch {
			case samePlayer(player, g.state.Scheduler.OpportunityHolder):
				handle := g.newViewHandle(
					player,
					constants.ViewHandleSubjectActionPass,
				)
				actions[handle] = constants.ActionPass
				for _, card := range g.legalActionCards(player) {
					handle := g.newViewHandle(
						player,
						constants.ViewHandleSubjectActionActivatePrefix+string(card),
					)
					activations[handle] = card
				}
				for _, attacker := range g.legalAttackers(player) {
					handle := g.newViewHandle(player, constants.ViewHandleSubjectActionAttackPrefix+string(attacker))
					attacks[handle] = attacker
				}
				for _, weapon := range g.legalWeapons(player) {
					if !g.canWield(player, weapon) {
						continue
					}
					handle := g.newViewHandle(player, constants.ViewHandleSubjectActionWieldPrefix+string(weapon))
					wields[handle] = weapon
				}
				for _, source := range g.legalCardistries(player) {
					handle := g.newViewHandle(player, constants.ViewHandleSubjectActionCardistryPrefix+string(source))
					cardistries[handle] = source
				}
				for _, source := range g.legalObjectAbilities(player) {
					handle := g.newViewHandle(player, constants.ViewHandleSubjectActionAbilityPrefix+string(source))
					objectAbilities[handle] = source
				}
			case samePlayer(player, g.state.Scheduler.TurnPlayer) && g.state.Scheduler.Phase == constants.PhaseMaterialize:
				for _, card := range g.legalMaterializations(player) {
					handle := g.newViewHandle(
						player,
						constants.ViewHandleSubjectMaterializePrefix+string(card),
					)
					materializations[handle] = card
				}
				handle := g.newViewHandle(
					player,
					constants.ViewHandleSubjectSkipMaterialize,
				)
				actions[handle] = constants.ActionSkipMaterialize
			}
		}
		if g.state.Knowledge.Choice != nil && ((g.state.AbilityChoice != nil && g.state.AbilityChoice.CanPass && samePlayer(g.state.AbilityChoice.Instance.Controller, player)) || (g.state.Knowledge.VeritaCost != nil && g.state.Knowledge.Choice.CanPass && samePlayer(g.state.Knowledge.VeritaCost.Controller, player))) {
			handle := g.newViewHandle(
				player,
				constants.ViewHandleSubjectActionPass,
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

// legalActions 將引擎目前允許的所有行動投影為指定玩家可提交的穩定編號順序。
// 輸入為檢視玩家；輸出為先按行動種類、再按不透明 handle 排序的合法行動，無副作用。
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
				Handle:         handle,
				Kind:           constants.ActionActivate,
				CardName:       g.state.Entities[entityID(card)].Name,
				ReserveCost:    g.activationReserveCost(player, card),
				ReserveOptions: g.visibleReserveCards(player, card),
			},
		)
	}
	for handle, attacker := range g.state.Knowledge.Attacks[player.UID] {
		card, exists := g.cardForObject(attacker)
		if !exists {
			continue
		}
		legalActions = append(
			legalActions,
			LegalAction{
				Handle:   handle,
				Kind:     constants.ActionAttack,
				CardName: g.cardName(card.ID),
			},
		)
	}
	for handle, weapon := range g.state.Knowledge.Wields[player.UID] {
		legalActions = append(legalActions, LegalAction{
			Handle:   handle,
			Kind:     constants.ActionWield,
			CardName: g.state.Entities[entityID(g.state.Objects[weapon].Card)].Name,
		})
	}
	for handle, source := range g.state.Knowledge.Cardistries[player.UID] {
		card := g.state.Objects[source].Card
		baseCost, _ := g.cardistryBaseCost(card)
		cost := g.cardistryCost(player, baseCost)
		memoryCount := len(g.state.Zones[player.UID].Memory)
		floatingMemoryRequired := max(cost-memoryCount, 0)
		legalActions = append(
			legalActions,
			LegalAction{
				Handle:                 handle,
				Kind:                   constants.ActionActivate,
				CardName:               g.cardName(card),
				FloatingMemoryRequired: floatingMemoryRequired,
				FloatingMemoryOptions:  g.visibleFloatingMemory(player),
			},
		)
	}
	for handle, source := range g.state.Knowledge.ObjectAbilities[player.UID] {
		legalActions = append(legalActions, LegalAction{Handle: handle, Kind: constants.ActionActivate, CardName: g.state.Entities[entityID(g.state.Objects[source].Card)].Name})
	}
	for index := range legalActions {
		legalActions[index].HeuristicRank = g.heuristicRank(
			player,
			legalActions[index],
		)
	}
	sort.Slice(
		legalActions,
		func(first, second int) bool {
			if legalActions[first].Kind == legalActions[second].Kind {
				return legalActions[first].Handle < legalActions[second].Handle
			}
			return legalActions[first].Kind < legalActions[second].Kind
		},
	)
	return legalActions
}

// activationReserveCost 回傳 PlayerView 與 Input payload 契約共同使用的實際 Reserve 張數。
// 輸入為啟動玩家與手牌卡牌；輸出為一般 reserve cost，或 Verita 可用替代費用時的零，無副作用。
func (g *Game) activationReserveCost(player *model.Player, card cardInstanceID) int {
	alternativeCostCards := g.veritaAlternativeCostCards(player)
	if g.state.Cards[card].Definition == veritaCardID && len(alternativeCostCards) > 0 {
		return 0
	}
	return g.actionReserveCost(player, card)
}

// visibleReserveCards 投影可支付指定行動 reserve cost 的其他手牌。
// 輸入為玩家與正在啟動的手牌；輸出為可選的不透明 handles，無副作用且不暴露對手資訊。
func (g *Game) visibleReserveCards(player *model.Player, source cardInstanceID) []VisibleCard {
	zones := g.state.Zones[player.UID]
	optionCapacity := len(zones.Hand)
	options := make([]VisibleCard, 0, optionCapacity)
	for _, card := range zones.Hand {
		if card == source {
			continue
		}
		handle, exists := g.state.Knowledge.Cards[player.UID][entityID(card)]
		if !exists {
			continue
		}
		options = append(
			options,
			VisibleCard{
				Handle: handle,
				Name:   g.cardName(card),
			},
		)
	}
	return options
}

// visibleFloatingMemory 投影目前可作為 Cardistry Floating Memory 的已追蹤墓地卡牌。
// 輸入為付款玩家；輸出為引擎已驗證可選的卡牌 handle，無副作用且不重新判定 Cardistry 成本。
func (g *Game) visibleFloatingMemory(player *model.Player) []VisibleCard {
	cards := g.floatingMemoryCards(player)
	options := make([]VisibleCard, 0, len(cards))
	for _, card := range cards {
		handle, exists := g.state.Knowledge.Cards[player.UID][entityID(card)]
		if !exists {
			continue
		}
		options = append(
			options,
			VisibleCard{
				Handle: handle,
				Name:   g.cardName(card),
			},
		)
	}
	sort.Slice(
		options,
		func(first, second int) bool {
			return options[first].Handle < options[second].Handle
		},
	)
	return options
}

// heuristicRank 依公開 Champion 狀態與已合法的 action 建立 bot 可消費的固定戰術優先級。
// 輸入為決策玩家與合法 action；輸出為越小越優先的 rank，無副作用且不讀取隱藏區域資料。
func (g *Game) heuristicRank(player *model.Player, action LegalAction) int {
	if action.Kind == constants.ActionAttack && g.attackWinsGame(player, action.Handle) {
		return 0
	}
	switch action.Kind {
	case constants.ActionAttack:
		return 1
	case constants.ActionActivate:
		if action.CardName == "Red Hare, Unrivaled Stallion" {
			return 2
		}
		if action.CardName == "Duchess, Six of Hearts" {
			return 2
		}
		if action.FloatingMemoryOptions != nil {
			return 2
		}
		return 3
	case constants.ActionWield:
		return 3
	case constants.ActionMaterialize:
		return 4
	case constants.ActionSkipMaterialize:
		return 5
	case constants.ActionPass:
		return 5
	case constants.ActionConcede:
		return 6
	default:
		return 7
	}
}

// attackWinsGame 判斷指定合法攻擊是否能以公開攻擊力擊敗任一可攻擊對方 Champion。
// 輸入為攻擊玩家與 action handle；輸出為是否有立即獲勝目標，無副作用且僅檢查公開場上物件。
func (g *Game) attackWinsGame(player *model.Player, handle ViewHandle) bool {
	attacker, exists := g.state.Knowledge.Attacks[player.UID][handle]
	if !exists {
		return false
	}
	power := g.characteristicsFor(attacker).Power
	for _, target := range g.attackTargets(player, attacker) {
		for _, opponent := range g.players {
			champion, championExists := g.state.Champions[opponent.UID]
			if !championExists || samePlayer(opponent, player) || champion.ID != target {
				continue
			}
			if power >= g.characteristicsFor(champion.ID).Life-champion.Damage {
				return true
			}
		}
	}
	return false
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
				Damage:   champion.Damage,
				Rested:   champion.Rested,
				Taunt:    champion.TauntUntilTurn > g.state.Scheduler.TurnNumber,
			},
		)
	}
	return champions
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

// visibleHand 投影指定玩家目前在 Hand zone 且仍具追蹤權的卡牌。
// 輸入為玩家；輸出為該玩家可見的手牌與既有不透明 handle，無副作用且不讀取對手手牌。
func (g *Game) visibleHand(player *model.Player) []VisibleCard {
	zones := g.state.Zones[player.UID]
	hand := make([]VisibleCard, 0, len(zones.Hand))
	for _, card := range zones.Hand {
		handle, exists := g.state.Knowledge.Cards[player.UID][entityID(card)]
		if !exists {
			continue
		}
		hand = append(
			hand,
			VisibleCard{
				Handle: handle,
				Name:   g.cardName(card),
			},
		)
	}
	return hand
}

// visibleField 投影所有公開場上物件，並複製可變欄位以隔離呼叫端修改。
// 輸入不含外部參數；輸出為依玩家座位與進場順序排序的公開物件，無副作用且不暴露 object ID。
func (g *Game) visibleField() []VisibleFieldObject {
	type orderedFieldObject struct {
		object      VisibleFieldObject
		id          objectID
		playerOrder int
		playOrder   uint64
	}
	playerOrders := make(map[string]int, len(g.players))
	for index, player := range g.players {
		playerOrders[player.UID] = index
	}
	objects := make([]orderedFieldObject, 0, len(g.state.Objects))
	for id, object := range g.state.Objects {
		objects = append(
			objects,
			orderedFieldObject{
				object: VisibleFieldObject{
					Owner:    object.Owner,
					CardName: g.cardName(object.Card),
					Types:    append([]string(nil), object.Types...),
					Rested:   object.Rested,
					Damage:   object.Damage,
					Counters: maps.Clone(object.Counters),
				},
				id:          id,
				playerOrder: playerOrders[object.Owner.UID],
				playOrder:   fieldPlayOrder(id),
			},
		)
	}
	sort.Slice(
		objects,
		func(first, second int) bool {
			if objects[first].playerOrder != objects[second].playerOrder {
				return objects[first].playerOrder < objects[second].playerOrder
			}
			if objects[first].playOrder != objects[second].playOrder {
				return objects[first].playOrder < objects[second].playOrder
			}
			return objects[first].id < objects[second].id
		},
	)
	field := make([]VisibleFieldObject, 0, len(objects))
	for _, object := range objects {
		field = append(
			field,
			object.object,
		)
	}
	return field
}

// fieldPlayOrder 從由 putFieldObject 建立的物件 ID 取回單局遞增進場順序。
// 輸入為場上物件 ID；輸出為建立順序，格式不符時以零作為穩定排序的後備值，無副作用。
func fieldPlayOrder(id objectID) uint64 {
	separator := strings.LastIndex(string(id), ":")
	if separator == -1 || separator == len(id)-1 {
		return 0
	}
	order, err := strconv.ParseUint(string(id[separator+1:]), 10, 64)
	if err != nil {
		return 0
	}
	return order
}

// visibleEffectsStack 投影公開效果堆疊，保留由底到頂的引擎順序。
// 輸入不含外部參數；輸出為每個 StackItem 的種類、控制者與公開來源名稱，無副作用且不暴露內部 ID。
func (g *Game) visibleEffectsStack() []VisibleEffectStackItem {
	stack := make([]VisibleEffectStackItem, 0, len(g.state.EffectsStack))
	for _, item := range g.state.EffectsStack {
		source := item.Source
		if source == "" {
			source = item.SourceLKI
		}
		stack = append(
			stack,
			VisibleEffectStackItem{
				Kind:       string(item.Kind),
				Controller: item.Controller,
				SourceName: g.cardName(source),
			},
		)
	}
	return stack
}

// cardName 取得卡牌實例對應的公開名稱，缺少實例時回傳空字串。
// 輸入為內部卡牌識別；輸出為僅供既有可見投影使用的名稱，無副作用且不將識別本身交給呼叫端。
func (g *Game) cardName(card cardInstanceID) string {
	return g.state.Entities[entityID(card)].Name
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
	choices := make([]VisibleChoice, 0, len(choice.Options))
	for handle, subject := range choice.Options {
		options = append(options, handle)
		cardName := ""
		if _, visible := g.state.Knowledge.Cards[player.UID][subject]; visible {
			cardName = g.state.Entities[subject].Name
		} else if card, objectExists := g.cardForObject(objectID(subject)); objectExists {
			cardName = g.state.Entities[entityID(card.ID)].Name
		}
		choices = append(
			choices,
			VisibleChoice{
				Handle:   handle,
				CardName: cardName,
				HeuristicRank: g.choiceHeuristicRank(
					player,
					subject,
				),
			},
		)
	}
	sort.Slice(
		options,
		func(first, second int) bool {
			return options[first] < options[second]
		},
	)
	sort.Slice(
		choices,
		func(first, second int) bool {
			return choices[first].Handle < choices[second].Handle
		},
	)
	return &PendingChoice{
		Options: options,
		Choices: choices,
		CanPass: choice.CanPass,
	}
}

// choiceHeuristicRank 對攻擊目標優先選取可立即擊敗的公開 Champion，其餘 choice 保持同分。
// 輸入為選擇玩家與已合法的選項 subject；輸出為越小越優先的 rank，無副作用且不讀取隱藏區域資料。
func (g *Game) choiceHeuristicRank(player *model.Player, subject entityID) int {
	attack := g.state.Knowledge.Attack
	if attack == nil || !samePlayer(attack.Controller, player) {
		return 0
	}
	for _, opponent := range g.players {
		champion, exists := g.state.Champions[opponent.UID]
		if !exists || samePlayer(opponent, player) || champion.ID != objectID(subject) {
			continue
		}
		if g.characteristicsFor(attack.Attacker).Power >= g.characteristicsFor(champion.ID).Life-champion.Damage {
			return 0
		}
	}
	return 2
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
