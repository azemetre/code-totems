package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	return nil
}

func updateSHA256(version, arch string, formula []byte) ([]byte, error) {
	baseURL := "https://github.com/shimman-dev/piscator/releases/download/v" + version
	url := baseURL + "/piscator-v" + version + "-" + arch + ".tar.gz.sha256"

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download %s: %v", url, err)
	}
	defer resp.Body.Close()

	sha256Bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	newSha256 := strings.TrimSpace(string(sha256Bytes))

	fmt.Printf("Got new SHA256 for %s: %s\n", arch, newSha256)

	sha256Pattern := `sha256 ".*" # ` + arch
	re := regexp.MustCompile(sha256Pattern)
	if !re.MatchString(string(formula)) {
		return nil, fmt.Errorf("failed to match pattern %s in formula", sha256Pattern)
	}
	newFormula := re.ReplaceAll(formula, []byte(`sha256 "`+newSha256+`" # `+arch))

	return newFormula, nil
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("No version number provided. Please provide a version number as an argument.")
	}

	version := os.Args[1]

	// Make a backup before modifying the formula file
	err := copyFile("./piscator.rb", "./piscator.rb.bak")
	if err != nil {
		log.Fatalf("failed to create backup of formula file: %v", err)
	}

	formula, err := ioutil.ReadFile("./piscator.rb")
	if err != nil {
		log.Fatalf("failed to read formula file: %v", err)
	}

	versionPattern := `version ".*?"`
	versionRe := regexp.MustCompile(versionPattern)
	if !versionRe.MatchString(string(formula)) {
		log.Fatalf("failed to match pattern %s in formula", versionPattern)
	}
	formula = versionRe.ReplaceAll(formula, []byte(`version "`+version+`"`))

	architectures := []string{"darwin-amd64", "darwin-arm64", "linux-amd64", "linux-arm64"}

	for _, arch := range architectures {
		formula, err = updateSHA256(version, arch, formula)
		if err != nil {
			log.Fatalf("failed to update formula for architecture %s: %v", arch, err)
		}
	}

	err = ioutil.WriteFile("./piscator.rb", formula, 0644)
	if err != nil {
		log.Fatalf("failed to write new formula file: %v", err)
	}

	fmt.Println("Homebrew formula updated successfully.")
}
