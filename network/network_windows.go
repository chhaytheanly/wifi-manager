package network

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

func ScanNetworks() []Network {
	activeSSID := ""
	outInterfaces, err := exec.Command("netsh", "wlan", "show", "interfaces").Output()
	if err == nil {
		lines := strings.Split(string(outInterfaces), "\n")
		for _, line := range lines {
			if strings.Contains(line, "SSID") && !strings.Contains(line, "BSSID") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					activeSSID = strings.TrimSpace(parts[1])
				}
				break
			}
		}
	}

	out, err := exec.Command("netsh", "wlan", "show", "networks", "mode=bssid").Output()
	if err != nil {
		return nil
	}

	var nets []Network
	seen := make(map[string]bool)

	lines := strings.Split(string(out), "\n")
	var currentNetwork *Network

	for _, line := range lines {
		line = strings.TrimRight(line, "\r")
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "SSID ") {
			if currentNetwork != nil && currentNetwork.SSID != "" {
				if !seen[currentNetwork.SSID] {
					nets = append(nets, *currentNetwork)
					seen[currentNetwork.SSID] = true
				}
			}
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				ssid := strings.TrimSpace(parts[1])
				currentNetwork = &Network{
					SSID:   ssid,
					Active: ssid == activeSSID && ssid != "",
				}
			}
		} else if currentNetwork != nil {
			if strings.HasPrefix(trimmed, "Authentication") {
				parts := strings.SplitN(trimmed, ":", 2)
				if len(parts) == 2 {
					currentNetwork.Security = strings.TrimSpace(parts[1])
				}
			} else if strings.HasPrefix(trimmed, "Signal") {
				parts := strings.SplitN(trimmed, ":", 2)
				if len(parts) == 2 {
					if currentNetwork.Signal == "" { // Use first BSSID signal
						currentNetwork.Signal = strings.TrimSpace(parts[1])
					}
				}
			}
		}
	}

	if currentNetwork != nil && currentNetwork.SSID != "" {
		if !seen[currentNetwork.SSID] {
			nets = append(nets, *currentNetwork)
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
	out, err := exec.Command("netsh", "wlan", "show", "profiles").Output()
	if err != nil {
		return false
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.Contains(line, "All User Profile") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				profile := strings.TrimSpace(parts[1])
				profile = strings.TrimRight(profile, "\r")
				if profile == ssid {
					return true
				}
			}
		}
	}
	return false
}

func Connect(ssid, password string, useSaved bool) error {
	if useSaved {
		return exec.Command("netsh", "wlan", "connect", "name=\""+ssid+"\"").Run()
	}

	if password != "" {
		xmlProfile := fmt.Sprintf(`<?xml version="1.0"?>
<WLANProfile xmlns="http://www.microsoft.com/networking/WLAN/profile/v1">
	<name>%s</name>
	<SSIDConfig>
		<SSID>
			<name>%s</name>
		</SSID>
	</SSIDConfig>
	<connectionType>ESS</connectionType>
	<connectionMode>auto</connectionMode>
	<MSM>
		<security>
			<authEncryption>
				<authentication>WPA2PSK</authentication>
				<encryption>AES</encryption>
				<useOneX>false</useOneX>
			</authEncryption>
			<sharedKey>
				<keyType>passPhrase</keyType>
				<protected>false</protected>
				<keyMaterial>%s</keyMaterial>
			</sharedKey>
		</security>
	</MSM>
</WLANProfile>`, ssid, ssid, password)

		tmpFile, err := os.CreateTemp("", "wifi-profile-*.xml")
		if err != nil {
			return err
		}
		defer os.Remove(tmpFile.Name())

		_, err = tmpFile.WriteString(xmlProfile)
		tmpFile.Close()
		if err != nil {
			return err
		}

		err = exec.Command("netsh", "wlan", "add", "profile", "filename=\""+tmpFile.Name()+"\"").Run()
		if err != nil {
			return err
		}
	}

	return exec.Command("netsh", "wlan", "connect", "name=\""+ssid+"\"").Run()
}
