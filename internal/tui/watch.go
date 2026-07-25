package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

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
	cwd            string
	viewModel      *format.ViewModel
	loaded         bool
	width          int
	height         int
	err            string
	scrollY        int
	scrollX        int
	contentLines   []string // cached split lines for scrolling (set in Update, read in View)
	maxLineLen     int      // longest line width for horizontal scroll
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
		m.scrollY = 0
		m.scrollX = 0
		return m, nil

	case tickMsg:
		m.loaded = true
		m.viewModel = m.loadViewModel()
		if m.viewModel != nil {
			m.rebuildContent()
		}
		m.scrollY = 0
		m.scrollX = 0
		return m, tea.Tick(2*time.Second, func(time.Time) tea.Msg {
			return tickMsg{}
		})

	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		visibleLines := m.height - 1 // -1 for help line

		// max vertical scroll
		maxY := len(m.contentLines) - visibleLines
		if maxY < 0 {
			maxY = 0
		}
		// max horizontal scroll (half terminal width by default)
		contentWidth := m.width - 4 // account for borders
		maxX := m.maxLineLen - contentWidth
		if maxX < 0 {
			maxX = 0
		}

		switch msg.String() {
		case "up", "k":
			if m.scrollY > 0 {
				m.scrollY--
			}
		case "down", "j":
			if m.scrollY < maxY {
				m.scrollY++
			}
		case "left", "h":
			if m.scrollX > 0 {
				m.scrollX -= 8
				if m.scrollX < 0 {
					m.scrollX = 0
				}
			}
		case "right", "l":
			if m.scrollX < maxX {
				m.scrollX += 8
				if m.scrollX > maxX {
					m.scrollX = maxX
				}
			}
		case "pgup":
			m.scrollY -= visibleLines
			if m.scrollY < 0 {
				m.scrollY = 0
			}
		case "pgdown", " ":
			m.scrollY += visibleLines
			if m.scrollY > maxY {
				m.scrollY = maxY
			}
		case "home":
			m.scrollY = 0
		case "end":
			m.scrollY = maxY
		}

	case tea.MouseMsg:
		if msg.Action == tea.MouseActionPress {
			visibleLines := m.height - 1
			maxY := len(m.contentLines) - visibleLines
			if maxY < 0 {
				maxY = 0
			}

			switch msg.Button {
			case tea.MouseButtonWheelUp:
				if msg.Shift {
					// horizontal scroll (shift + wheel)
					if m.scrollX > 0 {
						m.scrollX -= 8
						if m.scrollX < 0 {
							m.scrollX = 0
						}
					}
				} else {
					if m.scrollY > 0 {
						m.scrollY -= 3
						if m.scrollY < 0 {
							m.scrollY = 0
						}
					}
				}
			case tea.MouseButtonWheelDown:
				if msg.Shift {
					contentWidth := m.width - 4
					maxX := m.maxLineLen - contentWidth
					if maxX < 0 {
						maxX = 0
					}
					m.scrollX += 8
					if m.scrollX > maxX {
						m.scrollX = maxX
					}
				} else {
					m.scrollY += 3
					if m.scrollY > maxY {
						m.scrollY = maxY
					}
				}
			}
		}
	}
	return m, nil
}

// rebuildContent builds the rendered content from the view model and caches it
// for scrolling. Called from Update, where model changes persist.
func (m *model) rebuildContent() {
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

	// ── Build full content ──
	var content []string
	content = append(content, statsPanel, "", topPanel, "", skillPanel)

	if vm.WarningsLine != "" {
		content = append(content, "", warnStyle.Render("⚠ "+vm.WarningsLine))
	}

	fullContent := lipgloss.JoinVertical(lipgloss.Left, content...)

	m.contentLines = strings.Split(fullContent, "\n")
	m.maxLineLen = 0
	for _, line := range m.contentLines {
		w := ansi.StringWidth(line)
		if w > m.maxLineLen {
			m.maxLineLen = w
		}
	}
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

	// Clamp scroll offsets
	visibleLines := m.height - 1
	maxY := len(m.contentLines) - visibleLines
	if maxY < 0 {
		maxY = 0
	}
	if m.scrollY > maxY {
		m.scrollY = maxY
	}
	if m.scrollY < 0 {
		m.scrollY = 0
	}
	contentWidth := m.width - 4
	maxX := m.maxLineLen - contentWidth
	if maxX < 0 {
		maxX = 0
	}
	if m.scrollX > maxX {
		m.scrollX = maxX
	}
	if m.scrollX < 0 {
		m.scrollX = 0
	}

	// Show only the visible portion (vertical)
	top := m.scrollY
	bottom := top + visibleLines
	if bottom > len(m.contentLines) {
		bottom = len(m.contentLines)
	}

	visibleLinesSlice := m.contentLines[top:bottom]

	// Apply horizontal scroll
	var trimmed []string
	for _, line := range visibleLinesSlice {
		w := ansi.StringWidth(line)
		if m.scrollX > 0 && m.scrollX < w {
			trimmed = append(trimmed, ansi.Truncate(ansi.Cut(line, m.scrollX, w), contentWidth, ""))
		} else if m.scrollX >= w {
			trimmed = append(trimmed, "")
		} else {
			trimmed = append(trimmed, ansi.Truncate(line, contentWidth, ""))
		}
	}

	visible := strings.Join(trimmed, "\n")
	helpLine := helpStyle.Render("↑↓ scroll  h/l horiz  pgup/pgdn  q quit")

	return visible + "\n" + helpLine
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
