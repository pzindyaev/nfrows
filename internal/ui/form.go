package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FormField describes a single input in a form.
type FormField struct {
	Label        string
	Placeholder  string
	Value        string
	Required     bool
	Options      []string // if non-empty, rendered as a selector
	Autocomplete bool     // if true, shows nftables rule autocomplete
	optionIdx    int
}

// acState holds the live autocomplete dropdown state.
type acState struct {
	suggestions []string
	selected    int
	visible     bool
}

// Form is a modal dialog with labeled inputs.
type Form struct {
	Title   string
	Fields  []FormField
	inputs  []textinput.Model
	focused int
	err     string
	Width   int
	Height  int

	acField int    // index of the autocomplete-enabled field, or -1
	ac      acState
}

func NewForm(title string, fields []FormField) Form {
	acField := -1
	inputs := make([]textinput.Model, len(fields))
	for i, f := range fields {
		ti := textinput.New()
		ti.Placeholder = f.Placeholder
		ti.SetValue(f.Value)
		if i == 0 {
			ti.Focus()
		}
		inputs[i] = ti
		if f.Autocomplete {
			acField = i
		}
	}
	return Form{
		Title:   title,
		Fields:  fields,
		inputs:  inputs,
		acField: acField,
	}
}

// ConsumeEsc hides the autocomplete dropdown if it is visible and returns
// true, signalling that the modal should NOT be closed yet.
// Returns false when there is nothing for the form to consume.
func (f *Form) ConsumeEsc() bool {
	if f.acField >= 0 && f.ac.visible {
		f.ac.visible = false
		return true
	}
	return false
}

func (f *Form) Update(msg tea.Msg) (Form, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// ── autocomplete-aware navigation ────────────────────────────────────
		inACField := f.acField >= 0 && f.focused == f.acField

		switch msg.String() {
		case "up":
			if inACField && f.ac.visible {
				if f.ac.selected > 0 {
					f.ac.selected--
				}
				return *f, nil
			}
			f.prevField()
			f.syncACOnFocusChange()
			return *f, nil

		case "down":
			if inACField && f.ac.visible {
				if f.ac.selected < len(f.ac.suggestions)-1 {
					f.ac.selected++
				}
				return *f, nil
			}
			f.nextField()
			f.syncACOnFocusChange()
			return *f, nil

		case "tab":
			if inACField && f.ac.visible && len(f.ac.suggestions) > 0 {
				f.completeSelected()
				return *f, nil
			}
			f.nextField()
			f.syncACOnFocusChange()
			return *f, nil

		case "shift+tab":
			if inACField && f.ac.visible {
				f.ac.visible = false
				return *f, nil
			}
			f.prevField()
			f.syncACOnFocusChange()
			return *f, nil

		default:
			// Option-selector fields (h/l to cycle options).
			if len(f.Fields[f.focused].Options) > 0 {
				switch msg.String() {
				case "left", "h":
					if f.Fields[f.focused].optionIdx > 0 {
						f.Fields[f.focused].optionIdx--
					}
					return *f, nil
				case "right", "l":
					opts := f.Fields[f.focused].Options
					if f.Fields[f.focused].optionIdx < len(opts)-1 {
						f.Fields[f.focused].optionIdx++
					}
					return *f, nil
				}
				return *f, nil
			}

			// Regular text input – forward the key then refresh autocomplete.
			var cmd tea.Cmd
			f.inputs[f.focused], cmd = f.inputs[f.focused].Update(msg)
			if inACField {
				f.refreshAC()
			}
			return *f, cmd
		}
	}
	return *f, nil
}

// nextField advances focus to the next field.
func (f *Form) nextField() {
	f.inputs[f.focused].Blur()
	f.focused = (f.focused + 1) % len(f.Fields)
	if len(f.Fields[f.focused].Options) == 0 {
		f.inputs[f.focused].Focus()
	}
}

// prevField moves focus to the previous field.
func (f *Form) prevField() {
	f.inputs[f.focused].Blur()
	f.focused = (f.focused - 1 + len(f.Fields)) % len(f.Fields)
	if len(f.Fields[f.focused].Options) == 0 {
		f.inputs[f.focused].Focus()
	}
}

// syncACOnFocusChange hides the dropdown when moving away from the AC field.
func (f *Form) syncACOnFocusChange() {
	if f.acField < 0 {
		return
	}
	if f.focused != f.acField {
		f.ac.visible = false
	}
}

// refreshAC recomputes suggestions from the current value of the AC field.
func (f *Form) refreshAC() {
	if f.acField < 0 {
		return
	}
	val := f.inputs[f.acField].Value()
	suggestions := GetRuleSuggestions(val)
	f.ac.suggestions = suggestions
	if f.ac.selected >= len(suggestions) {
		f.ac.selected = 0
	}
	f.ac.visible = len(suggestions) > 0
}

