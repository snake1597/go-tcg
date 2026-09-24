package game

import "go-tcg/internal/model"

// cardSelectionSpec 只描述從哪個區域選牌及合格條件，不決定選後如何處理卡牌。
type cardSelectionSpec struct {
	Zone                     cardZone
	RequiredType             string
	RequiredSubtype          string
	MinimumCards             int
	MatchPrintedReserveTotal bool
	PrintedReserveTotal      int
}

// cardSelectionCandidates 依來源區域順序回傳符合單張條件的牌，不移動任何牌。
func (g *Game) cardSelectionCandidates(player *model.Player, spec cardSelectionSpec, excluded cardInstanceID) []cardInstanceID {
	candidates := make([]cardInstanceID, 0)
	for _, card := range cardsInZone(g.state.Zones[player.UID], spec.Zone) {
		candidate, exists := g.state.Cards[card]
		if !exists || card == excluded {
			continue
		}
		if spec.RequiredType != "" && !containsString(candidate.Types, spec.RequiredType) {
			continue
		}
		if spec.RequiredSubtype != "" && !g.cardHasSubtype(card, spec.RequiredSubtype) {
			continue
		}
		candidates = append(candidates, card)
	}
	return candidates
}

// validCardSelection 驗證已選牌均在來源區域、彼此不同，並符合數量與總費用條件；無副作用。
func (g *Game) validCardSelection(player *model.Player, spec cardSelectionSpec, excluded cardInstanceID, selected []cardInstanceID) bool {
	if len(selected) < spec.MinimumCards {
		return false
	}
	candidates := g.cardSelectionCandidates(player, spec, excluded)
	seen := make(map[cardInstanceID]bool, len(selected))
	total := 0
	for _, card := range selected {
		if seen[card] || !containsCard(candidates, card) {
			return false
		}
		seen[card] = true
		total += g.printedReserveCost(card)
	}
	return !spec.MatchPrintedReserveTotal || total == spec.PrintedReserveTotal
}

// findCardSelection 搜尋包含既有選牌的第一組完整合格組合；無解時回傳 nil，不移動卡牌。
func (g *Game) findCardSelection(player *model.Player, spec cardSelectionSpec, excluded cardInstanceID, selected []cardInstanceID) []cardInstanceID {
	candidates := g.cardSelectionCandidates(player, spec, excluded)
	var search func(start int, chosen []cardInstanceID) []cardInstanceID
	search = func(start int, chosen []cardInstanceID) []cardInstanceID {
		if g.validCardSelection(player, spec, excluded, chosen) {
			return append([]cardInstanceID(nil), chosen...)
		}
		for index := start; index < len(candidates); index++ {
			if containsCard(chosen, candidates[index]) {
				continue
			}
			next := append(append([]cardInstanceID(nil), chosen...), candidates[index])
			if found := search(index+1, next); len(found) > 0 {
				return found
			}
		}
		return nil
	}
	return search(0, selected)
}

// nextCardSelectionOptions 回傳每張仍可完成合格組合的下一張牌，保留來源區域順序，無副作用。
func (g *Game) nextCardSelectionOptions(player *model.Player, spec cardSelectionSpec, excluded cardInstanceID, selected []cardInstanceID) []cardInstanceID {
	options := make([]cardInstanceID, 0)
	for _, card := range g.cardSelectionCandidates(player, spec, excluded) {
		if containsCard(selected, card) {
			continue
		}
		next := append(append([]cardInstanceID(nil), selected...), card)
		if len(g.findCardSelection(player, spec, excluded, next)) > 0 {
			options = append(options, card)
		}
	}
	return options
}
