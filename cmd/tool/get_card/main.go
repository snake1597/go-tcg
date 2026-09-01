package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go-tcg/internal/entity"
	"io"
	"log"
	"net/http"
	"os"
)

// go run ./cmd/tool/get_card/main.go
func main() {
	slugList := []string{"spirit-of-fire",
		"alice-distorted-queen",
		"backup-charger",
		"grand-crusaders-ring",
		"nullifying-mirror",
		"safeguard-amulet",
		"smoke-bombs",
		"tariff-ring",
		"viridian-protective-trinket",
		"impact-hammer",
		"infernal-vessel",
		"mantle-of-the-abyss",
		"blighted-jewel",
		"lingering-banshee",
		"bill-chimney-sweep",
		"cinder-geyser",
		"creative-shock",
		"duchess-six-of-hearts",
		"emberwrath-witch",
		"fanatical-devotee",
		"fractal-of-sparks",
		"furnace-drone",
		"incinerated-templar",
		"incinerator-felindroid",
		"nether-dodobird",
		"searing-rebuke",
		"vengeful-paramour",
		"volda-smolders-spite",
		"everflame-staff",
		"fanned-synchron",
		"incinerator-felindroid",
		"reduce-to-ash",
		"restorative-flame",
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

		card := &entity.Card{}
		err = json.Unmarshal(body, card)
		if err != nil {
			log.Printf("json.Unmarshal error: %v, slug: %s", err, slug)
			continue
		}

		fileName := fmt.Sprintf("./card/%s.json", card.Slug)
		if err := appendJSONBody(fileName, body); err != nil {
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

func appendJSONBody(filename string, body []byte) error {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	var formatted bytes.Buffer
	if err := json.Indent(&formatted, body, "", "  "); err != nil {
		return err
	}

	if _, err := file.Write(formatted.Bytes()); err != nil {
		return err
	}

	return nil
}
