package core

import (
	"bytes"

	"catalyst/internal/app/styles"

	"github.com/charmbracelet/bubbles/v2/progress"
	"github.com/charmbracelet/bubbles/v2/spinner"
	"github.com/charmbracelet/bubbles/v2/viewport"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/charmbracelet/log"
	"github.com/muesli/termenv"
)

type ProgressUpdateMsg struct {
	Percent float64
	LogLine string
}

type LockScreenModel struct {
	progress   progress.Model
	spinner    spinner.Model
	viewport   viewport.Model
	logger     *log.Logger
	logOutput  *bytes.Buffer
	width      int
	height     int
	ActionText string
	theme      *styles.Theme
}

func NewLockScreen(
	width, availableHeight int,
	actionText string,
	theme *styles.Theme,
) *LockScreenModel {
	logOutput := new(bytes.Buffer)
	logger := log.New(logOutput)
	logger.SetColorProfile(termenv.TrueColor)
	logger.SetLevel(log.DebugLevel)

	pr := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(width/2),
	)

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(theme.Accent)

	// spinner(1) + progress(1) + action text(1) + margins(2) = 5
	viewportHeight := availableHeight - 5
	vp := viewport.New()
	vp.SetHeight(viewportHeight)
	vp.SetWidth(width / 2)

	return &LockScreenModel{
		progress:   pr,
		spinner:    s,
		viewport:   vp,
		logger:     logger,
		logOutput:  logOutput,
		width:      width,
		height:     availableHeight,
		ActionText: actionText,
		theme:      theme,
	}
}

func (m *LockScreenModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m *LockScreenModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case ProgressUpdateMsg:
		if msg.LogLine != "" {
			m.logger.Info(msg.LogLine)
			m.viewport.SetContent(m.logOutput.String())
			m.viewport.GotoBottom()
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

	// Also handle viewport scrolling
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *LockScreenModel) View() string {
	progressView := m.progress.View()
	spinnerView := m.spinner.View()
	actionView := lipgloss.NewStyle().Bold(true).Render(m.ActionText)
	VerticalSpace := lipgloss.NewStyle().Height(1).Render("")

	// Style for the logs container is now applied to the viewport
	m.viewport.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.Accent).
		Padding(0, 1)

	spinnerAndAction := lipgloss.JoinHorizontal(lipgloss.Center, spinnerView, " ", actionView)
	ui := lipgloss.JoinVertical(
		lipgloss.Center,
		spinnerAndAction,
		VerticalSpace,
		progressView,
		m.viewport.View(),
	)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		ui,
	)
}

func (m *LockScreenModel) Resize(width, availableHeight int) {
	m.width = width
	m.height = availableHeight
	m.progress.SetWidth(width / 2)
	m.viewport.SetWidth(width / 2)
	m.viewport.SetHeight(availableHeight - 5)
}
