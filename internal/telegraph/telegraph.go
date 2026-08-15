package telegraph

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/yuin/goldmark"
	"golang.org/x/net/html"
)

type NodeElement struct {
	Tag      string            `json:"tag"`
	Attrs    map[string]string `json:"attrs,omitempty"`
	Children []any             `json:"children,omitempty"`
}

type Account struct {
	ShortName   string `json:"short_name"`
	AuthorName  string `json:"author_name"`
	AuthorUrl   string `json:"author_url"`
	AccessToken string `json:"access_token,omitempty"`
	AuthUrl     string `json:"auth_url,omitempty"`
	PageCount   int    `json:"page_count,omitempty"`
}

type Page struct {
	Path        string `json:"path"`
	Url         string `json:"url"`
	Title       string `json:"title"`
	Description string `json:"description"`
	AuthorName  string `json:"author_name,omitempty"`
	AuthorUrl   string `json:"author_url,omitempty"`
	ImageUrl    string `json:"image_url,omitempty"`
	Content     []any  `json:"content,omitempty"`
	Views       int    `json:"views"`
	CanEdit     bool   `json:"can_edit,omitempty"`
}

var (
	tokenMutex        sync.Mutex
	cachedAccessToken string
	DefaultShortName  = "TelegramBot"
	DefaultAuthorName = "Ollama AI"
)

var allowedTelegraphTags = map[string]string{
	"a":          "a",
	"aside":      "aside",
	"b":          "b",
	"blockquote": "blockquote",
	"br":         "br",
	"code":       "code",
	"em":         "em",
	"figcaption": "figcaption",
	"figure":     "figure",
	"h1":         "h3",
	"h2":         "h3",
	"h3":         "h3",
	"h4":         "h4",
	"h5":         "h4",
	"h6":         "h4",
	"hr":         "hr",
	"i":          "em",
	"iframe":     "iframe",
	"img":        "img",
	"li":         "li",
	"ol":         "ol",
	"p":          "p",
	"pre":        "pre",
	"s":          "s",
	"strong":     "strong",
	"u":          "u",
	"ul":         "ul",
	"video":      "video",
}

// CreateAccount cria uma nova conta no Telegraph
func CreateAccount(shortName, authorName, authorUrl string) (*Account, error) {
	data := url.Values{}
	data.Set("short_name", shortName)
	if authorName != "" {
		data.Set("author_name", authorName)
	}
	if authorUrl != "" {
		data.Set("author_url", authorUrl)
	}

	resp, err := http.PostForm("https://api.telegra.ph/createAccount", data)
	if err != nil {
		return nil, fmt.Errorf("http post error: %w", err)
	}
	defer resp.Body.Close()

	var apiResp struct {
		Ok     bool    `json:"ok"`
		Result Account `json:"result"`
		Error  string  `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	if !apiResp.Ok {
		return nil, fmt.Errorf("telegraph api error: %s", apiResp.Error)
	}

	return &apiResp.Result, nil
}

// CreatePage publica uma nova página na API do Telegraph
func CreatePage(accessToken, title, authorName, authorUrl string, content []any) (*Page, error) {
	contentBytes, err := json.Marshal(content)
	if err != nil {
		return nil, fmt.Errorf("error marshaling content: %w", err)
	}

	data := url.Values{}
	data.Set("access_token", accessToken)
	data.Set("title", title)
	if authorName != "" {
		data.Set("author_name", authorName)
	}
	if authorUrl != "" {
		data.Set("author_url", authorUrl)
	}
	data.Set("content", string(contentBytes))

	resp, err := http.PostForm("https://api.telegra.ph/createPage", data)
	if err != nil {
		return nil, fmt.Errorf("http post error: %w", err)
	}
	defer resp.Body.Close()

	var apiResp struct {
		Ok     bool   `json:"ok"`
		Result Page   `json:"result"`
		Error  string `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	if !apiResp.Ok {
		return nil, fmt.Errorf("telegraph api error: %s", apiResp.Error)
	}

	return &apiResp.Result, nil
}

// MarkdownToNodes converte uma string Markdown em um slice de nós DOM do Telegraph ([]any)
func MarkdownToNodes(markdownText string) ([]any, error) {
	trimmed := strings.TrimSpace(markdownText)
	if strings.HasPrefix(trimmed, "{") {
		var jsonObj map[string]any
		if err := json.Unmarshal([]byte(trimmed), &jsonObj); err == nil {
			for _, key := range []string{"response", "answer", "text", "content", "mensagem"} {
				if val, ok := jsonObj[key].(string); ok && val != "" {
					markdownText = val
					break
				}
			}
		}
	}

	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(markdownText), &buf); err != nil {
		return nil, fmt.Errorf("goldmark convert error: %w", err)
	}

	doc, err := html.Parse(strings.NewReader(buf.String()))
	if err != nil {
		return nil, fmt.Errorf("html parse error: %w", err)
	}

	var body *html.Node
	var findBody func(*html.Node)
	findBody = func(n *html.Node) {
		if n.Type == html.ElementNode && strings.ToLower(n.Data) == "body" {
			body = n
			return
		}
		for c := n.FirstChild; c != nil && body == nil; c = c.NextSibling {
			findBody(c)
		}
	}
	findBody(doc)

	if body == nil {
		body = doc
	}

	nodes := htmlToTelegraphNodes(body)
	if len(nodes) == 0 {
		nodes = []any{markdownText}
	}
	return nodes, nil
}

