package game

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"go-tcg/internal/constants"
	"go-tcg/internal/model"
)

// ViewHandle 是單一玩家視角中的短期不透明識別。
type ViewHandle string

// grantCardTracking 為玩家首次追蹤到的卡牌配置 handle，後續追蹤保留同一 handle。
// 是否應讓玩家追蹤此牌由呼叫端決定，此函式不檢查牌所在區域或可見性。
func (g *Game) grantCardTracking(player *model.Player, card entityID) {
	if _, exists := g.state.Knowledge.Players[player.UID].Cards[card]; exists {
		return
	}
	handle := g.newViewHandle(
		player,
		constants.ViewHandleSubjectCardPrefix+string(card),
	)
	g.state.Knowledge.Players[player.UID].Cards[card] = handle
}

func (g *Game) revokeCardTracking(player *model.Player, card entityID) {
	handle, exists := g.state.Knowledge.Players[player.UID].Cards[card]
	if !exists {
		return
	}
	delete(g.state.Knowledge.Players[player.UID].Cards, card)
	choice := g.state.Knowledge.Choice
	if choice == nil || !samePlayer(choice.Actor, player) {
		return
	}
	delete(choice.Options, handle)
	if len(choice.Options) == 0 {
		g.state.Knowledge.Choice = nil
	}
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
