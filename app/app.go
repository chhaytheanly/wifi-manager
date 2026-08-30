package app

import (
	"fmt"

	"wifi-tui/network"
	"wifi-tui/ui"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type ConnectionResultMsg struct {
	Success bool
	SSID    string
}

type Model struct {
	Networks      []network.Network
	Cursor        int
	Status        string
	State         string
	TargetSSID    string
	PasswordInput textinput.Model
	RetryCount    int
	MaxRetries    int
	Width         int
	UsedSaved     bool
}

func NewModel() Model {
	ti := textinput.New()
	ti.Placeholder = "password"
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '•'
	ti.Focus()
	ti.CharLimit = 64
	ti.Width = 32

	return Model{
		Networks:      network.ScanNetworks(),
		Cursor:        0,
		Status:        "Ready",
		State:         "list",
		PasswordInput: ti,
		RetryCount:    0,
		MaxRetries:    3,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
	case ConnectionResultMsg:
		return m.handleConnectionResult(msg)
	}

	switch m.State {
	case "list":
		return m.updateList(msg)
	case "password":
		return m.updatePassword(msg)
	case "connecting":
		return m, nil
	default:
		return m, nil
	}
}

func (m Model) handleConnectionResult(msg ConnectionResultMsg) (tea.Model, tea.Cmd) {
	if msg.Success {
		m.Status = fmt.Sprintf("Connected to %s!", msg.SSID)
		m.State = "list"
		m.PasswordInput.Blur()
		m.RetryCount = 0
		m.Networks = network.ScanNetworks()
	} else {
		m.RetryCount++
		if m.RetryCount < m.MaxRetries {
			m.State = "password"
			m.PasswordInput.Focus()
			m.PasswordInput.Reset()
			if m.UsedSaved {
				m.Status = fmt.Sprintf("Saved failed for %s. Enter password:", m.TargetSSID)
			} else {
				m.Status = fmt.Sprintf("Wrong password (%d/%d)", m.RetryCount, m.MaxRetries)
			}
		} else {
			m.Status = fmt.Sprintf("Failed after %d attempts", m.MaxRetries)
			m.State = "list"
			m.PasswordInput.Blur()
			m.RetryCount = 0
		}
	}
	return m, nil
}

func (m Model) updateList(msg tea.Msg) (tea.Model, tea.Cmd) {
	if kmsg, ok := msg.(tea.KeyMsg); ok {
		switch kmsg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "down", "j":
			if m.Cursor < len(m.Networks)-1 {
				m.Cursor++
			}
		case "r":
			m.Status = "Scanning..."
			m.Networks = network.ScanNetworks()
			m.Status = fmt.Sprintf("Found %d networks", len(m.Networks))
		case "enter":
			if len(m.Networks) == 0 {
				break
			}
			target := m.Networks[m.Cursor]

			if target.Active {
				m.Status = "Already connected"
				return m, nil
			}

			hasSaved := network.HasSavedConnection(target.SSID)
			m.UsedSaved = hasSaved

			if !hasSaved && target.Security != "--" && target.Security != "" {
				m.State = "password"
				m.TargetSSID = target.SSID
				m.PasswordInput.Focus()
				m.PasswordInput.Reset()
				m.Status = fmt.Sprintf("Password for %s", target.SSID)
			} else {
				m.State = "connecting"
				m.TargetSSID = target.SSID
				m.Status = fmt.Sprintf("Connecting to %s...", target.SSID)
				return m, m.connectCmd(target.SSID, "", hasSaved)
			}
		}
	}
	return m, nil
}

func (m Model) updatePassword(msg tea.Msg) (tea.Model, tea.Cmd) {
	if kmsg, ok := msg.(tea.KeyMsg); ok {
		switch kmsg.String() {
		case "esc":
			m.State = "list"
			m.PasswordInput.Blur()
			m.Status = "Cancelled"
			return m, nil
		case "enter":
			password := m.PasswordInput.Value()
			if password == "" {
				m.Status = "Password required"
				return m, nil
			}
			m.State = "connecting"
			m.Status = fmt.Sprintf("Connecting to %s...", m.TargetSSID)
			return m, m.connectCmd(m.TargetSSID, password, false)
		}
	}
	var cmd tea.Cmd
	m.PasswordInput, cmd = m.PasswordInput.Update(msg)
	return m, cmd
}

func (m Model) connectCmd(ssid, password string, useSaved bool) tea.Cmd {
	return func() tea.Msg {
		err := network.Connect(ssid, password, useSaved)
		return ConnectionResultMsg{
			Success: err == nil,
			SSID:    ssid,
		}
	}
}

func (m Model) View() string {
	data := ui.RenderData{
		Networks:          m.Networks,
		Cursor:            m.Cursor,
		State:             m.State,
		TargetSSID:        m.TargetSSID,
		Status:            m.Status,
		Width:             m.Width,
		RetryCount:        m.RetryCount,
		MaxRetries:        m.MaxRetries,
		PasswordInputView: m.PasswordInput.View(),
	}
	return ui.Render(data)
}
