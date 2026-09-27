package game

import (
	"sort"

	"go-tcg/internal/model"
)

// selectTargets 列舉公開場域中的合法單位，並保留固定順序。
// 輸入為已編譯 selector；輸出為 visibility-safe 的目標 ID，副作用為零。
func (g *Game) selectTargets(selector targetSelector) []objectID {
	switch selector.kind {
	case selectorUnits:
		return g.legalTargets()
	default:
		return nil
	}
}

// targetMatchesSelector 在宣告提交時重新檢查目標仍屬 selector 的合法選項。
// 輸入為 selector 與目標 ID；輸出為是否合法，副作用為零。
func (g *Game) targetMatchesSelector(selector targetSelector, target objectID) bool {
	for _, candidate := range g.selectTargets(selector) {
		if candidate == target {
			return true
		}
	}
	return false
}

// selectObjects 列舉控制者場上符合公開特徵的物件，順序由物件 ID 決定。
// 輸入為 selector 與控制者；輸出為物件 ID 序列，副作用為零。
func (g *Game) selectObjects(selector selectorKind, controller *model.Player) []objectID {
	objects := make([]objectID, 0)
	for id, object := range g.state.Objects {
		if selector == selectorControlledSuited && samePlayer(object.Owner, controller) && g.cardHasSubtype(object.Card, "SUITED") {
			objects = append(objects, id)
		}
	}
	sort.Slice(
		objects,
		func(first, second int) bool {
			return objects[first] < objects[second]
		},
	)
	return objects
}

// evaluateValue 在效果結算當下依控制者的目前場面求出整數值。
// 輸入為已驗證 expression 與控制者；輸出為整數，副作用為零。
func (g *Game) evaluateValue(value valueExpression, controller *model.Player) int {
	switch value.Kind {
	case valueConstant:
		return value.Constant
	case valueAdd:
		return g.evaluateValue(*value.Left, controller) + g.evaluateValue(*value.Right, controller)
	case valueDistinctPrintedCosts:
		costs := make(map[int]struct{})
		for _, id := range g.selectObjects(value.Selector, controller) {
			costs[g.printedReserveCost(g.state.Objects[id].Card)] = struct{}{}
		}
		return len(costs)
	default:
		panic("uncompiled value expression")
	}
}
