package sitemap

import (
	"compress/gzip"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type SitemapIndex struct {
	XMLName  xml.Name       `xml:"sitemapindex"`
	Sitemaps []SitemapEntry `xml:"sitemap"`
}

type SitemapEntry struct {
	Loc     string `xml:"loc"`
	LastMod string `xml:"lastmod"`
}

type URLSet struct {
	XMLName xml.Name   `xml:"urlset"`
	URLs    []URLEntry `xml:"url"`
}

type URLEntry struct {
	Loc     string `xml:"loc"`
	LastMod string `xml:"lastmod"`
}

func fetchXML(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			// do nothing
			return
		}
	}()

	if strings.HasSuffix(url, ".gz") {
		gzr, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, err
		}
		defer func() {
			if err := gzr.Close(); err != nil {
				// do nothing
				return
			}
		}()
		return io.ReadAll(gzr)
	}
	return io.ReadAll(resp.Body)
}

// parseSitemap fetches and parses a sitemap or sitemap index from the given URL.
// It appends the URLs and their last modified dates to the provided records slice.
// If the URL points to a sitemap index, it recursively fetches and parses each sitemap.
// The records slice should be initialized with a header row before calling this function.
// If processedMap is not nil, record the number of loc entries for each urlset.
func parseSitemap(url string, records *[][]string, processedMap map[string]int) error {
	data, err := fetchXML(url)
	if err != nil {
		return err
	}

	if strings.Contains(string(data), "<sitemapindex") {
		var idx SitemapIndex
		if err := xml.Unmarshal(data, &idx); err != nil {
			return err
		}
		if len(idx.Sitemaps) == 0 {
			return fmt.Errorf("invalid sitemapindex xml: no <sitemap> entries found")
		}
		for _, sm := range idx.Sitemaps {
			loc := strings.Trim(sm.Loc, " \n")
			if err := parseSitemap(loc, records, processedMap); err != nil {
				return err
			}
		}
	} else if strings.Contains(string(data), "<urlset") {
		var us URLSet
		if err := xml.Unmarshal(data, &us); err != nil {
			return err
		}
		if len(us.URLs) == 0 {
			return fmt.Errorf("invalid urlset xml: no <url> entries found")
		}
		for _, u := range us.URLs {
			loc := strings.Trim(u.Loc, " \n")
			*records = append(*records, []string{loc, u.LastMod})
		}
		if processedMap != nil {
			processedMap[url] = len(us.URLs)
		}
	} else {
		return fmt.Errorf("invalid sitemap xml: missing <sitemapindex> or <urlset>")
	}
	return nil
}

// Result holds the result of GetSitemapRecords, including records, processed URLs, and loc counts.
type Result struct {
	Records       [][]string     // a slice of records, each containing the URL and its last modified date.
	ProcessedURLs map[string]int // a map of processed sitemap URLs and the number of locs found in each.
}

// GetSitemapRecords fetches and parses a sitemap or sitemap index from the given URL.
// It returns a Result containing:
// - Records: a slice of records, each containing the URL and its last modified date.
// - ProcessedURLs: a map of processed sitemap URLs and the number of locs found in each.
// The first record is a header row with "loc" and "lastmod".
// If the URL points to a sitemap index, it recursively fetches and parses each sitemap.
// If an error occurs during fetching or parsing, it returns the error.
func GetSitemapRecords(url string) (*Result, error) {
	records := [][]string{{"loc", "lastmod"}}
	processed := make(map[string]int)
	if err := parseSitemap(url, &records, processed); err != nil {
		return nil, err
	}
	return &Result{
		Records:       records,
		ProcessedURLs: processed,
	}, nil
}
