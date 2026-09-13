package tui

import (
	"context"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/1garo/tinycloud/internal/app"
	"github.com/1garo/tinycloud/internal/runtime"
	"github.com/1garo/tinycloud/internal/store"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type mode int

const (
	modeList mode = iota
	modeName
	modeDirectory
	modePort
)

type item app.App

func (i item) FilterValue() string { return i.Name }
func (i item) Title() string       { return i.Name }
func (i item) Description() string { return string(i.Status) + " · " + app.App(i).URL() }

type model struct {
	store    *store.Store
	runtime  runtime.Runtime
	list     list.Model
	input    textinput.Model
	help     help.Model
	mode     mode
	newApp   app.App
	message  string
	busy     bool
	showLogs bool
	logs     string
	width    int
	height   int
}

type operationFinished struct {
	app      app.App
	logs     string
	showLogs bool
	err      error
}

func NewModel(appStore *store.Store, appRuntime runtime.Runtime) tea.Model {
	items := makeItems(appStore.List())
	apps := list.New(items, list.NewDefaultDelegate(), 80, 20)
	apps.Title = "TinyCloud"
	apps.AdditionalFullHelpKeys = func() []key.Binding { return []key.Binding{} }

	input := textinput.New()
	input.Prompt = "  > "
	input.CharLimit = 160
	input.Width = 60
	input.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	input.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
	input.PlaceholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	input.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return &model{
		store:   appStore,
		runtime: appRuntime,
		list:    apps,
		input:   input,
		help:    help.New(),
	}
}

func (m *model) Init() tea.Cmd {
	return nil
}

func (m *model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch message := message.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = message.Width, message.Height
		m.list.SetSize(message.Width, message.Height-4)
	case operationFinished:
		m.busy = false
		if message.err != nil {
			m.message = message.err.Error()
			if err := m.store.Update(message.app); err != nil {
				m.message = err.Error()
			}
			m.refresh()
			return m, nil
		}
		m.message = "Deployment finished"
		m.logs = message.logs
		m.showLogs = message.showLogs
		if err := m.store.Update(message.app); err != nil {
			m.message = err.Error()
		}
		m.refresh()
	}

	if m.showLogs {
		keyMsg, ok := message.(tea.KeyMsg)
		if ok && key.Matches(keyMsg, key.NewBinding(key.WithKeys("esc", "q"))) {
			m.showLogs = false
			return m, nil
		}
		return m, nil
	}

	if m.mode != modeList {
		return m.updateInput(message)
	}
	if m.busy {
		return m, nil
	}
	if m.list.FilterState() == list.Filtering {
		var command tea.Cmd
		m.list, command = m.list.Update(message)
		return m, command
	}
	if keyMsg, ok := message.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "n":
			m.startInput(modeName, "App name")
			return m, nil
		case "d":
			return m, m.deploySelected()
		case "s":
			return m, m.stopSelected()
		case "l":
			return m, m.loadLogs()
		case "enter":
			selected, ok := m.selected()
			if ok {
				m.message = selected.URL()
			}
		}
	}
	var command tea.Cmd
	m.list, command = m.list.Update(message)
	return m, command
}

func (m *model) View() string {
	if m.showLogs {
		return lipgloss.NewStyle().Padding(1, 2).Render("Logs\n\n" + m.logs + "\n\nPress esc to return")
	}
	if m.mode != modeList {
		return lipgloss.NewStyle().Padding(2, 4).Render(
			lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Render(inputTitle(m.mode)) +
				"\n\n" +
				m.input.View() +
				"\n\nenter: continue · esc: cancel",
		)
	}
	footer := "n: new · d: deploy · l: logs · s: stop · enter: show URL · q: quit"
	if m.busy {
		footer = "Working..."
	}
	if m.message != "" {
		footer += "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render(m.message)
	}
	return m.list.View() + "\n" + footer
}

func (m *model) updateInput(message tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := message.(tea.KeyMsg); ok && keyMsg.String() == "esc" {
		m.mode = modeList
		m.input.Reset()
		return m, nil
	}
	if keyMsg, ok := message.(tea.KeyMsg); ok && keyMsg.String() == "enter" {
		return m.advanceInput()
	}
	var command tea.Cmd
	m.input, command = m.input.Update(message)
	return m, command
}

