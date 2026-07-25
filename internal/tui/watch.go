package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Morpa/wick/internal/format"
	"github.com/Morpa/wick/internal/session"
)

// Styles
var (
	statPanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("14")).
			PaddingLeft(1).
			PaddingRight(1)

	topPanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("5")).
			PaddingLeft(1).
			PaddingRight(1)

	skillPanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("4")).
			PaddingLeft(1).
			PaddingRight(1)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("14"))

	totalStyle = lipgloss.NewStyle().Bold(true)

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))

	greenStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	yellowStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	redStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	warnStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))

	headerTop = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("5"))
	headerSkl = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("4"))

	errStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("11")).
			PaddingLeft(1).
			PaddingRight(1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Padding(0, 1)
)

// Messages
type tickMsg struct{}

// Model
type model struct {
	cwd       string
	viewModel *format.ViewModel
	loaded    bool
	width     int
	height    int
	err       string
	viewport  viewport.Model
}

// NewModel creates the initial Bubble Tea model for the watch dashboard.
func NewModel(cwd string) model {
	return model{
		cwd: cwd,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Tick(0, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}

func (m model) loadViewModel() *format.ViewModel {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	encoded := session.EncodeProjectDir(m.cwd)
	projectDir := filepath.Join(home, ".claude", "projects", encoded)
	sessionFile, ok := session.FindActiveSession(projectDir)
	if !ok {
		return nil
	}
	events, warnings := session.ParseSession(sessionFile)
	turns := session.GroupIntoTurns(events)
	totals := session.ComputeTotals(events, turns)
	sessionID := filepath.Base(sessionFile)
	// strip .jsonl extension
	sessionID = sessionID[:len(sessionID)-6]
	vm := format.BuildViewModel(totals, sessionID, filepath.Base(m.cwd), warnings)
	return &vm
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 1 // leave room for help line
		return m, nil

	case tickMsg:
		m.loaded = true
		m.viewModel = m.loadViewModel()
		return m, tea.Tick(2*time.Second, func(time.Time) tea.Msg {
			return tickMsg{}
		})

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			m.viewport.LineUp(1)
			return m, nil
		case "down", "j":
			m.viewport.LineDown(1)
			return m, nil
		case "pgup":
			m.viewport.HalfViewUp()
			return m, nil
		case "pgdown", " ":
			m.viewport.HalfViewDown()
			return m, nil
		case "home":
			m.viewport.GotoTop()
			return m, nil
		case "end":
			m.viewport.GotoBottom()
			return m, nil
		}

	case tea.MouseMsg:
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonWheelUp {
			m.viewport.LineUp(3)
			return m, nil
		}
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonWheelDown {
			m.viewport.LineDown(3)
			return m, nil
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.viewModel == nil {
		if !m.loaded {
			return errStyle.Render(
				lipgloss.JoinVertical(lipgloss.Left,
					warnStyle.Render("⏳ wick"),
					"",
					dimStyle.Render("Loading…"),
					dimStyle.Render("Looking for active Claude Code session."),
				),
			) + "\n"
		}
		return errStyle.Render(
			lipgloss.JoinVertical(lipgloss.Left,
				warnStyle.Render("⚠ wick"),
				"",
				dimStyle.Render("No active session found."),
				dimStyle.Render("Run this command from within a project"),
				dimStyle.Render("with an active Claude Code session."),
			),
		) + "\n"
	}

	vm := m.viewModel

	// ── Stats panel ──
	statsLines := []string{
		titleStyle.Render(vm.Header),
		"",
		totalStyle.Render("🔥 Total: " + vm.TotalTokens + " tok"),
		dimStyle.Render("   " + vm.Breakdown),
		"",
	}

	// Context bar
	if vm.ContextPercentage >= 0 {
		bar := ctxBar(vm.ContextPercentage, 10)
		ctxColor := greenStyle
		label := "🧠 Context"
		switch vm.ContextSeverity {
		case "danger":
			ctxColor = redStyle
			label = "🔥 Context"
		case "warn":
			ctxColor = yellowStyle
			label = "⚠ Context"
		}
		statsLines = append(statsLines,
			ctxColor.Render(fmt.Sprintf("%s: %.1f%% used", label, vm.ContextPercentage)),
		)
		if vm.ContextDetail != "" {
			statsLines = append(statsLines,
				dimStyle.Render(fmt.Sprintf("   %s / %s", bar, vm.ContextDetail)),
			)
		}
		statsLines = append(statsLines,
			dimStyle.Render(fmt.Sprintf("     %s", vm.ContextDetail)),
		)
	} else {
		statsLines = append(statsLines, dimStyle.Render("🧠 Context: no data available yet"))
	}
	statsLines = append(statsLines, "")

	if len(vm.TopRows) > 0 {
		statsLines = append(statsLines,
			dimStyle.Render(fmt.Sprintf("🏆 Top: \"%s\"", vm.TopRows[0].Preview)),
		)
	} else {
		statsLines = append(statsLines, dimStyle.Render("🏆 Top: (no prompts have been recorded yet)"))
	}

	statsPanel := statPanelStyle.Width(m.width - 4).Render(lipgloss.JoinVertical(lipgloss.Left, statsLines...))

	// ── Top prompts ──
	var topLines []string
	topLines = append(topLines, headerTop.Render("🏆 Top prompts"))
	if len(vm.TopRows) > 0 {
		for _, row := range vm.TopRows {
			preview := truncate(row.Preview, 45)
			topLines = append(topLines,
				fmt.Sprintf(" %d. %7s  \"%s\"", row.Rank, row.Tokens, preview),
			)
		}
	} else {
		topLines = append(topLines, dimStyle.Render(" (no prompts have been recorded yet)"))
	}
	topPanel := topPanelStyle.Width(m.width - 4).Render(lipgloss.JoinVertical(lipgloss.Left, topLines...))

	// ── Skill breakdown ──
	var skillLines []string
	skillLines = append(skillLines, headerSkl.Render("🧩 By skill/agent"))
	if len(vm.SkillRows) > 0 {
		for _, row := range vm.SkillRows {
			skillLines = append(skillLines,
				fmt.Sprintf(" %-12s %s", row.Skill, row.Tokens),
			)
		}
	} else {
		skillLines = append(skillLines, dimStyle.Render(" (no data)"))
	}
	skillPanel := skillPanelStyle.Width(m.width - 4).Render(lipgloss.JoinVertical(lipgloss.Left, skillLines...))

	// ── Warnings ──
	var content []string
	content = append(content, statsPanel, "", topPanel, "", skillPanel)

	if vm.WarningsLine != "" {
		content = append(content, "", warnStyle.Render("⚠ "+vm.WarningsLine))
	}

	body := lipgloss.JoinVertical(lipgloss.Left, content...)
	m.viewport.SetContent(body)

	helpLine := helpStyle.Render("↑↓ scroll  j/k  pgup/pgdn  q quit")

	return m.viewport.View() + "\n" + helpLine
}

// ctxBar renders a simple progress bar (e.g. ██████░░░░).
func ctxBar(pct float64, segs int) string {
	filled := min(int(pct/100*float64(segs)), segs)
	filled = max(filled, 0)
	empty := segs - filled
	fillCh := lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render(repeat("█", filled))
	emptCh := lipgloss.NewStyle().Foreground(lipgloss.Color("236")).Render(repeat("░", empty))

	return fillCh + emptCh
}

func repeat(s string, n int) string {
	if n <= 0 {
		return ""
	}
	var b strings.Builder
	for range n {
		b.WriteString(s)
	}
	return b.String()
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}

// RunWatch starts the Bubble Tea live dashboard.
func RunWatch(cwd string) {
	p := tea.NewProgram(NewModel(cwd), tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "wick: %v\n", err)
		os.Exit(1)
	}
}
