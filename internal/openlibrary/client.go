package openlibrary

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Client struct {
	HTTP *http.Client
}

func New() *Client {
	return &Client{HTTP: &http.Client{Timeout: 15 * time.Second}}
}

type Metadata struct {
	Title   string
	Authors []string
	Year    *int64
}

type isbnWork struct {
	Title   string   `json:"title"`
	Publish []string `json:"publish_date"`
	Authors []struct {
		Key string `json:"key"`
	} `json:"authors"`
}

type searchDoc struct {
	Title     string   `json:"title"`
	Author    []string `json:"author_name"`
	FirstYear *int     `json:"first_publish_year"`
}

func (c *Client) LookupISBN(isbn string) (*Metadata, error) {
	isbn = strings.ReplaceAll(strings.TrimSpace(isbn), "-", "")
	if isbn == "" {
		return nil, fmt.Errorf("empty isbn")
	}
	req, err := http.NewRequest(http.MethodGet, "https://openlibrary.org/isbn/"+isbn+".json", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "book-pile/1.0")
	res, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return c.searchISBN(isbn)
	}
	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(res.Body, 1024))
		return nil, fmt.Errorf("open library isbn: %s %s", res.Status, body)
	}
	var work isbnWork
	if err := json.NewDecoder(res.Body).Decode(&work); err != nil {
		return nil, err
	}
	meta := &Metadata{Title: work.Title}
	if len(work.Publish) > 0 && len(work.Publish[0]) >= 4 {
		if y, err := parseYear(work.Publish[0][:4]); err == nil {
			meta.Year = &y
		}
	}
	for _, a := range work.Authors {
		if name, err := c.authorName(a.Key); err == nil && name != "" {
			meta.Authors = append(meta.Authors, name)
		}
	}
	if meta.Title == "" {
		return c.searchISBN(isbn)
	}
	return meta, nil
}

func (c *Client) searchISBN(isbn string) (*Metadata, error) {
	req, err := http.NewRequest(http.MethodGet, "https://openlibrary.org/search.json?isbn="+isbn+"&limit=1", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "book-pile/1.0")
	res, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("open library search: %s", res.Status)
	}
	var payload struct {
		Docs []searchDoc `json:"docs"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return nil, err
	}
	if len(payload.Docs) == 0 {
		return nil, fmt.Errorf("no open library match for isbn %s", isbn)
	}
	d := payload.Docs[0]
	meta := &Metadata{Title: d.Title, Authors: d.Author}
	if d.FirstYear != nil {
		y := int64(*d.FirstYear)
		meta.Year = &y
	}
	return meta, nil
}

func (c *Client) authorName(key string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, "https://openlibrary.org"+key+".json", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "book-pile/1.0")
	res, err := c.HTTP.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("author %s", res.Status)
	}
	var payload struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return "", err
	}
	return payload.Name, nil
}

func (c *Client) DownloadCover(isbn, destPath string) (bool, error) {
	isbn = strings.ReplaceAll(strings.TrimSpace(isbn), "-", "")
	if isbn == "" {
		return false, nil
	}
	url := "https://covers.openlibrary.org/b/isbn/" + isbn + "-L.jpg"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("User-Agent", "book-pile/1.0")
	res, err := c.HTTP.Do(req)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return false, nil
	}
	data, err := io.ReadAll(io.LimitReader(res.Body, 8<<20))
	if err != nil {
		return false, err
	}
	if len(data) < 2048 {
		return false, nil
	}
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return false, err
	}
	return true, os.WriteFile(destPath, data, 0o644)
}

func parseYear(s string) (int64, error) {
	var y int64
	_, err := fmt.Sscan(s, &y)
	return y, err
}
