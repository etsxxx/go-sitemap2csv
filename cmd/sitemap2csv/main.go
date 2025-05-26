package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"

	"github.com/etsxxx/go-sitemap2csv/pkg/sitemap"
)

var version, gitcommit string

func main() {
	var (
		showVersion bool
		showHelp    bool
		noHeader    bool
	)
	flag.BoolVar(&showVersion, "v", false, "show version")
	flag.BoolVar(&showVersion, "version", false, "show version")
	flag.BoolVar(&showHelp, "h", false, "show help")
	flag.BoolVar(&showHelp, "help", false, "show help")
	flag.BoolVar(&noHeader, "no-header", false, "output CSV without header row")
	flag.Usage = func() {
		fmt.Println("Usage: sitemap2csv [options] <sitemap_url> <output_csv_file>")
		flag.PrintDefaults()
	}
	flag.Parse()

	if showVersion {
		fmt.Printf("sitemap2csv version: %s (rev:%s)\n", version, gitcommit)
		os.Exit(0)
	}
	if showHelp {
		flag.Usage()
		os.Exit(0)
	}

	if flag.NArg() < 2 {
		fmt.Printf("sitemap2csv needs at least 2 arguments\n")
		flag.Usage()
		os.Exit(1)
	}
	url := flag.Arg(0)
	outputFile := flag.Arg(1)

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
	records := result.Records
	if noHeader && len(records) >= 1 {
		// If noHeader is true, remove the header row from the records
		// This assumes the first row is the header
		records = records[1:]
	}
	if err := w.WriteAll(records); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing CSV: %v\n", err)
		os.Exit(1)
	}
	if err := w.Error(); err != nil {
		fmt.Fprintf(os.Stderr, "CSV Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("CSV file created: %s\n", outputFile)
}
