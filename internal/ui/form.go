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
	Label       string
	Placeholder string
	Value       string
	Required    bool
	Options     []string // if non-empty, rendered as a selector
	optionIdx   int
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
}

func NewForm(title string, fields []FormField) Form {
	inputs := make([]textinput.Model, len(fields))
	for i, f := range fields {
		ti := textinput.New()
		ti.Placeholder = f.Placeholder
		ti.SetValue(f.Value)
		if i == 0 {
			ti.Focus()
		}
		inputs[i] = ti
	}
	return Form{
		Title:  title,
		Fields: fields,
		inputs: inputs,
	}
}

func (f *Form) Update(msg tea.Msg) (Form, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "down":
			f.nextField()
			return *f, nil
		case "shift+tab", "up":
			f.prevField()
			return *f, nil
		default:
			// If current field has options, handle left/right.
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
			// Regular text input.
			var cmd tea.Cmd
			f.inputs[f.focused], cmd = f.inputs[f.focused].Update(msg)
			return *f, cmd
		}
	}
	return *f, nil
}

func (f *Form) nextField() {
	f.inputs[f.focused].Blur()
	f.focused = (f.focused + 1) % len(f.Fields)
	if len(f.Fields[f.focused].Options) == 0 {
		f.inputs[f.focused].Focus()
	}
}

func (f *Form) prevField() {
	f.inputs[f.focused].Blur()
	f.focused = (f.focused - 1 + len(f.Fields)) % len(f.Fields)
	if len(f.Fields[f.focused].Options) == 0 {
		f.inputs[f.focused].Focus()
	}
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
	}

	if f.err != "" {
		rows = append(rows, styleError.Render("  "+f.err))
	}

	rows = append(rows, "")
	rows = append(rows, lipgloss.NewStyle().Foreground(colorSubtext).Render(
		"  tab/↑↓ navigate  •  enter confirm  •  esc cancel",
	))

	content := strings.Join(rows, "\n")

	w := f.Width
	if w < 40 {
		w = 60
	}

	return styleDialog.Width(w).Render(content)
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
