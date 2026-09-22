package game

import (
	"go-tcg/internal/model"
	"path/filepath"
	"testing"
)

// TestNewStandardGameCreatesCompleteFixedDeckGame 驗證正式入口以固定牌組建立雙方的完整開局狀態。
// 輸入為正常 StandardGameConfig；輸出為雙方各有七張起手牌的 Game，副作用為讀取 repository 內的卡面資料。
func TestNewStandardGameCreatesCompleteFixedDeckGame(t *testing.T) {
	configuration := StandardGameConfig{
		Players: [2]*model.Player{
			model.PlayerOne,
			model.PlayerTwo,
		},
		RepositoryRoot: filepath.Join(
			"..",
			"..",
		),
	}
	game, err := NewStandardGame(configuration)
	if err != nil {
		t.Fatalf("NewStandardGame() error = %v", err)
	}
	for _, player := range configuration.Players {
		zones, exists := game.state.Zones[player.UID]
		if !exists {
			t.Fatalf("NewStandardGame() has no zones for player %q", player.UID)
		}
		if len(zones.Hand) != 7 {
			t.Fatalf("opening hand for player %q = %d, want 7", player.UID, len(zones.Hand))
		}
	}
}
