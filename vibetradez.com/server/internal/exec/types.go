package exec

import "time"

/*
Decision is one row of the daily go/no-go pipeline. There is at most
one Decision per trade_date by schema constraint. Decisions start as
'pending', then transition to 'execute' (user clicked Execute within
the 5-minute window), 'decline' (user clicked Don't Execute), or
'timeout' (window expired without a click).
*/
type Decision struct {
	ID            int
	TradeDate     string
	Symbol        string
	ContractType  string
	StrikePrice   float64
	Expiration    string
	OCCSymbol     string
	ContractPrice float64
	Score         int
	TradeID       int
	TokenHash     string
	Decision      string
	DecidedAt     *time.Time
	ExpiresAt     time.Time
	CreatedAt     time.Time
}

/*
Execution is one order lifecycle. A Decision with decision='execute'
has exactly one Execution with side='open'. If the open fills, the
3:55pm cron creates a second Execution with side='close'. PaperTrader
fills are synthetic (no SchwabOrderID); LiveTrader fills carry the
Schwab order id.
*/
type Execution struct {
	ID                int
	DecisionID        int
	Mode              string
	Side              string
	SchwabOrderID     *string
	Status            string
	FillPrice         *float64
	FilledQuantity    int
	RequestedQuantity int
	SubmittedAt       time.Time
	FilledAt          *time.Time
	ErrorMessage      string
	CreatedAt         time.Time
}
