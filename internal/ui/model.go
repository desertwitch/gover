package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/desertwitch/gover/internal/queue"
	"github.com/dustin/go-humanize"
)

//nolint:gochecknoglobals
var (
	// titleStyle defines the style for a panel's title.
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4"))

	// borderStyle defines the style for a panel's borders.
	borderStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4"))

	// infoStyle defines the style for a panel's text.
	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA"))

	// helpStyle defines the style for the help panel's text.
	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Padding(0, 1)
)

// QueueProgressMsg is a [tea.Msg] containing [queue.Progress] information.
type QueueProgressMsg struct {
	t               time.Time
	enumerationData queue.Progress
	evaluationData  queue.Progress
	ioData          queue.Progress
}

// TeaModel is the principal [tea.Model] for the command-line user interface.
type TeaModel struct {
	width  int
	height int

	cancel context.CancelFunc

	uiHandler    *Handler
	queueManager *queue.Manager

	fullWidthWithBorders  int
	splitWidthWithBorders int

	enumerationData queue.Progress
	evaluationData  queue.Progress
	ioData          queue.Progress

	enumerationProgress progress.Model
	evaluationProgress  progress.Model
	ioProgress          progress.Model
	logsViewport        viewport.Model
	logs                []string

	ready bool
}

// NewTeaModel returns an initial new [TeaModel].
//
//nolint:mnd
func NewTeaModel(uiHandler *Handler, queueManager *queue.Manager, cancel context.CancelFunc) TeaModel {
	enumerationProgress := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(80),
	)
	evaluationProgress := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(80),
	)
	ioProgress := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(80),
	)

	logsViewport := viewport.New(80, 20)

	return TeaModel{
		uiHandler:           uiHandler,
		queueManager:        queueManager,
		enumerationProgress: enumerationProgress,
		evaluationProgress:  evaluationProgress,
		ioProgress:          ioProgress,
		enumerationData:     queue.Progress{},
		evaluationData:      queue.Progress{},
		ioData:              queue.Progress{},
		logsViewport:        logsViewport,
		logs:                make([]string, 0, 100),
		cancel:              cancel,
		ready:               false,
	}
}

// Init initializes the model within a [tea.Program].
func (m TeaModel) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		updateQueueProgress(m.queueManager),
	)
}

// updateQueueProgress produces a [tea.Cmd] for later scheduling in a
// [tea.Program]. When executed, a [QueueProgressMsg] with a [queue.Manager]'s
// [queue.Progress] is returned.
func updateQueueProgress(q *queue.Manager) tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg { //nolint:mnd
		queueProgressMsg := QueueProgressMsg{
			t:               t,
			enumerationData: q.EnumerationManager.Progress(),
			evaluationData:  q.EvaluationManager.Progress(),
			ioData:          q.IOManager.Progress(),
		}

		return queueProgressMsg
	})
}

// Update is the principal message handling method of the model.
// It sets the internal state of the model, for later rendering.
//
//nolint:mnd,funlen,ireturn
func (m TeaModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.cancel()

			return m, tea.Quit
		case "q":
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		m.fullWidthWithBorders = m.width - 2
		m.splitWidthWithBorders = (m.width / 3) - 2

		// Progress bars should match the content width.
		m.enumerationProgress.Width = m.splitWidthWithBorders
		m.evaluationProgress.Width = m.splitWidthWithBorders
		m.ioProgress.Width = m.splitWidthWithBorders

		// We want upper panels to take about 40% of the height.
		upperHeight := m.height * 2 / 5
		lowerHeight := m.height - upperHeight

		// Viewport height: lower section minus borders and title.
		viewportHeight := lowerHeight - 3

		// Set viewport width to full width minus borders.
		m.logsViewport.Width = m.fullWidthWithBorders
		m.logsViewport.Height = viewportHeight

		// Update viewport content with current logs.
		if len(m.logs) > 0 {
			logs := lipgloss.NewStyle().
				Width(m.logsViewport.Width).
				Render(strings.TrimSuffix(strings.Join(m.logs, ""), "\n"))

			m.logsViewport.SetContent(logs)
			m.logsViewport.GotoBottom()
		}

		if !m.ready {
			m.ready = true
			m.uiHandler.Ready.Store(true)
		}

	case QueueProgressMsg:
		m.enumerationData = msg.enumerationData
		m.evaluationData = msg.evaluationData
		m.ioData = msg.ioData

		cmds = append(cmds,
			m.enumerationProgress.SetPercent(float64(m.enumerationData.ProgressPct)/100),
			m.evaluationProgress.SetPercent(float64(m.evaluationData.ProgressPct)/100),
			m.ioProgress.SetPercent(float64(m.ioData.ProgressPct)/100),
		)

		// Queue the next update.
		cmds = append(cmds, updateQueueProgress(m.queueManager))

	case LogMsg:
		logMsg := string(msg)

		if len(m.logs) >= 100 {
			m.logs = m.logs[1:]
		}

		m.logs = append(m.logs, logMsg)

		logs := lipgloss.NewStyle().
			Width(m.logsViewport.Width).
			Render(strings.TrimSuffix(strings.Join(m.logs, ""), "\n"))

		m.logsViewport.SetContent(logs)
		m.logsViewport.GotoBottom()

	case progress.FrameMsg:
		// Update enumeration progress.
		updatedEnum, cmd := m.enumerationProgress.Update(msg)
		if progressModel, ok := updatedEnum.(progress.Model); ok {
			m.enumerationProgress = progressModel
		}
		cmds = append(cmds, cmd)

		// Update evaluation progress.
		updatedEval, cmd := m.evaluationProgress.Update(msg)
		if progressModel, ok := updatedEval.(progress.Model); ok {
			m.evaluationProgress = progressModel
		}
		cmds = append(cmds, cmd)

		// Update IO progress.
		updatedIO, cmd := m.ioProgress.Update(msg)
		if progressModel, ok := updatedIO.(progress.Model); ok {
			m.ioProgress = progressModel
		}
		cmds = append(cmds, cmd)
	}

	// Handle viewport updates.
	m.logsViewport, cmd = m.logsViewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// View is the principal rendering function of the model.
