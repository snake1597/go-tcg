// Package carddata versions and verifies the repository's immutable card-data snapshot.
package carddata

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"

	"go-tcg/internal/constants"
)

type Manifest struct {
	SchemaVersion int         `json:"schema_version"`
	DataVersion   string      `json:"data_version"`
	Source        string      `json:"source"`
	DatasetSHA256 string      `json:"dataset_sha256"`
	Cards         []CardEntry `json:"cards"`
}

type CardEntry struct {
	Path       string `json:"path"`
	CardID     string `json:"card_id"`
	Slug       string `json:"slug"`
	Name       string `json:"name"`
	LastUpdate string `json:"last_update"`
	FileSHA256 string `json:"file_sha256"`
}

type cardMetadata struct {
	UUID       string `json:"uuid"`
	Slug       string `json:"slug"`
	Name       string `json:"name"`
	LastUpdate string `json:"last_update"`
}

func BuildManifest(cardDir, dataVersion string) (Manifest, error) {
	if strings.TrimSpace(dataVersion) == "" {
		return Manifest{}, errors.New("card data version is required")
	}

	directoryEntries, err := os.ReadDir(cardDir)
	if err != nil {
		return Manifest{}, fmt.Errorf("read card directory: %w", err)
	}

	var names []string
	for _, entry := range directoryEntries {
		if !entry.IsDir() && strings.EqualFold(filepath.Ext(entry.Name()), ".json") {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	if len(names) == 0 {
		return Manifest{}, errors.New("card directory contains no JSON files")
	}

	manifest := Manifest{
		SchemaVersion: constants.CardDataSchemaVersion,
		DataVersion:   dataVersion,
		Source:        constants.CardDataSourcePattern,
		Cards:         make([]CardEntry, 0, len(names)),
	}
	cardIDs := make(map[string]string, len(names))
	slugs := make(map[string]string, len(names))
	datasetHash := sha256.New()

	for _, name := range names {
		path := filepath.Join(cardDir, name)
		contents, err := os.ReadFile(path)
		if err != nil {
			return Manifest{}, fmt.Errorf("read %s: %w", name, err)
		}

		metadata, err := decodeCardMetadata(contents)
		if err != nil {
			return Manifest{}, fmt.Errorf("%s: %w", name, err)
		}
		if previous, exists := cardIDs[metadata.UUID]; exists {
			return Manifest{}, fmt.Errorf("duplicate card ID %q in %s and %s", metadata.UUID, previous, name)
		}
		if previous, exists := slugs[metadata.Slug]; exists {
			return Manifest{}, fmt.Errorf("duplicate card slug %q in %s and %s", metadata.Slug, previous, name)
		}
		if want := metadata.Slug + ".json"; name != want {
			return Manifest{}, fmt.Errorf("file name %q does not match card slug %q", name, metadata.Slug)
		}
		cardIDs[metadata.UUID] = name
		slugs[metadata.Slug] = name

		fileSum := sha256.Sum256(contents)
		fileDigest := hex.EncodeToString(fileSum[:])
		_, _ = io.WriteString(datasetHash, name)
		_, _ = datasetHash.Write([]byte{0})
		_, _ = datasetHash.Write(fileSum[:])
		_, _ = datasetHash.Write([]byte{'\n'})

		manifest.Cards = append(manifest.Cards, CardEntry{
			Path:       name,
			CardID:     metadata.UUID,
			Slug:       metadata.Slug,
			Name:       metadata.Name,
			LastUpdate: metadata.LastUpdate,
			FileSHA256: fileDigest,
		})
	}

	manifest.DatasetSHA256 = hex.EncodeToString(datasetHash.Sum(nil))
	return manifest, nil
}

func VerifyManifest(cardDir string, expected Manifest) error {
	if expected.SchemaVersion != constants.CardDataSchemaVersion {
		return fmt.Errorf("unsupported manifest schema version %d", expected.SchemaVersion)
	}
	if expected.Source != constants.CardDataSourcePattern {
		return fmt.Errorf("unsupported card data source %q", expected.Source)
	}

	actual, err := BuildManifest(cardDir, expected.DataVersion)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(actual, expected) {
		return fmt.Errorf("card data does not match manifest %q: expected dataset sha256 %s, got %s", expected.DataVersion, expected.DatasetSHA256, actual.DatasetSHA256)
	}
	return nil
}

func ReadManifest(path string) (Manifest, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return Manifest{}, fmt.Errorf("read manifest: %w", err)
	}

	var manifest Manifest
	decoder := json.NewDecoder(bytes.NewReader(contents))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&manifest); err != nil {
		return Manifest{}, fmt.Errorf("decode manifest: %w", err)
	}
	if err := requireEOF(decoder); err != nil {
		return Manifest{}, fmt.Errorf("manifest must contain exactly one JSON object: %w", err)
	}
	return manifest, nil
}

func WriteManifest(path string, manifest Manifest) error {
	contents, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("encode manifest: %w", err)
	}
	contents = append(contents, '\n')
	if err := os.WriteFile(path, contents, 0o644); err != nil {
		return fmt.Errorf("write manifest: %w", err)
	}
	return nil
}

func decodeCardMetadata(contents []byte) (cardMetadata, error) {
	var metadata cardMetadata
	decoder := json.NewDecoder(bytes.NewReader(contents))
	if err := decoder.Decode(&metadata); err != nil {
		return cardMetadata{}, fmt.Errorf("decode card JSON: %w", err)
	}
	if err := requireEOF(decoder); err != nil {
		return cardMetadata{}, fmt.Errorf("card file must contain exactly one JSON object: %w", err)
	}
	if metadata.UUID == "" || metadata.Slug == "" || metadata.Name == "" || metadata.LastUpdate == "" {
		return cardMetadata{}, errors.New("uuid, slug, name, and last_update are required")
	}
	if _, err := time.Parse(time.RFC3339Nano, metadata.LastUpdate); err != nil {
		return cardMetadata{}, fmt.Errorf("invalid last_update %q: %w", metadata.LastUpdate, err)
	}
	return metadata, nil
}

func requireEOF(decoder *json.Decoder) error {
	var extra json.RawMessage
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("found another JSON value")
		}
		return err
	}
	return nil
}
