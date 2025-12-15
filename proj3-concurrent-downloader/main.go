package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	url := flag.String("url", "", "URL of the file to download")
	output := flag.String("output", "", "Output file path (default: filename from URL)")
	workers := flag.Int("workers", 4, "Number of concurrent workers")

	flag.Parse()

	if *url == "" {
		fmt.Println("Error: URL is required")
		flag.Usage()
		os.Exit(1)
	}

	// Determine output filename
	outputPath := *output
	if outputPath == "" {
		outputPath = filepath.Base(*url)
		if outputPath == "/" || outputPath == "." {
			outputPath = "downloaded_file"
		}
	}

	// Create downloader
	downloader, err := NewDownloader(*url, outputPath, *workers)
	if err != nil {
		fmt.Printf("Error creating downloader: %v\n", err)
		os.Exit(1)
	}

	// Start download
	if err := downloader.Download(); err != nil {
		fmt.Printf("\nError: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("File saved to: %s\n", outputPath)
}
