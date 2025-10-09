package core

import (
	"bytes"
	"fmt"

	"github.com/charmbracelet/bubbles/v2/progress"
	"github.com/charmbracelet/bubbles/v2/spinner"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/charmbracelet/log"
)

type ProgressUpdateMsg struct {
	Percent float64
	LogLine string
}

type LockScreenModel struct {
	progress     progress.Model
	spinner      spinner.Model
	logger       *log.Logger
	logOutput    *bytes.Buffer
	width        int
	height       int
	debugPercent float64
}

func NewLockScreen(width, height int) *LockScreenModel {
	logOutput := new(bytes.Buffer)
	logger := log.NewWithOptions(logOutput, log.Options{
		Formatter: log.TextFormatter,
	})
	logger.SetLevel(log.DebugLevel)

	pr := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(width/2),
	)

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return &LockScreenModel{
		progress:  pr,
		spinner:   s,
		logger:    logger,
		logOutput: logOutput,
		width:     width,
		height:    height,
	}
}

func (m *LockScreenModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m *LockScreenModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.progress = progress.New(
			progress.WithDefaultGradient(),
			progress.WithWidth(msg.Width/2),
		)

	case ProgressUpdateMsg:
		m.debugPercent = msg.Percent
		if msg.LogLine != "" {
			m.logger.Info(msg.LogLine)
		}
		cmd = m.progress.SetPercent(msg.Percent)
		cmds = append(cmds, cmd)

	case progress.FrameMsg:
		newProgress, newCmd := m.progress.Update(msg)
		m.progress = newProgress
		cmd = newCmd
		cmds = append(cmds, cmd)

	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *LockScreenModel) View() string {
	progressView := m.progress.View()
	spinnerView := m.spinner.View()
	percentView := fmt.Sprintf("Received Percent: %.2f", m.debugPercent)
	logsView := m.logOutput.String()

	// Style for the logs container
	logsStyle := lipgloss.NewStyle().
		Width(m.width/2).
		MaxHeight(m.height/2).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1, 2)

	spinnerAndPercent := lipgloss.JoinHorizontal(lipgloss.Center, spinnerView, " ", percentView)
	ui := lipgloss.JoinVertical(
		lipgloss.Center,
		spinnerAndPercent,
		progressView,
		logsStyle.Render(logsView),
	)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		ui,
	)
}
