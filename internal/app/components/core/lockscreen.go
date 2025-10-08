package core

import (
	"bytes"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/charmbracelet/log"
)

type ProgressUpdateMsg struct {
	Percent float64
	LogLine string
}

type LockScreenModel struct {
	progress  *ProgressModel
	logger    *log.Logger
	logOutput *bytes.Buffer
	width     int
	height    int
}

func NewLockScreen(width, height int) *LockScreenModel {
	logOutput := new(bytes.Buffer)
	logger := log.New(logOutput)
	logger.SetLevel(log.DebugLevel)

	return &LockScreenModel{
		progress:  NewProgress(WithDefaultGradient()),
		logger:    logger,
		logOutput: logOutput,
		width:     width,
		height:    height,
	}
}

func (m *LockScreenModel) Init() tea.Cmd {
	return nil
}

func (m *LockScreenModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case ProgressUpdateMsg:
		if msg.LogLine != "" {
			m.logger.Info(msg.LogLine)
		}
		cmd = m.progress.SetPercent(msg.Percent)
		cmds = append(cmds, cmd)
	case FrameMsg:
		var newProgress tea.Model
		newProgress, cmd = m.progress.Update(msg)
		if newProgress != nil {
			m.progress = newProgress.(*ProgressModel)
		}
		cmds = append(cmds, cmd)
	default:
		// Also forward other messages to the progress bar
		var newProgress tea.Model
		newProgress, cmd = m.progress.Update(msg)
		if newProgress != nil {
			m.progress = newProgress.(*ProgressModel)
		}
		cmds = append(cmds, cmd)

	}

	return m, tea.Batch(cmds...)
}

func (m *LockScreenModel) View() string {
	progressView := m.progress.View()
	logsView := m.logOutput.String()

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		lipgloss.JoinVertical(lipgloss.Center, progressView, logsView),
	)
}
