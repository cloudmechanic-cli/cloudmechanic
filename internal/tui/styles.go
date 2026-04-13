package tui

import "github.com/charmbracelet/lipgloss"

// Color palette.
var (
	colorCritical = lipgloss.Color("#FF4444")
	colorWarning  = lipgloss.Color("#FFAA00")
	colorGreen    = lipgloss.Color("#00CC66")
	colorCyan     = lipgloss.Color("#00BFFF")
	colorDim      = lipgloss.Color("#666666")
	colorWhite    = lipgloss.Color("#FFFFFF")
	colorSelected = lipgloss.Color("#0F3460")
	colorSearch   = lipgloss.Color("#E040FB")
)

// Header.
var headerStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(colorCyan).
	Padding(0, 1)

// Pane borders — focused vs unfocused.
var (
	focusedBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorCyan)

	unfocusedBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorDim)
)

// Sidebar.
var (
	sidebarTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorCyan)

	regionSelectedStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorWhite).
				Background(colorSelected).
				Padding(0, 1)

	regionNormalStyle = lipgloss.NewStyle().
				Foreground(colorWhite).
				Padding(0, 1)

	regionDimStyle = lipgloss.NewStyle().
			Foreground(colorDim).
			Padding(0, 1)
)

// Issue list.
var (
	criticalBadge = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorWhite).
			Background(colorCritical).
			Padding(0, 1)

	warningBadge = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#000000")).
			Background(colorWarning).
			Padding(0, 1)

	issueNormalStyle = lipgloss.NewStyle().
				Padding(0, 1)

	issueSelectedStyle = lipgloss.NewStyle().
				Background(colorSelected).
				Foreground(colorWhite).
				Bold(true).
				Padding(0, 1)

	tableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorCyan).
				Padding(0, 1).
				Underline(true)

	groupHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorWarning).
				Padding(0, 1).
				MarginTop(1)
)

// Detail pane.
var (
	detailLabelStyle = lipgloss.NewStyle().
				Foreground(colorCyan).
				Bold(true)

	detailValueStyle = lipgloss.NewStyle().
				Foreground(colorWhite)

	suggestionStyle = lipgloss.NewStyle().
			Foreground(colorGreen).
			Italic(true)
)

// Status / footer.
var (
	statusBarStyle = lipgloss.NewStyle().
			Foreground(colorDim).
			Padding(0, 1)

	filterActiveStyle = lipgloss.NewStyle().
				Foreground(colorCyan).
				Bold(true)

	searchStyle = lipgloss.NewStyle().
			Foreground(colorSearch).
			Bold(true)
)

// Misc.
var (
	spinnerStyle = lipgloss.NewStyle().
			Foreground(colorCyan)

	errorStyle = lipgloss.NewStyle().
			Foreground(colorCritical)

	countCriticalStyle = lipgloss.NewStyle().
				Foreground(colorCritical).
				Bold(true)

	countWarningStyle = lipgloss.NewStyle().
				Foreground(colorWarning).
				Bold(true)

	countGreenStyle = lipgloss.NewStyle().
			Foreground(colorGreen).
			Bold(true)

	dimStyle = lipgloss.NewStyle().
			Foreground(colorDim)
)
