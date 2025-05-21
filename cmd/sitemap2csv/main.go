package main

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/etsxxx/go-sitemap2csv/pkg/sitemap"
)

var version, gitcommit string

func main() {
	if len(os.Args) == 2 {
		if os.Args[1] == "-v" || os.Args[1] == "--version" {
			version = fmt.Sprintf("%s (rev:%s)", version, gitcommit)
			fmt.Printf("sitemap2csv version: %s\n", version)
			os.Exit(0)
		} else if os.Args[1] == "-h" || os.Args[1] == "--help" {
			fmt.Println("Usage: sitemap2csv <sitemap_url> <output_csv_file>")
			os.Exit(0)
		}
	}

	if len(os.Args) < 3 {
		fmt.Println("Usage: sitemap2csv <sitemap_url> <output_csv_file>")
		os.Exit(1)
	}
	url := os.Args[1]
	outputFile := os.Args[2]

	result, err := sitemap.GetSitemapRecords(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	for url, count := range result.ProcessedURLs {
		fmt.Printf("Processed %d URLs from %s\n", count, url)
	}

	f, err := os.Create(outputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "File Error: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing file: %v\n", err)
			os.Exit(1)
		}
	}()
	w := csv.NewWriter(f)
	if err := w.WriteAll(result.Records); err != nil {
		// log to stderr
		fmt.Fprintf(os.Stderr, "Error writing CSV: %v\n", err)
		os.Exit(1)
	}
	if err := w.Error(); err != nil {
		fmt.Fprintf(os.Stderr, "CSV Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("CSV file created: %s\n", outputFile)
}
