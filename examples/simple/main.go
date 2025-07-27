package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ethanbaker/grabber"
)

func main() {
	// Create a grabber with default options
	g := grabber.New()

	// Check if URL was provided as argument
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run main.go <url>")
	}

	url := os.Args[1]

	// Identify the platform
	platform := g.IdentifyPlatform(url)
	if platform == grabber.Unknown {
		log.Fatal("Unsupported platform for URL:", url)
	}

	fmt.Printf("Detected platform: %s\n", platform)
	fmt.Printf("Downloading from: %s\n", url)

	// Download the media
	name := fmt.Sprintf("download-%s", time.Now().Format("20060102_150405"))
	result, err := g.Download(url, name)
	if err != nil {
		log.Fatalf("Download failed: %v", err)
	}

	// Print the results
	for _, file := range result {
		fmt.Printf("Successfully downloaded file %s!\n", file.Name)
		fmt.Printf("\tPath: %s\n", file.Path)
		fmt.Printf("\tType: %s\n", file.MediaType)
	}
}
