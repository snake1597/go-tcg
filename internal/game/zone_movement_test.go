package game

import "testing"

// TestBanishCardsFromGraveyardDoesNotDependOnSelection 驗證墓地放逐只依輸入卡牌付款，且錯誤不改動原區域。
func TestBanishCardsFromGraveyardDoesNotDependOnSelection(t *testing.T) {
	zones := playerZones{
		Graveyard: []cardInstanceID{
			"first",
			"second",
		},
	}
	if _, err := banishCardsFromGraveyard(
		zones,
		[]cardInstanceID{
			"first",
			"first",
		},
	); err == nil {
		t.Fatal("duplicate banishment was accepted")
	}
	updated, err := banishCardsFromGraveyard(
		zones,
		[]cardInstanceID{
			"first",
		},
	)
	if err != nil {
		t.Fatalf("banishCardsFromGraveyard() error = %v", err)
	}
	if len(zones.Graveyard) != 2 || len(zones.Banishment) != 0 {
		t.Fatalf("original zones changed: %#v", zones)
	}
	if len(updated.Graveyard) != 1 || updated.Graveyard[0] != "second" || len(updated.Banishment) != 1 || updated.Banishment[0] != "first" {
		t.Fatalf("updated zones = %#v", updated)
	}
}