// completeSelected replaces the current prefix in the AC field with the
// selected suggestion and advances the cursor to the end.
func (f *Form) completeSelected() {
	if f.acField < 0 || !f.ac.visible || len(f.ac.suggestions) == 0 {
		return
	}
	selected := f.ac.suggestions[f.ac.selected]

	input := f.inputs[f.acField].Value()
	prevTokens, _, endsWithSpace := tokenizeRuleInput(input)

	var newVal string
	base := strings.Join(prevTokens, " ")
	if base != "" || endsWithSpace {
		newVal = base + " " + selected + " "
	} else {
		newVal = selected + " "
	}

	f.inputs[f.acField].SetValue(newVal)
	// Move cursor to end.
	for i := 0; i < len(newVal); i++ {
		f.inputs[f.acField].CursorEnd()
	}

	f.refreshAC()
}

// Values returns the current field values as a slice of strings.
func (f *Form) Values() []string {
	vals := make([]string, len(f.Fields))
	for i, field := range f.Fields {
		if len(field.Options) > 0 {
			vals[i] = field.Options[field.optionIdx]
		} else {
			vals[i] = f.inputs[i].Value()
		}
	}
	return vals
}

// Validate checks required fields and sets f.err. Returns true if valid.
func (f *Form) Validate() bool {
	for i, field := range f.Fields {
		var val string
		if len(field.Options) > 0 {
			val = field.Options[field.optionIdx]
		} else {
			val = f.inputs[i].Value()
		}
		if field.Required && strings.TrimSpace(val) == "" {
			f.err = fmt.Sprintf("%q is required", field.Label)
			return false
		}
	}
	f.err = ""
	return true
}

func (f *Form) SetError(err string) { f.err = err }

func (f *Form) View() string {
	var rows []string

	title := styleDialogTitle.Render(f.Title)
	rows = append(rows, title)

	for i, field := range f.Fields {
		label := lipgloss.NewStyle().Foreground(colorSubtext).Render(field.Label + ":")
		var input string
		if len(field.Options) > 0 {
			input = f.renderSelector(i)
		} else {
			input = f.inputs[i].View()
		}
		focused := i == f.focused
		row := label + " " + input
		if focused {
			row = lipgloss.NewStyle().Foreground(colorText).Render(row)
		} else {
			row = lipgloss.NewStyle().Foreground(colorSubtext).Render(row)
		}
		rows = append(rows, row)

		// Render autocomplete dropdown directly below the focused AC field.
		if i == f.acField && focused && f.ac.visible && len(f.ac.suggestions) > 0 {
			rows = append(rows, f.renderACDropdown())
		}
	}

	if f.err != "" {
		rows = append(rows, styleError.Render("  "+f.err))
	}

	rows = append(rows, "")

	// Hint line: adapt when autocomplete is active.
	var hint string
	if f.acField >= 0 && f.focused == f.acField && f.ac.visible {
		hint = "  ↑↓ select suggestion  •  tab complete  •  esc dismiss  •  enter confirm"
	} else {
		hint = "  tab/↑↓ navigate  •  enter confirm  •  esc cancel"
	}
	rows = append(rows, lipgloss.NewStyle().Foreground(colorSubtext).Render(hint))

	content := strings.Join(rows, "\n")

	w := f.Width
	if w < 40 {
		w = 60
	}

	return styleDialog.Width(w).Render(content)
}

func (f *Form) renderACDropdown() string {
	const maxVisible = 8

	total := len(f.ac.suggestions)
	start := 0
	end := total
	if total > maxVisible {
		// Scroll window so the selected item stays visible.
		start = f.ac.selected - maxVisible + 1
		if start < 0 {
			start = 0
		}
		end = start + maxVisible
		if end > total {
			end = total
			start = end - maxVisible
		}
	}

	var lines []string
	for i := start; i < end; i++ {
		s := f.ac.suggestions[i]
		var rendered string
		if i == f.ac.selected {
			rendered = styleACItemSelected.Render(s)
		} else {
			rendered = styleACItem.Render(s)
		}
		lines = append(lines, rendered)
	}

	// Add a scroll indicator when there are more items.
	if total > maxVisible {
		indicator := fmt.Sprintf("  %d/%d", f.ac.selected+1, total)
		lines = append(lines, lipgloss.NewStyle().Foreground(colorSubtext).Render(indicator))
	}

	dropdown := strings.Join(lines, "\n")
	return styleACBorder.Render(dropdown)
}

func (f *Form) renderSelector(fieldIdx int) string {
	field := f.Fields[fieldIdx]
	focused := fieldIdx == f.focused
	var parts []string
	for i, opt := range field.Options {
		if i == field.optionIdx {
			if focused {
				parts = append(parts, styleTabActive.Render(opt))
			} else {
				parts = append(parts, lipgloss.NewStyle().
					Foreground(colorAccent).Bold(true).Padding(0, 1).Render(opt))
			}
		} else {
			parts = append(parts, styleTab.Render(opt))
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}
