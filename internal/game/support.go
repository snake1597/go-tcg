package game

import (
	"fmt"
	"go-tcg/internal/model"
	"path/filepath"
)

type StandardGameConfig struct {
	Players        []*model.Player
	RepositoryRoot string
	Seed           uint64
}

// NewStandardGame 載入已驗證的固定卡面資料與固定 Standard 牌組，建立可直接執行的單局。
// 輸入為兩名不同玩家、repository root 與隨機種子；輸出為初始化後的 Game 或資料／設定錯誤，副作用為讀取卡面檔案。
func NewStandardGame(configuration StandardGameConfig) (*Game, error) {
	if len(configuration.Players) != 2 ||
		configuration.Players[0] == nil ||
		configuration.Players[1] == nil ||
		configuration.Players[0].UID == "" ||
		configuration.Players[1].UID == "" ||
		samePlayer(configuration.Players[0], configuration.Players[1]) {
		return nil, fmt.Errorf("standard game requires two distinct players")
	}

	definitions, err := loadCardDefinitions(
		filepath.Join(configuration.RepositoryRoot, "card"),
		filepath.Join(configuration.RepositoryRoot, "card-data-manifest.json"),
	)
	if err != nil {
		return nil, fmt.Errorf("load fixed card data: %w", err)
	}

	deck := fixedStandardDeck()
	if err := validateDeckReferences(deck, definitions); err != nil {
		return nil, fmt.Errorf("validate fixed deck references: %w", err)
	}

	return newStandardSetup(configuration, definitions, deck, deck)
}