func (m TeaModel) View() string {
	if !m.ready {
		return "Loading the GUI..."
	}

	var s strings.Builder

	enumerationView := m.formatProgressView("Enumeration", m.enumerationProgress.View(), m.enumerationData)
	evaluationView := m.formatProgressView("Evaluation", m.evaluationProgress.View(), m.evaluationData)
	ioView := m.formatProgressView("IO", m.ioProgress.View(), m.ioData)

	progressSection := lipgloss.JoinHorizontal(
		lipgloss.Top,
		borderStyle.Width(m.splitWidthWithBorders).Render(enumerationView),
		borderStyle.Width(m.splitWidthWithBorders).Render(evaluationView),
		borderStyle.Width(m.splitWidthWithBorders).Render(ioView),
	)

	logsSection := borderStyle.
		Width(m.fullWidthWithBorders).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				titleStyle.Width(m.fullWidthWithBorders).Render("Process Information"),
				lipgloss.NewStyle().Width(m.fullWidthWithBorders).Render(m.logsViewport.View()),
			),
		)

	helpSection := helpStyle.
		Width(m.fullWidthWithBorders).
		Render("q: quit gui • ctrl+c: quit program")

	s.WriteString(lipgloss.JoinVertical(
		lipgloss.Left,
		progressSection,
		logsSection,
		helpSection,
	))

	return s.String()
}

// formatProgressView is a helper function for rendering the progress panels.
func (m TeaModel) formatProgressView(title string, progressBar string, progress queue.Progress) string {
	var timeLeft time.Duration
	var timeLeftMin float64

	if !progress.ETA.IsZero() {
		timeLeft = time.Until(progress.ETA)
		timeLeftMin = float64(timeLeft.Minutes())
	}

	var transferSpeed string
	if progress.TransferSpeedUnit == "bytes/sec" {
		transferSpeed = humanize.Bytes(uint64(progress.TransferSpeed)) + "/s"
	} else {
		transferSpeed = fmt.Sprintf("%d %s", int(progress.TransferSpeed), progress.TransferSpeedUnit)
	}

	var details string
	if !progress.HasFinished {
		details = fmt.Sprintf(
			"Progress: %.2f%% (%d/%d)\n"+
				"Items: InProgress=%d, Success=%d, Skipped=%d\n"+
				"Time: Started=%v, ETA=%v (%.1f%s left)\n"+
				"Speed: %s\n",
			progress.ProgressPct,
			progress.ProcessedItems,
			progress.TotalItems,
			progress.InProgressItems,
			progress.SuccessItems,
			progress.SkippedItems,
			progress.StartTime.Format("15:04:05"),
			progress.ETA.Format("15:04:05"),
			timeLeftMin, "min",
			transferSpeed,
		)
	} else {
		details = fmt.Sprintf(
			"Progress: %.2f%% (%d/%d)\n"+
				"Items: InProgress=%d, Success=%d, Skipped=%d\n"+
				"Time: Started=%v, Finished=%v\n\n",
			progress.ProgressPct,
			progress.ProcessedItems,
			progress.TotalItems,
			progress.InProgressItems,
			progress.SuccessItems,
			progress.SkippedItems,
			progress.StartTime.Format("15:04:05"),
			progress.FinishTime.Format("15:04:05"),
		)
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		titleStyle.Width(m.splitWidthWithBorders).Render(title),
		"", // Empty line for spacing.
		progressBar,
		"", // Empty line for spacing.
		infoStyle.Width(m.splitWidthWithBorders).Render(details),
	)

	return content
}
