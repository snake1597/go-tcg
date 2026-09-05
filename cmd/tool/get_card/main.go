package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go-tcg/internal/game"

	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"unicode"
)

// go run ./cmd/tool/get_card/main.go
func main() {
	slugList := []string{
		"spirit-of-fire",
		"tonoris-lone-mercenary",
		"bulwark-sword",
		"grand-crusaders-ring",
		"safeguard-amulet",
		"smoke-bombs",
		"viridian-protective-trinket",
		"water-resonance-bauble",
		"wind-resonance-bauble",
		"impact-hammer",
		"infernal-vessel",
		"the-duchesss-thornes",
		"five-of-spades",
		"four-of-spades",
		"noire-ace-of-spades",
		"three-of-spades",
		"trump-set",
		"two-of-spades",
		"wonderlands-reign",
		"arthur-young-heir",
		"blazing-throw",
		"duchess-six-of-hearts",
		"fiery-interference",
		"four-of-hearts",
		"heated-vengeance",
		"peppered-chef",
		"red-hare-unrivaled-stallion",
		"rouge-ace-of-hearts",
		"straight-flare",
		"three-of-hearts",
		"two-of-hearts",
		"verita-queen-of-hearts",
	}

	for _, slug := range slugList {
		req := fmt.Sprintf("https://api.gatcg.com/cards/%s", slug)
		resp, err := http.Get(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
			panic("card API returned " + resp.Status)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		card := &game.Card{}
		err = json.Unmarshal(body, card)
		if err != nil {
			log.Printf("json.Unmarshal error: %v, slug: %s", err, slug)
			continue
		}

		fileName := fmt.Sprintf("./card/%s.json", card.Slug)
		if err := writeJSONBody(fileName, body); err != nil {
			panic(err)
		}

		// file, err := os.ReadFile(fileName)
		// if err != nil {
		// 	panic(err)
		// }
		// card = &entity.Card{}
		// err = json.Unmarshal(file, card)
		// if err != nil {
		// 	panic(err)
		// }
		// log.Printf("Card: %+v", card)
	}
}

func writeJSONBody(filename string, body []byte) error {
	var formatted bytes.Buffer
	if err := json.Indent(&formatted, body, "", "  "); err != nil {
		return err
	}
	return os.WriteFile(filename, formatted.Bytes(), 0o644)
}

func cardNameToSlug(name string) string {
	var slug strings.Builder
	previousHyphen := false

	for _, char := range strings.ToLower(name) {
		if unicode.IsLetter(char) || unicode.IsDigit(char) {
			slug.WriteRune(char)
			previousHyphen = false
			continue
		}

		if slug.Len() > 0 && !previousHyphen {
			slug.WriteByte('-')
			previousHyphen = true
		}
	}

	return strings.TrimSuffix(slug.String(), "-")
}