func (m *model) advanceInput() (tea.Model, tea.Cmd) {
	value := strings.TrimSpace(m.input.Value())
	if value == "" {
		return m, nil
	}
	switch m.mode {
	case modeName:
		name := slug(value)
		if name == "" {
			m.message = "Use letters, numbers, and hyphens for the app name"
			return m, nil
		}
		m.newApp = app.App{Name: name, ID: name, Status: app.StatusCreated}
		m.startInput(modeDirectory, "Project directory")
	case modeDirectory:
		m.newApp.SourceDir = filepath.Clean(value)
		m.startInput(modePort, "Container port")
	case modePort:
		port, err := strconv.Atoi(value)
		if err != nil || port < 1 || port > 65535 {
			m.message = "Enter a valid port between 1 and 65535"
			return m, nil
		}
		m.newApp.ContainerPort = value
		m.newApp.HostPort = value
		m.newApp.CreatedAt = time.Now()
		m.newApp.UpdatedAt = m.newApp.CreatedAt
		if err := m.store.Add(m.newApp); err != nil {
			m.message = err.Error()
		} else {
			m.message = "App created; press d to deploy"
		}
		m.mode = modeList
		m.input.Reset()
		m.refresh()
	}
	return m, nil
}

func (m *model) startInput(next mode, placeholder string) {
	m.mode = next
	m.input.Reset()
	m.input.Placeholder = placeholder
	_ = m.input.Focus()
}

func inputTitle(current mode) string {
	switch current {
	case modeName:
		return "Create app · enter a name"
	case modeDirectory:
		return "Create app · enter the project directory"
	case modePort:
		return "Create app · enter the container port"
	default:
		return "Create app"
	}
}

func (m *model) deploySelected() tea.Cmd {
	selected, ok := m.selected()
	if !ok {
		m.message = "Create an app first with n"
		return nil
	}
	m.busy = true
	selected.Status = app.StatusBuilding
	selected.UpdatedAt = time.Now()
	_ = m.store.Update(selected)
	m.refresh()
	return func() tea.Msg {
		containerID, err := m.runtime.Deploy(context.Background(), selected.SourceDir, "tinycloud/"+selected.Name, "tinycloud-"+selected.Name, selected.ContainerPort)
		if err != nil {
			selected.Status = app.StatusFailed
			selected.LastError = err.Error()
			return operationFinished{app: selected, err: err}
		}
		selected.ContainerID = containerID
		selected.Status = app.StatusRunning
		selected.LastError = ""
		selected.UpdatedAt = time.Now()
		return operationFinished{app: selected}
	}
}

func (m *model) stopSelected() tea.Cmd {
	selected, ok := m.selected()
	if !ok {
		return nil
	}
	m.busy = true
	return func() tea.Msg {
		err := m.runtime.Stop(context.Background(), "tinycloud-"+selected.Name)
		selected.Status = app.StatusStopped
		selected.UpdatedAt = time.Now()
		return operationFinished{app: selected, err: err}
	}
}

func (m *model) loadLogs() tea.Cmd {
	selected, ok := m.selected()
	if !ok {
		return nil
	}
	m.busy = true
	return func() tea.Msg {
		logs, err := m.runtime.Logs(context.Background(), "tinycloud-"+selected.Name)
		return operationFinished{app: selected, logs: logs, showLogs: err == nil, err: err}
	}
}

func (m *model) selected() (app.App, bool) {
	selected := m.list.SelectedItem()
	if selected == nil {
		return app.App{}, false
	}
	return app.App(selected.(item)), true
}

func (m *model) refresh() {
	items := makeItems(m.store.List())
	m.list.SetItems(items)
}

func makeItems(apps []app.App) []list.Item {
	items := make([]list.Item, 0, len(apps))
	for _, current := range apps {
		items = append(items, item(current))
	}
	return items
}

var appNamePattern = regexp.MustCompile(`[^a-z0-9-]+`)

func slug(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = appNamePattern.ReplaceAllString(value, "-")
	return strings.Trim(value, "-")
}

var _ tea.Model = (*model)(nil)
