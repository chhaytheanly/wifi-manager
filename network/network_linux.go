package network

import (
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

func parseNmcliLine(line string) []string {
	var parts []string
	var current strings.Builder
	escape := false
	for _, r := range line {
		if escape {
			current.WriteRune(r)
			escape = false
		} else if r == '\\' {
			escape = true
		} else if r == ':' {
			parts = append(parts, current.String())
			current.Reset()
		} else {
			current.WriteRune(r)
		}
	}
	parts = append(parts, current.String())
	return parts
}

func ScanNetworks() []Network {
	out, err := exec.Command("nmcli", "-t", "-f", "IN-USE,SSID,SIGNAL,SECURITY", "device", "wifi", "list").Output()
	if err != nil {
		return nil
	}

	lines := strings.Split(string(out), "\n")
	var nets []Network
	seen := make(map[string]bool)

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := parseNmcliLine(line)
		if len(parts) >= 4 {
			ssid := parts[1]
			if ssid == "" || seen[ssid] {
				continue
			}
			seen[ssid] = true

			nets = append(nets, Network{
				Active:   parts[0] == "*",
				SSID:     ssid,
				Signal:   parts[2] + "%",
				Security: parts[3],
			})
		}
	}

	sort.Slice(nets, func(i, j int) bool {
		if nets[i].Active != nets[j].Active {
			return nets[i].Active
		}
		sigI, _ := strconv.Atoi(strings.TrimSuffix(nets[i].Signal, "%"))
		sigJ, _ := strconv.Atoi(strings.TrimSuffix(nets[j].Signal, "%"))
		return sigI > sigJ
	})

	return nets
}

func HasSavedConnection(ssid string) bool {
	out, err := exec.Command("nmcli", "-t", "-f", "NAME", "connection", "show").Output()
	if err != nil {
		return false
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		var current strings.Builder
		escape := false
		for _, r := range line {
			if escape {
				current.WriteRune(r)
				escape = false
			} else if r == '\\' {
				escape = true
			} else {
				current.WriteRune(r)
			}
		}
		if strings.TrimSpace(current.String()) == ssid {
			return true
		}
	}
	return false
}

func Connect(ssid, password string, useSaved bool) error {
	var cmd *exec.Cmd
	if useSaved {
		cmd = exec.Command("nmcli", "connection", "up", ssid)
	} else if password == "" {
		cmd = exec.Command("nmcli", "device", "wifi", "connect", ssid)
	} else {
		cmd = exec.Command("nmcli", "device", "wifi", "connect", ssid, "password", password)
	}
	return cmd.Run()
}
