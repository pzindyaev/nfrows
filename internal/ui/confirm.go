package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// ConfirmDialog is a yes/no prompt.
type ConfirmDialog struct {
	Message string
	Width   int
}

func (c ConfirmDialog) View() string {
	body := fmt.Sprintf("%s\n\n%s  %s",
		styleDialogTitle.Render(c.Message),
		styleKeyHint.Render("y")+" "+styleKeyDesc.Render("yes"),
		styleKeyHint.Render("n/esc")+" "+styleKeyDesc.Render("no"),
	)
	w := c.Width
	if w < 40 {
		w = 50
	}
	return lipgloss.Place(
		w, 7,
		lipgloss.Center, lipgloss.Center,
		styleDialog.Render(body),
	)
}
