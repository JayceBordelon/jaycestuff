package exec

import (
	"testing"

	"vibetradez.com/internal/trades"
)

func mkTrade(symbol, kind string, rank, score int, price float64) trades.Trade {
	return trades.Trade{
		Symbol:         symbol,
		ContractType:   kind,
		EstimatedPrice: price,
		Rank:           rank,
		Score:          score,
	}
}

func TestQualifyingPick_Rank1HighScoreUnderCap(t *testing.T) {
	in := []trades.Trade{
		mkTrade("AAPL", "CALL", 1, 9, 3.50),
		mkTrade("MSFT", "CALL", 2, 8, 2.10),
	}
	pick, ok := QualifyingPick(in)
	if !ok {
		t.Fatal("expected qualifying pick")
	}
	if pick.Symbol != "AAPL" {
		t.Errorf("got symbol %s want AAPL", pick.Symbol)
	}
}

func TestQualifyingPick_RejectsBelowScoreFloor(t *testing.T) {
	in := []trades.Trade{
		mkTrade("AAPL", "CALL", 1, MinExecutionScore-1, 3.50),
	}
	if _, ok := QualifyingPick(in); ok {
		t.Fatal("expected no pick when score below floor")
	}
}

func TestQualifyingPick_RejectsNonRank1(t *testing.T) {
	in := []trades.Trade{
		mkTrade("AAPL", "CALL", 2, 10, 3.50),
		mkTrade("MSFT", "CALL", 3, 10, 2.10),
	}
	if _, ok := QualifyingPick(in); ok {
		t.Fatal("expected no pick when no trade is rank 1")
	}
}

func TestQualifyingPick_RejectsAbovePriceCap(t *testing.T) {
	in := []trades.Trade{
		mkTrade("NVDA", "CALL", 1, 10, MaxContractPremium+0.01),
	}
	if _, ok := QualifyingPick(in); ok {
		t.Fatal("expected no pick when price exceeds cap")
	}
}

func TestQualifyingPick_AcceptsPriceExactlyAtCap(t *testing.T) {
	in := []trades.Trade{
		mkTrade("NVDA", "CALL", 1, 9, MaxContractPremium),
	}
	if _, ok := QualifyingPick(in); !ok {
		t.Fatal("expected pick at exactly the cap")
	}
}

func TestQualifyingPick_AcceptsPutAtRank1(t *testing.T) {
	in := []trades.Trade{
		mkTrade("AAPL", "PUT", 1, 9, 3.50),
	}
	pick, ok := QualifyingPick(in)
	if !ok || pick.ContractType != "PUT" {
		t.Fatal("expected PUT pick to be selectable")
	}
}

func TestQualifyingPick_EmptyInput(t *testing.T) {
	if _, ok := QualifyingPick(nil); ok {
		t.Fatal("expected no pick from nil input")
	}
	if _, ok := QualifyingPick([]trades.Trade{}); ok {
		t.Fatal("expected no pick from empty input")
	}
}
