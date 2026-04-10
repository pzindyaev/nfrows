package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pzindyaev/nfrows/internal/nft"
)

func cmdLoadRuleset() tea.Cmd {
	return func() tea.Msg {
		rs, err := nft.ListRuleset()
		if err != nil {
			return rulesetLoadedMsg{err: err}
		}
		return rulesetLoadedMsg{tables: nft.ParseRuleset(rs)}
	}
}

func cmdLoadRuleTexts(family, table, chain string) tea.Cmd {
	return func() tea.Msg {
		return ruleTextsMsg{texts: nft.ListChainRuleTexts(family, table, chain)}
	}
}

func cmdOp(fn func() error) tea.Cmd {
	return func() tea.Msg {
		return opResultMsg{err: fn()}
	}
}
