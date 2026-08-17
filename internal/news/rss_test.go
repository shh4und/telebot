package news

import (
	"testing"
)

const sampleRSS = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
	<channel>
		<title>Notícias Teste</title>
		<link>https://example.com</link>
		<description>Canal de Teste</description>
		<item>
			<title><![CDATA[Governo anuncia novo plano econ&ocirc;mico]]></title>
			<link>https://example.com/noticia-1</link>
			<description><![CDATA[<p>Medida visa impulsionar investimentos no setor industrial.</p>]]></description>
			<pubDate>Mon, 17 Aug 2026 12:00:00 -0300</pubDate>
		</item>
		<item>
			<title>Segunda notícia sem CDATA &amp; com tags &lt;b&gt;HTML&lt;/b&gt;</title>
			<link>https://example.com/noticia-2</link>
			<description>Descricao curta sem HTML</description>
			<pubDate>Mon, 17 Aug 2026 11:30:00 -0300</pubDate>
		</item>
	</channel>
</rss>`

func TestParseRSS(t *testing.T) {
	articles, err := ParseRSS([]byte(sampleRSS), "Fonte Teste", "br")
	if err != nil {
		t.Fatalf("esperava sucesso no parsing de RSS, obteve erro: %v", err)
	}

	if len(articles) != 2 {
		t.Fatalf("esperava 2 artigos, obteve %d", len(articles))
	}

	first := articles[0]
	if first.Title != "Governo anuncia novo plano econômico" {
		t.Errorf("título incorreto: esperado 'Governo anuncia novo plano econômico', obteve '%s'", first.Title)
	}
	if first.Description != "Medida visa impulsionar investimentos no setor industrial." {
		t.Errorf("descrição incorreta: obteve '%s'", first.Description)
	}
	if first.URL != "https://example.com/noticia-1" {
		t.Errorf("URL incorreta: obteve '%s'", first.URL)
	}
	if first.Source != "Fonte Teste" {
		t.Errorf("Fonte incorreta: obteve '%s'", first.Source)
	}

	second := articles[1]
	if second.Title != "Segunda notícia sem CDATA & com tags HTML" {
		t.Errorf("título 2 incorreto: obteve '%s'", second.Title)
	}
}

func TestNormalizeCategory(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"br", "br"},
		{"BRASIL", "br"},
		{"nacional", "br"},
		{"mundo", "mundo"},
		{"world", "mundo"},
		{"eco", "eco"},
		{"economia", "eco"},
		{"tech", "tech"},
		{"tecnologia", "tech"},
		{"invalido", ""},
	}

	for _, tt := range tests {
		result := NormalizeCategory(tt.input)
		if result != tt.expected {
			t.Errorf("NormalizeCategory(%q) = %q; esperado %q", tt.input, result, tt.expected)
		}
	}
}
