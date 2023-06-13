package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

type Architecture struct {
	URL  string   `json:"url"`
	Bin  []string `json:"bin"`
	Hash string   `json:"hash"`
}

type ScoopPkg struct {
	Version      string                  `json:"version"`
	Architecture map[string]Architecture `json:"architecture"`
	Homepage     string                  `json:"homepage"`
	License      string                  `json:"license"`
	Description  string                  `json:"description"`
}

func main() {
	if len(os.Args) < 3 {
		log.Fatal("Usage: go run script.go <version> <path to JSON file>")
	}

	version := os.Args[1]
	jsonPath := os.Args[2]

	pkg := ScoopPkg{}

	data, err := ioutil.ReadFile(jsonPath)
	if err != nil {
		log.Fatalf("Failed to read JSON file: %v", err)
	}

	err = json.Unmarshal(data, &pkg)
	if err != nil {
		log.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	pkg.Version = version

	// Fetch new SHA256 hashes
	for arch, details := range pkg.Architecture {
		url := strings.ReplaceAll(details.URL, "v"+pkg.Version, "v"+version)
		url += ".sha256"

		resp, err := http.Get(url)
		if err != nil {
			log.Fatalf("Failed to download %s: %v", url, err)
		}
		defer resp.Body.Close()

		sha256Bytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("Failed to read response body: %v", err)
		}

		details.Hash = "sha256:" + strings.TrimSpace(string(sha256Bytes))
		details.URL = strings.ReplaceAll(details.URL, "v"+pkg.Version, "v"+version)

		pkg.Architecture[arch] = details
	}

	updatedData, err := json.MarshalIndent(pkg, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal updated package: %v", err)
	}

	err = ioutil.WriteFile(jsonPath, updatedData, 0644)
	if err != nil {
		log.Fatalf("Failed to write updated data to JSON file: %v", err)
	}

	fmt.Println("Scoop package updated successfully.")
}
