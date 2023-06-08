package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("No version number provided. Please provide a version number as an argument.")
	}

	version := os.Args[1]
	base_url := "https://github.com/shimman-dev/piscator/releases/download/v" + version
	architectures := []string{"darwin-amd64", "darwin-arm64", "linux-amd64", "linux-arm64"}

	formula, err := ioutil.ReadFile("./piscator.rb")
	if err != nil {
		log.Fatalf("Failed to read formula file: %v", err)
	}

	for _, arch := range architectures {
		url := base_url + "/piscator-v" + version + "-" + arch + ".tar.gz.sha256"
		resp, err := http.Get(url)
		if err != nil {
			log.Fatalf("Failed to download %s: %v", url, err)
		}
		defer resp.Body.Close()

		sha256Bytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("Failed to read response body: %v", err)
		}

		newSha256 := strings.TrimSpace(string(sha256Bytes))

		sha256Pattern := `sha256 ".*?" # ` + arch
		re := regexp.MustCompile(sha256Pattern)
		if !re.MatchString(string(formula)) {
			log.Fatalf("Failed to match pattern %s in formula", sha256Pattern)
		}
		newFormula := re.ReplaceAllString(string(formula), `sha256 "`+newSha256+`" # `+arch)

		err = ioutil.WriteFile("./piscator.rb", []byte(newFormula), 0644)
		if err != nil {
			log.Fatalf("Failed to write new formula file: %v", err)
		}

		formula = []byte(newFormula)
	}

	fmt.Println("Homebrew formula updated successfully.")
}
