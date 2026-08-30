package ui

import (
	"fmt"
	"strconv"
	"strings"

	"wifi-tui/network"

	"github.com/charmbracelet/lipgloss"
)

const (
	IconCursor   = "\uf105"
	IconActive   = "\uf00c"
	IconLock     = "\uf023"
	IconScanning = "\uf110"
	IconError    = "\uf06a"
	IconSuccess  = "\uf058"
	IconWifi     = "\uf1eb"
)

var (
	TitleStyle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#89B4FA")).MarginLeft(2).MarginBottom(1)
	HelpStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("#585B70"))
	SelectedStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#A6E3A1"))
	NormalStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#CDD6F4"))
	ActiveStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#A6E3A1"))
	ErrorStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#F38BA8"))
	SuccessStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#A6E3A1"))
	InputPrompt     = lipgloss.NewStyle().Foreground(lipgloss.Color("#89B4FA"))
	InputBox        = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#89B4FA"))
	ConnectingStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#F9E2AF"))
)

type RenderData struct {
	Networks          []network.Network
	Cursor            int
	State             string
	TargetSSID        string
	Status            string
	Width             int
	RetryCount        int
	MaxRetries        int
	PasswordInputView string
}

func Render(data RenderData) string {
	s := TitleStyle.Render("  "+"°˖* ૮(  • ᴗ ｡)っ🍸 "+" WiFi Manager") + "\n"

	switch data.State {
	case "list":
		if len(data.Networks) == 0 {
			s += "\n" + NormalStyle.Render("No networks found") + "\n"
			s += HelpStyle.Render("  Press 'r' to scan again") + "\n"
		} else {
			var items []string
			for i, net := range data.Networks {
				cursor := "  "
				if data.Cursor == i {
					cursor = ConnectingStyle.Render(IconCursor + " ")
				}

				activeTag := "  "
				if net.Active {
					activeTag = "  " + ActiveStyle.Render(IconActive)
				}

				lock := "  "
				if net.Security != "--" && net.Security != "" {
					lock = "  " + ConnectingStyle.Render(IconLock)
				}

				ssid := net.SSID
				if len([]rune(ssid)) > 22 {
					ssid = string([]rune(ssid)[:20]) + ".."
				}

				sig, _ := strconv.Atoi(strings.TrimSuffix(net.Signal, "%"))
				var signalStyle lipgloss.Style
				switch {
				case sig >= 70:
					signalStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#A6E3A1"))
				case sig >= 40:
					signalStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#F9E2AF"))
				default:
					signalStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#F38BA8"))
				}

				name := fmt.Sprintf("%s%-24s%s %s %s", cursor, ssid, lock, signalStyle.Render(net.Signal), activeTag)

				if data.Cursor == i {
					items = append(items, SelectedStyle.Render(name))
				} else {
					items = append(items, NormalStyle.Render(name))
				}
			}
			s += "\n" + strings.Join(items, "\n") + "\n"
		}
		s += HelpStyle.Render("\n  j/k ↑↓ move • enter connect • r scan • q quit") + "\n"

	case "password":
		s += "\n"
		s += InputPrompt.Render(fmt.Sprintf("  Password for %s", data.TargetSSID)) + "\n"
		s += InputBox.Render(data.PasswordInputView) + "\n"
		s += HelpStyle.Render("  enter connect · esc cancel") + "\n"

	case "connecting":
		s += "\n"
		s += ConnectingStyle.Render(fmt.Sprintf("  %s Connecting to %s...", IconScanning, data.TargetSSID)) + "\n"
		s += HelpStyle.Render("  Please wait") + "\n"
	}

	width := data.Width
	if width == 0 {
		width = 80
	}

	content := s
	switch {
	case strings.Contains(data.Status, "Connected"):
		content += SuccessStyle.Render(fmt.Sprintf("  %s %s", IconSuccess, data.Status)) + "\n"
	case strings.Contains(data.Status, "Failed") || strings.Contains(data.Status, "Wrong") || strings.Contains(data.Status, "required") || strings.Contains(data.Status, "Saved failed"):
		content += ErrorStyle.Render(fmt.Sprintf("  %s %s", IconError, data.Status)) + "\n"
	case strings.Contains(data.Status, "Scanning") || strings.Contains(data.Status, "Connecting") || strings.Contains(data.Status, "Password"):
		content += ConnectingStyle.Render(fmt.Sprintf("  %s %s", IconScanning, data.Status)) + "\n"
	default:
		content += NormalStyle.Render(fmt.Sprintf("  %s %s", IconWifi, data.Status)) + "\n"
	}

	return lipgloss.NewStyle().Width(width).Align(lipgloss.Center).Render(content)
}
