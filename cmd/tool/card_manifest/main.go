package main

import (
	"flag"
	"fmt"
	carddata "go-tcg/internal/card_data"
	"log"
)

func main() {
	cardDir := flag.String("card-dir", "card", "directory containing the authoritative card JSON snapshot")
	manifestPath := flag.String("manifest", "card-data-manifest.json", "card data manifest path")
	dataVersion := flag.String("version", "card-data-v1", "data version to write")
	write := flag.Bool("write", false, "write the manifest instead of verifying it")
	flag.Parse()

	if *write {
		manifest, err := carddata.BuildManifest(*cardDir, *dataVersion)
		if err != nil {
			log.Fatal(err)
		}
		if err := carddata.WriteManifest(*manifestPath, manifest); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("wrote %s (%d cards, sha256 %s)\n", *manifestPath, len(manifest.Cards), manifest.DatasetSHA256)
		return
	}

	manifest, err := carddata.ReadManifest(*manifestPath)
	if err != nil {
		log.Fatal(err)
	}
	if err := carddata.VerifyManifest(*cardDir, manifest); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("verified %s (%d cards, sha256 %s)\n", manifest.DataVersion, len(manifest.Cards), manifest.DatasetSHA256)
}
