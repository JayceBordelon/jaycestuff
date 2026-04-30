package exec

import "vibetradez.com/internal/trades"

/*
MaxContractPremium is the hard upper bound on a single contract's
premium for auto-execution. Quarter of the existing $200 model-prompt
cap; the lower number reflects "we're auto-trading this so the cap
must be aggressive". Worst-case daily loss = MaxContractPremium *
MaxContracts * 100 (dollars per option contract multiplier) per
position; with the daily cap of 1 position via UNIQUE(trade_date) the
worst-case daily loss is MaxContractPremium * 100 = $500 for a $5
contract. Premium quoted in this codebase is per-share, options are
100 shares, so $5 premium = $500 capital exposure per contract.
Adjust both this constant and the prompt language together if the cap
ever changes.
*/
const MaxContractPremium = 5.00

/*
MinExecutionScore is the conviction floor for auto-execution. With the
single-model pipeline, a high Claude conviction score is the only signal
that replaces the prior "both models ranked it #1" consensus gate.
Anything below 9/10 stays advisory.
*/
const MinExecutionScore = 9

/*
QualifyingPick returns the trade that should be auto-executed today,
if any. The qualification rule:
  - Claude ranked the trade #1 for the day
  - Claude's conviction score is >= MinExecutionScore
  - the contract premium is at or below MaxContractPremium ($5/share = $500/contract)

If no trade meets all three criteria, the function returns (nil, false)
and the day is skipped, no email, no order, no DB row.
*/
func QualifyingPick(picks []trades.Trade) (*trades.Trade, bool) {
	for i := range picks {
		t := &picks[i]
		if t.Rank != 1 {
			continue
		}
		if t.Score < MinExecutionScore {
			continue
		}
		if t.EstimatedPrice > MaxContractPremium {
			continue
		}
		return t, true
	}
	return nil, false
}
