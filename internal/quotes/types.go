package quotes

import "time"

// CurrencyQuote dados de cotação de moeda fiduciária (ex: USD, EUR).
type CurrencyQuote struct {
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	Bid       float64   `json:"bid"`
	Ask       float64   `json:"ask"`
	High      float64   `json:"high"`
	Low       float64   `json:"low"`
	PctChange float64   `json:"pctChange"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CryptoQuote dados de cotação de criptomoeda (ex: BTC, ETH, SOL).
type CryptoQuote struct {
	ID        string  `json:"id"`
	Symbol    string  `json:"symbol"`
	Name      string  `json:"name"`
	PriceBRL  float64 `json:"price_brl"`
	PriceUSD  float64 `json:"price_usd"`
	Change24h float64 `json:"change_24h"`
}

// MarketSummary resumo consolidado de moedas e criptomoedas.
type MarketSummary struct {
	Dollar    CurrencyQuote
	Euro      CurrencyQuote
	Cryptos   []CryptoQuote
	UpdatedAt time.Time
}
