package sitemap

import (
	"bytes"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestFetchXML_Plain(t *testing.T) {
	h := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<urlset><url><loc>http://127.0.0.1/</loc><lastmod>2020-01-01</lastmod></url></urlset>`))
	}
	ts := httptest.NewServer(http.HandlerFunc(h))
	defer ts.Close()
	data, err := fetchXML(ts.URL)
	if err != nil {
		t.Fatalf("fetchXML failed: %v", err)
	}
	if !strings.Contains(string(data), "<urlset>") {
		t.Errorf("unexpected data: %s", data)
	}
}

func TestFetchXML_Gzip(t *testing.T) {
	h := func(w http.ResponseWriter, r *http.Request) {
		var buf bytes.Buffer
		gz := gzip.NewWriter(&buf)
		if _, err := gz.Write([]byte(`<urlset><url><loc>http://127.0.0.1/</loc><lastmod>2021-01-01</lastmod></url></urlset>`)); err != nil {
			t.Fatalf("failed to write gzipped data: %v", err)
		}
		if err := gz.Close(); err != nil {
			t.Fatalf("failed to close gzip writer: %v", err)
		}
		if _, err := w.Write(buf.Bytes()); err != nil {
			t.Fatalf("failed to write gzipped data: %v", err)
		}
	}
	ts := httptest.NewServer(http.HandlerFunc(h))
	defer ts.Close()
	data, err := fetchXML(ts.URL + "/sitemap.xml.gz")
	if err != nil {
		t.Fatalf("fetchXML (gz) failed: %v", err)
	}
	if !strings.Contains(string(data), "<urlset>") {
		t.Errorf("unexpected gzipped data: %s", data)
	}
}

func TestParseSitemap_Urlset(t *testing.T) {
	h := func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte(`<urlset><url><loc>http://127.0.0.1/</loc><lastmod>2022-01-01</lastmod></url></urlset>`)); err != nil {
			t.Fatalf("failed to write response: %v", err)
		}
	}
	ts := httptest.NewServer(http.HandlerFunc(h))
	defer ts.Close()
	records := [][]string{{"loc", "lastmod"}}
	processed := make(map[string]int)
	err := parseSitemap(ts.URL, &records, processed)
	if err != nil {
		t.Fatalf("parseSitemap failed: %v", err)
	}
	want := [][]string{{"loc", "lastmod"}, {"http://127.0.0.1/", "2022-01-01"}}
	if !reflect.DeepEqual(records, want) {
		t.Errorf("got %v, want %v", records, want)
	}
	if processed[ts.URL] != 1 {
		t.Errorf("expected processed[%s] == 1, got %d", ts.URL, processed[ts.URL])
	}
}

func TestParseSitemap_Index(t *testing.T) {
	// index -> urlset
	urlsetXML := `<urlset><url><loc>http://127.0.0.1/</loc><lastmod>2023-01-01</lastmod></url></urlset>`
	urlsetSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte(urlsetXML)); err != nil {
			t.Fatalf("failed to write response: %v", err)
		}
	}))
	defer urlsetSrv.Close()
	indexXML := `<sitemapindex><sitemap><loc>` + urlsetSrv.URL + `</loc><lastmod>2023-01-01</lastmod></sitemap></sitemapindex>`
	indexSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte(indexXML)); err != nil {
			t.Fatalf("failed to write response: %v", err)
		}
	}))
	defer indexSrv.Close()
	records := [][]string{{"loc", "lastmod"}}
	processed := make(map[string]int)
	err := parseSitemap(indexSrv.URL, &records, processed)
	if err != nil {
		t.Fatalf("parseSitemap index failed: %v", err)
	}
	want := [][]string{{"loc", "lastmod"}, {"http://127.0.0.1/", "2023-01-01"}}
	if !reflect.DeepEqual(records, want) {
		t.Errorf("got %v, want %v", records, want)
	}
	if processed[urlsetSrv.URL] != 1 {
		t.Errorf("expected processed[%s] == 1, got %d", urlsetSrv.URL, processed[urlsetSrv.URL])
	}
}

func TestGetSitemapRecords(t *testing.T) {
	h := func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte(`<urlset><url><loc>http://127.0.0.1/</loc><lastmod>2024-01-01</lastmod></url></urlset>`)); err != nil {
			t.Fatalf("failed to write response: %v", err)
		}
	}
	ts := httptest.NewServer(http.HandlerFunc(h))
	defer ts.Close()
	result, err := GetSitemapRecords(ts.URL)
	if err != nil {
		t.Fatalf("GetSitemapRecords failed: %v", err)
	}
	want := [][]string{{"loc", "lastmod"}, {"http://127.0.0.1/", "2024-01-01"}}
	if !reflect.DeepEqual(result.Records, want) {
		t.Errorf("got %v, want %v", result.Records, want)
	}
	if result.ProcessedURLs[ts.URL] != 1 {
		t.Errorf("expected ProcessedURLs[%s] == 1, got %d", ts.URL, result.ProcessedURLs[ts.URL])
	}
}

func TestParseSitemap_InvalidXML(t *testing.T) {
	h := func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte(`notxml`)); err != nil {
			t.Fatalf("failed to write response: %v", err)
		}
	}
	ts := httptest.NewServer(http.HandlerFunc(h))
	defer ts.Close()
	records := [][]string{{"loc", "lastmod"}}
	processed := make(map[string]int)
	err := parseSitemap(ts.URL, &records, processed)
	if err == nil {
		t.Error("expected error for invalid XML, got nil")
	}
}

func TestFetchXML_HTTPError(t *testing.T) {
	// Use a non-routable address to force error
	_, err := fetchXML("http://127.0.0.1:0/404")
	if err == nil {
		t.Error("expected error for HTTP error, got nil")
	}
}
