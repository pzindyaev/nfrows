package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Column defines a table column.
type Column struct {
	Title string
	Width int
	Flex  bool // if true, take remaining space
}

// Row is a slice of cell strings.
type Row []string

// TableWidget is a simple table renderer.
type TableWidget struct {
	Columns  []Column
	Rows     []Row
	Selected int
	Offset   int
	Height   int // visible rows (excluding header)
	Width    int
}

func (t *TableWidget) MoveUp() {
	if t.Selected > 0 {
		t.Selected--
		if t.Selected < t.Offset {
			t.Offset = t.Selected
		}
	}
}

func (t *TableWidget) MoveDown() {
	if t.Selected < len(t.Rows)-1 {
		t.Selected++
		if t.Selected >= t.Offset+t.Height {
			t.Offset = t.Selected - t.Height + 1
		}
	}
}

func (t *TableWidget) GoToTop() {
	t.Selected = 0
	t.Offset = 0
}

func (t *TableWidget) GoToBottom() {
	if len(t.Rows) == 0 {
		return
	}
	t.Selected = len(t.Rows) - 1
	t.Offset = max(0, t.Selected-t.Height+1)
}

func (t *TableWidget) HalfPageUp() {
	step := max(1, t.Height/2)
	t.Selected = max(0, t.Selected-step)
	if t.Selected < t.Offset {
		t.Offset = t.Selected
	}
}

func (t *TableWidget) HalfPageDown() {
	if len(t.Rows) == 0 {
		return
	}
	step := max(1, t.Height/2)
	t.Selected = min(len(t.Rows)-1, t.Selected+step)
	if t.Selected >= t.Offset+t.Height {
		t.Offset = t.Selected - t.Height + 1
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (t *TableWidget) SetRows(rows []Row) {
	t.Rows = rows
	if t.Selected >= len(rows) {
		t.Selected = max(0, len(rows)-1)
	}
	if t.Offset > t.Selected {
		t.Offset = t.Selected
	}
}

func (t *TableWidget) SelectedRow() (Row, bool) {
	if len(t.Rows) == 0 || t.Selected >= len(t.Rows) {
		return nil, false
	}
	return t.Rows[t.Selected], true
}

func (t *TableWidget) View() string {
	// Compute column widths.
	widths := t.computeWidths()

	var sb strings.Builder

	// Header.
	headerCells := make([]string, len(t.Columns))
	for i, col := range t.Columns {
		cell := truncate(col.Title, widths[i])
		cell = pad(cell, widths[i])
		headerCells[i] = styleTableHeader.Render(cell)
	}
	sb.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, headerCells...))
	sb.WriteString("\n")

	// Separator between header and rows.
	sepStyle := lipgloss.NewStyle().Foreground(colorBorder)
	totalW := 0
	for _, w := range widths {
		totalW += w + 2 // +2 for cell padding
	}
	sb.WriteString(sepStyle.Render(strings.Repeat("─", totalW)))
	sb.WriteString("\n")

	// Rows.
	visible := t.Rows
	if t.Offset < len(visible) {
		visible = visible[t.Offset:]
	}
	if len(visible) > t.Height {
		visible = visible[:t.Height]
	}

	for idx, row := range visible {
		absIdx := idx + t.Offset
		cells := make([]string, len(t.Columns))
		for i := range t.Columns {
			var cell string
			if i < len(row) {
				cell = row[i]
			}
			cell = truncate(cell, widths[i])
			cell = pad(cell, widths[i])
			if absIdx == t.Selected {
				cells[i] = styleTableRowSelected.Render(cell)
			} else {
				cells[i] = styleTableRow.Render(cell)
			}
		}
		sb.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, cells...))
		sb.WriteString("\n")
	}

	// Empty rows to fill height.
	for i := len(visible); i < t.Height; i++ {
		cells := make([]string, len(t.Columns))
		for j, w := range widths {
			cells[j] = styleTableRow.Render(strings.Repeat(" ", w))
		}
		sb.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, cells...))
		sb.WriteString("\n")
	}

	// Trim the trailing newline so that lipgloss does not create a phantom
	// empty row when rendering the content inside a border box.
	return strings.TrimRight(sb.String(), "\n")
}

func (t *TableWidget) computeWidths() []int {
	widths := make([]int, len(t.Columns))
	flexIdx := -1
	fixed := 0

	for i, col := range t.Columns {
		if col.Flex {
			flexIdx = i
		} else {
			widths[i] = col.Width
			fixed += col.Width + 2 // +2 for padding
		}
	}

	if flexIdx >= 0 {
		remaining := t.Width - fixed - 2
		if remaining < 10 {
			remaining = 10
		}
		widths[flexIdx] = remaining
	}

	return widths
}

func truncate(s string, w int) string {
	if w <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= w {
		return s
	}
	if w <= 3 {
		return string(runes[:w])
	}
	return string(runes[:w-3]) + "..."
}

func pad(s string, w int) string {
	runes := []rune(s)
	n := w - len(runes)
	if n <= 0 {
		return s
	}
	return s + strings.Repeat(" ", n)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
