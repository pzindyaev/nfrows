package ui

import (
	"github.com/pzindyaev/nfrows/internal/nft"
)

// rulesetLoadedMsg carries a freshly loaded ruleset.
type rulesetLoadedMsg struct {
	tables map[string]*nft.TableData
	err    error
}

// ruleTextsMsg carries rule handle→text for a chain.
type ruleTextsMsg struct {
	texts map[int]string
}

// opResultMsg carries the outcome of a mutation operation.
type opResultMsg struct {
	err error
}
