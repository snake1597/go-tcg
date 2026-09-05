package carddata

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildManifestIsDeterministic(t *testing.T) {
	dir := t.TempDir()
	writeCard(t, dir, "zeta.json", `{"uuid":"card-z","slug":"zeta","name":"Zeta","last_update":"2026-01-02T03:04:05Z"}`)
	writeCard(t, dir, "alpha.json", `{"uuid":"card-a","slug":"alpha","name":"Alpha","last_update":"2025-01-02T03:04:05Z"}`)

	first, err := BuildManifest(dir, "card-data-v1")
	if err != nil {
		t.Fatalf("BuildManifest() error = %v", err)
	}
	second, err := BuildManifest(dir, "card-data-v1")
	if err != nil {
		t.Fatalf("BuildManifest() second error = %v", err)
	}

	if first.DatasetSHA256 != second.DatasetSHA256 {
		t.Fatalf("dataset digest changed: %q != %q", first.DatasetSHA256, second.DatasetSHA256)
	}
	if len(first.Cards) != 2 {
		t.Fatalf("card count = %d, want 2", len(first.Cards))
	}
	if first.Cards[0].Path != "alpha.json" || first.Cards[1].Path != "zeta.json" {
		t.Fatalf("cards are not sorted by path: %+v", first.Cards)
	}
	if first.Source != "./card/*.json" {
		t.Fatalf("source = %q, want ./card/*.json", first.Source)
	}
}

func TestBuildManifestRejectsMultipleJSONDocuments(t *testing.T) {
	dir := t.TempDir()
	writeCard(t, dir, "broken.json", `{"uuid":"one","slug":"broken","name":"Broken","last_update":"2026-01-02T03:04:05Z"}{"uuid":"two"}`)

	_, err := BuildManifest(dir, "card-data-v1")
	if err == nil || !strings.Contains(err.Error(), "exactly one JSON object") {
		t.Fatalf("BuildManifest() error = %v, want multiple-document error", err)
	}
}

func TestBuildManifestRejectsDuplicateCardID(t *testing.T) {
	dir := t.TempDir()
	writeCard(t, dir, "first.json", `{"uuid":"duplicate","slug":"first","name":"First","last_update":"2026-01-02T03:04:05Z"}`)
	writeCard(t, dir, "second.json", `{"uuid":"duplicate","slug":"second","name":"Second","last_update":"2026-01-02T03:04:05Z"}`)

	_, err := BuildManifest(dir, "card-data-v1")
	if err == nil || !strings.Contains(err.Error(), "duplicate card ID") {
		t.Fatalf("BuildManifest() error = %v, want duplicate card ID error", err)
	}
}

func TestVerifyManifestDetectsChangedCardData(t *testing.T) {
	dir := t.TempDir()
	writeCard(t, dir, "alpha.json", `{"uuid":"card-a","slug":"alpha","name":"Alpha","last_update":"2025-01-02T03:04:05Z"}`)

	manifest, err := BuildManifest(dir, "card-data-v1")
	if err != nil {
		t.Fatalf("BuildManifest() error = %v", err)
	}
	writeCard(t, dir, "alpha.json", `{"uuid":"card-a","slug":"alpha","name":"Changed","last_update":"2025-01-02T03:04:05Z"}`)

	err = VerifyManifest(dir, manifest)
	if err == nil || !strings.Contains(err.Error(), "does not match") {
		t.Fatalf("VerifyManifest() error = %v, want mismatch error", err)
	}
}

func writeCard(t *testing.T, dir, name, contents string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(contents), 0o600); err != nil {
		t.Fatalf("write card: %v", err)
	}
}
