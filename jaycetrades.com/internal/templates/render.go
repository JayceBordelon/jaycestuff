package templates

import (
	"bytes"
	"embed"
	"html/template"
	"time"
)

//go:embed email.html summary.html
var templateFS embed.FS

type Trade struct {
	Symbol         string
	ContractType   string
	StrikePrice    float64
	Expiration     string
	DTE            int
	EstimatedPrice float64
	Thesis         string
	SentimentScore float64
	CurrentPrice   float64
	TargetPrice    float64
	StopLoss       float64
	ProfitTarget   float64
	RiskLevel      string
	Catalyst       string
	MentionCount   int
}

type EmailData struct {
	Subject string
	Date    string
	Trades  []Trade
}

type SummaryTrade struct {
	Symbol         string
	ContractType   string
	StrikePrice    float64
	Expiration     string
	EntryPrice     float64
	ClosingPrice   float64
	PriceChange    float64
	PctChange      float64
	StockOpen      float64
	StockClose     float64
	StockPctChange float64
	Result         string
	Notes          string
}

type SummaryEmailData struct {
	Subject     string
	Date        string
	Trades      []SummaryTrade
	TotalTrades int
	Winners     int
	Losers      int
	TotalPnL    float64
}

var funcMap = template.FuncMap{
	"mul": func(a, b float64) float64 { return a * b },
	"div": func(a, b float64) float64 {
		if b == 0 {
			return 0
		}
		return a / b
	},
	"sub": func(a, b float64) float64 { return a - b },
	"add": func(a, b any) any {
		switch av := a.(type) {
		case int:
			if bv, ok := b.(int); ok {
				return av + bv
			}
		case float64:
			if bv, ok := b.(float64); ok {
				return av + bv
			}
		}
		return 0
	},
	"abs": func(a float64) float64 {
		if a < 0 {
			return -a
		}
		return a
	},
	"gt": func(a, b float64) bool { return a > b },
	"lt": func(a, b float64) bool { return a < b },
}

func RenderEmail(trades []Trade) (string, error) {
	tmpl, err := template.New("email.html").Funcs(funcMap).ParseFS(templateFS, "email.html")
	if err != nil {
		return "", err
	}

	data := EmailData{
		Subject: "Today's Top Options Plays",
		Date:    time.Now().Format("Monday, Jan 2, 2006"),
		Trades:  trades,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func RenderSummaryEmail(summaryTrades []SummaryTrade) (string, error) {
	tmpl, err := template.New("summary.html").Funcs(funcMap).ParseFS(templateFS, "summary.html")
	if err != nil {
		return "", err
	}

	winners, losers := 0, 0
	totalPnL := 0.0
	for i := range summaryTrades {
		t := &summaryTrades[i]
		t.PriceChange = t.ClosingPrice - t.EntryPrice
		if t.EntryPrice > 0 {
			t.PctChange = (t.PriceChange / t.EntryPrice) * 100
		}
		if t.StockOpen > 0 {
			t.StockPctChange = ((t.StockClose - t.StockOpen) / t.StockOpen) * 100
		}
		if t.PriceChange > 0 {
			t.Result = "PROFIT"
			winners++
		} else if t.PriceChange < 0 {
			t.Result = "LOSS"
			losers++
		} else {
			t.Result = "FLAT"
		}
		totalPnL += t.PriceChange * 100 // per contract
	}

	data := SummaryEmailData{
		Subject:     "End of Day Trade Summary",
		Date:        time.Now().Format("Monday, Jan 2, 2006"),
		Trades:      summaryTrades,
		TotalTrades: len(summaryTrades),
		Winners:     winners,
		Losers:      losers,
		TotalPnL:    totalPnL,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