func htmlToTelegraphNodes(n *html.Node) []any {
	var nodes []any
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		converted := convertHTMLNode(c)
		if converted != nil {
			nodes = append(nodes, converted)
		}
	}
	return nodes
}

func convertHTMLNode(n *html.Node) any {
	switch n.Type {
	case html.TextNode:
		text := n.Data
		if text == "" {
			return nil
		}
		return text

	case html.ElementNode:
		tag, allowed := allowedTelegraphTags[strings.ToLower(n.Data)]
		if !allowed {
			children := htmlToTelegraphNodes(n)
			if len(children) == 1 {
				return children[0]
			}
			if len(children) > 0 {
				return NodeElement{
					Tag:      "p",
					Children: children,
				}
			}
			return nil
		}

		elem := NodeElement{
			Tag: tag,
		}

		attrs := make(map[string]string)
		for _, a := range n.Attr {
			attrName := strings.ToLower(a.Key)
			if (tag == "a" && attrName == "href") || ((tag == "img" || tag == "iframe" || tag == "video") && attrName == "src") {
				attrs[attrName] = a.Val
			}
		}
		if len(attrs) > 0 {
			elem.Attrs = attrs
		}

		children := htmlToTelegraphNodes(n)
		if len(children) > 0 {
			elem.Children = children
		}

		return elem

	default:
		return nil
	}
}

// PublishMarkdown faz o fluxo completo: garante conta do Telegraph, converte Markdown e publica a página
func PublishMarkdown(title, markdownText string) (string, error) {
	tokenMutex.Lock()
	if cachedAccessToken == "" {
		acc, err := CreateAccount(DefaultShortName, DefaultAuthorName, "")
		if err != nil {
			tokenMutex.Unlock()
			return "", fmt.Errorf("failed to create telegraph account: %w", err)
		}
		cachedAccessToken = acc.AccessToken
	}
	tok := cachedAccessToken
	tokenMutex.Unlock()

	nodes, err := MarkdownToNodes(markdownText)
	if err != nil {
		return "", fmt.Errorf("failed to convert markdown: %w", err)
	}

	if title == "" {
		title = "Resposta da IA"
	}
	if len(title) > 256 {
		title = title[:253] + "..."
	}

	page, err := CreatePage(tok, title, DefaultAuthorName, "", nodes)
	if err != nil {
		return "", fmt.Errorf("failed to create page: %w", err)
	}

	return page.Url, nil
}
