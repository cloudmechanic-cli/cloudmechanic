package tui

import (
	"context"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/cloudmechanic-cli/cloudmechanic/internal/scanner"
)

// --- Messages ---

// scanResultMsg is sent when the background scan completes.
type scanResultMsg struct {
	issues  []scanner.Issue
	errors  []error
	elapsed time.Duration
}

// scanStartMsg triggers a new scan.
type scanStartMsg struct{}

// --- Model ---

// ScannerBuilder abstracts how scanners + regions are created so the TUI
// doesn't import cmd directly (avoids circular deps).
type ScannerBuilder struct {
	Cfg        aws.Config
	Regions    []string
	AllRegions bool
	Build      func(cfg aws.Config, regions []string) []scanner.Scanner
	Run        func(ctx context.Context, scanners []scanner.Scanner) ([]scanner.Issue, []error)
}

type state int

const (
	stateLoading state = iota
	stateReady
)

// Model is the top-level Bubble Tea model.
type Model struct {
	state    state
	builder  ScannerBuilder
	spinner  spinner.Model
	issues   []scanner.Issue
	errors   []error
	elapsed  time.Duration
	cursor   int
	expanded bool // true when the selected issue is expanded
	width    int
	height   int
}

// NewModel creates a Model ready to be run.
func NewModel(b ScannerBuilder) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle

	return Model{
		state:   stateLoading,
		builder: b,
		spinner: s,
	}
}

// Init starts the spinner and fires the first scan.
func (m Model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, startScan())
}

func startScan() tea.Cmd {
	return func() tea.Msg { return scanStartMsg{} }
}

// runScanCmd performs the scan in a goroutine and sends the result back.
func (m Model) runScanCmd() tea.Cmd {
	b := m.builder
	return func() tea.Msg {
		ctx := context.Background()
		start := time.Now()

		scanners := b.Build(b.Cfg, b.Regions)
		issues, errs := b.Run(ctx, scanners)

		// Sort: criticals first, then warnings.
		sort.Slice(issues, func(i, j int) bool {
			return issues[i].Severity > issues[j].Severity
		})

		return scanResultMsg{
			issues:  issues,
			errors:  errs,
			elapsed: time.Since(start),
		}
	}
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case scanStartMsg:
		m.state = stateLoading
		m.issues = nil
		m.errors = nil
		m.cursor = 0
		m.expanded = false
		return m, tea.Batch(m.spinner.Tick, m.runScanCmd())

	case scanResultMsg:
		m.state = stateReady
		m.issues = msg.issues
		m.errors = msg.errors
		m.elapsed = msg.elapsed
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "r":
		if m.state == stateReady {
			return m, startScan()
		}
	case "up", "k":
		if m.state == stateReady && m.cursor > 0 {
			m.cursor--
			m.expanded = false
		}
	case "down", "j":
		if m.state == stateReady && m.cursor < len(m.issues)-1 {
			m.cursor++
			m.expanded = false
		}
	case "enter":
		if m.state == stateReady && len(m.issues) > 0 {
			m.expanded = !m.expanded
		}
	case "esc":
		m.expanded = false
	}
	return m, nil
}
