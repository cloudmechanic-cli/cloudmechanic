package tui

import "github.com/charmbracelet/lipgloss"

// ── Catppuccin Macchiato palette ──────────────────────────────────────────────
var (
	// Surfaces (dark → light)
	catBase     = lipgloss.Color("#24273a")
	catMantle   = lipgloss.Color("#1e2030")
	catSurface0 = lipgloss.Color("#363a4f")
	catSurface1 = lipgloss.Color("#494d64")
	catSurface2 = lipgloss.Color("#5b6078")
	catOverlay0 = lipgloss.Color("#6e738d")
	catOverlay1 = lipgloss.Color("#8087a2")
	catOverlay2 = lipgloss.Color("#939ab7")

	// Text
	catText     = lipgloss.Color("#cad3f5")
	catSubtext1 = lipgloss.Color("#b8c0e0")
	catSubtext0 = lipgloss.Color("#a5adcb")

	// Accent colours
	catLavender = lipgloss.Color("#b7bdf8")
	catBlue     = lipgloss.Color("#8aadf4")
	catSapphire = lipgloss.Color("#7dc4e4")
	catTeal     = lipgloss.Color("#8bd5ca")
	catGreen    = lipgloss.Color("#a6da95")
	catYellow   = lipgloss.Color("#eed49f")
	catPeach    = lipgloss.Color("#f5a97f")
	catRed      = lipgloss.Color("#ed8796")
	catMauve    = lipgloss.Color("#c6a0f6")
)

// Compatibility aliases — used across view.go and remediation.go.
var (
	colorCritical = catRed
	colorWarning  = catYellow
	colorGreen    = catGreen
	colorCyan     = catSapphire
	colorDim      = catOverlay0
	colorWhite    = catText
	colorSelected = catSurface1
	colorSearch   = catMauve
)

// ── Header ────────────────────────────────────────────────────────────────────

// headerStyle kept for compilation (no longer renders ASCII art).
var headerStyle = lipgloss.NewStyle()

var (
	headerBarStyle = lipgloss.NewStyle().
			Background(catMantle).
			Padding(0, 1)

	headerTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(catBlue)

	headerDotStyle = lipgloss.NewStyle().
			Foreground(catSurface2)

	headerSubStyle = lipgloss.NewStyle().
			Foreground(catSubtext0)
)

// ── Tab bar ───────────────────────────────────────────────────────────────────

