package news

import (
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var (
	htmlTagRegex = regexp.MustCompile(`<[^>]*>`)
	spaceRegex   = regexp.MustCompile(`\s+`)
)

// Estruturas para parsing de RSS 2.0 XML
type rssRoot struct {
	XMLName xml.Name   `xml:"rss"`
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Title string    `xml:"title"`
	Items []rssItem `xml:"item"`
}

type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	GUID        string `xml:"guid"`
}

// Categories disponíveis no bot
var DefaultCategories = map[string]CategoryInfo{
	"br": {
		Key:         "br",
		Title:       "Brasil (Nacional)",
		Emoji:       "🇧🇷",
		Description: "Manchetes nacionais (CNN Brasil e G1)",
		FeedURLs: []FeedSource{
			{Source: "CNN Brasil", URL: "https://www.cnnbrasil.com.br/feed/"},
			{Source: "G1", URL: "https://g1.globo.com/rss/g1/"},
		},
	},
	"mundo": {
		Key:         "mundo",
		Title:       "Internacional (Mundo)",
		Emoji:       "🌍",
		Description: "Notícias internacionais (BBC Brasil)",
		FeedURLs: []FeedSource{
			{Source: "BBC News Brasil", URL: "https://feeds.bbci.co.uk/portuguese/rss.xml"},
		},
	},
	"eco": {
		Key:         "eco",
		Title:       "Economia",
		Emoji:       "📈",
		Description: "Mercado e economia (CNN Brasil)",
		FeedURLs: []FeedSource{
			{Source: "CNN Brasil Economia", URL: "https://www.cnnbrasil.com.br/economia/feed/"},
		},
	},
	"tech": {
		Key:         "tech",
		Title:       "Tecnologia",
		Emoji:       "💻",
		Description: "Inovação e tecnologia (Canaltech)",
		FeedURLs: []FeedSource{
			{Source: "Canaltech", URL: "https://canaltech.com.br/rss/"},
		},
	},
}

// NormalizeCategory normaliza aliases de categorias
func NormalizeCategory(input string) string {
	lower := strings.ToLower(strings.TrimSpace(input))
	switch lower {
	case "br", "brasil", "nacional", "geral":
		return "br"
	case "mundo", "internacional", "int", "world":
		return "mundo"
	case "eco", "economia", "mercado", "finance":
		return "eco"
	case "tech", "tecnologia", "canaltech":
		return "tech"
	default:
		return ""
	}
}

// FetchFeed realiza a requisição HTTP e faz o parse do XML RSS.
func FetchFeed(ctx context.Context, source FeedSource, category string) ([]Article, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, source.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("falha ao criar requisição para %s: %w", source.URL, err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/rss+xml, application/xml, text/xml;q=0.9, */*;q=0.8")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("falha na requisição para %s: %w", source.URL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("feed retornou status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erro lendo corpo da resposta: %w", err)
	}

	return ParseRSS(body, source.Source, category)
}

// ParseRSS faz o unmarshal do XML de RSS 2.0 e limpa os campos de texto.
func ParseRSS(data []byte, sourceName string, category string) ([]Article, error) {
	var root rssRoot
	if err := xml.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("erro no parse do XML: %w", err)
	}

	var articles []Article
	for _, item := range root.Channel.Items {
		cleanTitle := cleanText(item.Title)
		cleanDesc := cleanText(item.Description)
		link := strings.TrimSpace(item.Link)

		if cleanTitle == "" || link == "" {
			continue
		}

		pubDate := parseDate(item.PubDate)

		articles = append(articles, Article{
			Title:       cleanTitle,
			Description: cleanDesc,
			URL:         link,
			Source:      sourceName,
			PublishedAt: pubDate,
			Category:    category,
		})
	}

	return articles, nil
}

func cleanText(raw string) string {
	stripped := htmlTagRegex.ReplaceAllString(raw, " ")
	unescaped := html.UnescapeString(stripped)
	cleaned := spaceRegex.ReplaceAllString(strings.TrimSpace(unescaped), " ")
	return cleaned
}

func parseDate(dateStr string) time.Time {
	dateStr = strings.TrimSpace(dateStr)
	if dateStr == "" {
		return time.Now()
	}

	layouts := []string{
		time.RFC1123Z,
		time.RFC1123,
		time.RFC822Z,
		time.RFC822,
		time.RFC3339,
		"2006-01-02T15:04:05-0700",
		"Mon, 2 Jan 2006 15:04:05 -0700",
		"Mon, 02 Jan 2006 15:04:05 MST",
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, dateStr); err == nil {
			return t
		}
	}

	return time.Now()
}
