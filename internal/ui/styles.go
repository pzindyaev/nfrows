package ui

import "github.com/charmbracelet/lipgloss"

var (
	colorPrimary   = lipgloss.Color("#7C3AED")
	colorSecondary = lipgloss.Color("#4C1D95")
	colorAccent    = lipgloss.Color("#A78BFA")
	colorMuted     = lipgloss.Color("#6B7280")
	colorDanger    = lipgloss.Color("#EF4444")
	colorSuccess   = lipgloss.Color("#10B981")
	colorWarning   = lipgloss.Color("#F59E0B")
	colorBg        = lipgloss.Color("#1E1E2E")
	colorSurface   = lipgloss.Color("#2A2A3E")
	colorBorder    = lipgloss.Color("#44415A")
	colorText      = lipgloss.Color("#CDD6F4")
	colorSubtext   = lipgloss.Color("#6C7086")

	styleBase = lipgloss.NewStyle().
			Background(colorBg).
			Foreground(colorText)

	styleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorAccent).
			Padding(0, 1)

	styleBreadcrumb = lipgloss.NewStyle().
			Foreground(colorSubtext).
			Padding(0, 1)

	styleBreadcrumbActive = lipgloss.NewStyle().
				Foreground(colorAccent).
				Bold(true)

	styleStatusBar = lipgloss.NewStyle().
			Background(colorSecondary).
			Foreground(colorText).
			Padding(0, 1)

	styleKeyHint = lipgloss.NewStyle().
			Foreground(colorAccent).
			Bold(true)

	styleKeyDesc = lipgloss.NewStyle().
			Foreground(colorSubtext)

	styleTableHeader = lipgloss.NewStyle().
				Background(colorSecondary).
				Foreground(colorAccent).
				Bold(true).
				Padding(0, 1)

	styleTableRow = lipgloss.NewStyle().
			Foreground(colorText).
			Padding(0, 1)

	styleTableRowSelected = lipgloss.NewStyle().
				Background(colorPrimary).
				Foreground(colorText).
				Bold(true).
				Padding(0, 1)

	styleTableBorder = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(colorBorder)

	styleTab = lipgloss.NewStyle().
			Foreground(colorSubtext).
			Padding(0, 2)

	styleTabActive = lipgloss.NewStyle().
			Foreground(colorBg).
			Background(colorPrimary).
			Bold(true).
			Padding(0, 2)

	styleTabBorder = lipgloss.NewStyle().
			BorderBottom(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottomForeground(colorBorder)

	styleDialog = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPrimary).
			Background(colorSurface).
			Padding(1, 2)

	styleDialogTitle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorAccent).
				MarginBottom(1)

	styleError = lipgloss.NewStyle().
			Foreground(colorDanger).
			Bold(true)

	styleSuccess = lipgloss.NewStyle().
			Foreground(colorSuccess).
			Bold(true)

	styleFamilyTag = map[string]lipgloss.Style{
		"ip":     lipgloss.NewStyle().Foreground(lipgloss.Color("#60A5FA")),
		"ip6":    lipgloss.NewStyle().Foreground(lipgloss.Color("#818CF8")),
		"inet":   lipgloss.NewStyle().Foreground(lipgloss.Color("#34D399")),
		"arp":    lipgloss.NewStyle().Foreground(lipgloss.Color("#FBBF24")),
		"bridge": lipgloss.NewStyle().Foreground(lipgloss.Color("#F472B6")),
		"netdev": lipgloss.NewStyle().Foreground(lipgloss.Color("#FB923C")),
	}

	stylePolicyAccept = lipgloss.NewStyle().Foreground(colorSuccess)
	stylePolicyDrop   = lipgloss.NewStyle().Foreground(colorDanger)

	styleACItem = lipgloss.NewStyle().
			Foreground(colorSubtext).
			PaddingLeft(2)

	styleACItemSelected = lipgloss.NewStyle().
				Background(colorPrimary).
				Foreground(colorText).
				Bold(true).
				PaddingLeft(2)

	styleACBorder = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Background(colorSurface).
			PaddingLeft(1).
			PaddingRight(1)
)

func familyStyle(family string) lipgloss.Style {
	if s, ok := styleFamilyTag[family]; ok {
		return s
	}
	return lipgloss.NewStyle().Foreground(colorMuted)
}

func policyStyle(policy string) lipgloss.Style {
	switch policy {
	case "accept":
		return stylePolicyAccept
	case "drop":
		return stylePolicyDrop
	default:
		return lipgloss.NewStyle().Foreground(colorSubtext)
	}
}