var (
	tabActiveStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(catText).
			Background(catSurface1).
			Padding(0, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(catBlue)

	tabInactiveStyle = lipgloss.NewStyle().
				Foreground(catOverlay1).
				Padding(0, 2).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(catSurface0)
)

// ── Pane borders ──────────────────────────────────────────────────────────────

var (
	focusedBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(catBlue)

	unfocusedBorder = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(catSurface0)

	// Applied to content inside the inactive pane.
	faintStyle = lipgloss.NewStyle().Faint(true)
)

// ── Sidebar ───────────────────────────────────────────────────────────────────

var (
	sidebarTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(catLavender).
				Padding(0, 1)

	regionSelectedStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(catBase).
				Background(catBlue).
				Padding(0, 1)

	regionNormalStyle = lipgloss.NewStyle().
				Foreground(catText).
				Padding(0, 1)

	regionDimStyle = lipgloss.NewStyle().
			Foreground(catOverlay1).
			Padding(0, 1)
)

// ── Issue list ────────────────────────────────────────────────────────────────

var (
	criticalBadge = lipgloss.NewStyle().
			Bold(true).
			Foreground(catBase).
			Background(catRed).
			Padding(0, 1)

	warningBadge = lipgloss.NewStyle().
			Bold(true).
			Foreground(catBase).
			Background(catYellow).
			Padding(0, 1)

	issueNormalStyle = lipgloss.NewStyle().
				Foreground(catSubtext1).
				Padding(0, 1)

	issueSelectedStyle = lipgloss.NewStyle().
				Background(catSurface1).
				Foreground(catText).
				Bold(true).
				Padding(0, 1)

	tableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(catLavender).
				Padding(0, 1).
				Underline(true)

	groupHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(catMauve).
				Padding(0, 1).
				MarginTop(1)
)

// ── Detail pane (kept for compatibility) ─────────────────────────────────────

var (
	detailLabelStyle = lipgloss.NewStyle().
				Foreground(catSapphire).
				Bold(true)

	detailValueStyle = lipgloss.NewStyle().
				Foreground(catText)

	suggestionStyle = lipgloss.NewStyle().
			Foreground(catGreen).
			Italic(true)
)

// ── Status bar ────────────────────────────────────────────────────────────────

var (
	statusBarStyle = lipgloss.NewStyle().
			Background(catMantle).
			Foreground(catSubtext0)

	// Pill: key cap (solid accent background).
	statusKeyStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(catBase).
			Background(catSapphire).
			Padding(0, 1)

	// Pill: action label (darker background).
	statusActionStyle = lipgloss.NewStyle().
				Foreground(catSubtext0).
				Background(catSurface0).
				Padding(0, 1)

	// Left-side region badge.
	statusRegionStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(catBase).
				Background(catTeal).
				Padding(0, 1)

	filterActiveStyle = lipgloss.NewStyle().
				Foreground(catMauve).
				Bold(true)

	searchStyle = lipgloss.NewStyle().
			Foreground(catMauve).
			Bold(true)
)

// ── Misc ──────────────────────────────────────────────────────────────────────

var (
	spinnerStyle = lipgloss.NewStyle().
			Foreground(catBlue)

	errorStyle = lipgloss.NewStyle().
			Foreground(catRed)

	countCriticalStyle = lipgloss.NewStyle().
				Foreground(catRed).
				Bold(true)

	countWarningStyle = lipgloss.NewStyle().
				Foreground(catYellow).
				Bold(true)

	countGreenStyle = lipgloss.NewStyle().
			Foreground(catGreen).
			Bold(true)

	dimStyle = lipgloss.NewStyle().
			Foreground(catOverlay0)
)

// ── Remediation view ──────────────────────────────────────────────────────────

// Legacy aliases used in view.go remediation helpers.
var (
	colorAmber  = catYellow
	colorOrange = catPeach
	colorViolet = catMauve
	colorBg     = catBase
)

var (
	remHeaderTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(catBase).
				Background(catBlue).
				Padding(0, 2)

	remSepStyle = lipgloss.NewStyle().
			Foreground(catSurface1)

	remIssueLabelStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(catSapphire)

	remIssueValueStyle = lipgloss.NewStyle().
				Foreground(catText)

	remFixTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(catYellow).
				Padding(0, 1)

	remFixDescStyle = lipgloss.NewStyle().
			Foreground(catSubtext0).
			Padding(0, 2)

	remCodeBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(catBlue).
				Padding(0, 1).
				Background(catBase)

	remCodeFileStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(catTeal).
				Background(catBase).
				Padding(0, 1)

	remCodeSepStyle = lipgloss.NewStyle().
			Foreground(catSurface1).
			Background(catBase)

	remScrollHintStyle = lipgloss.NewStyle().
				Foreground(catOverlay0).
				Background(catBase).
				Italic(true)

	// HCL syntax highlighting.
	remHCLKeywordStyle   = lipgloss.NewStyle().Bold(true).Foreground(catMauve)
	remHCLBlockArgStyle  = lipgloss.NewStyle().Foreground(catYellow)
	remHCLAttrStyle      = lipgloss.NewStyle().Foreground(catSapphire)
	remHCLEqualsStyle    = lipgloss.NewStyle().Foreground(catOverlay1)
	remHCLStringStyle    = lipgloss.NewStyle().Foreground(catGreen)
	remHCLBoolStyle      = lipgloss.NewStyle().Foreground(catPeach)
	remHCLRefStyle       = lipgloss.NewStyle().Foreground(catBlue)
	remHCLBraceStyle     = lipgloss.NewStyle().Foreground(catOverlay2)
	remHCLInnerBlockStyle = lipgloss.NewStyle().Foreground(catYellow)
	remCommentStyle      = lipgloss.NewStyle().Foreground(catOverlay0).Italic(true)
	remDefaultLineStyle  = lipgloss.NewStyle().Foreground(catText)

	remFooterStyle = lipgloss.NewStyle().
			Foreground(catOverlay1).
			Padding(0, 1)
)
