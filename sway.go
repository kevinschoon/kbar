package main

import (
	"encoding/json"
	"os/exec"

	"barista.run/bar"
)

type swayMsg struct {
	Change    string `json:"change"`
	Container struct {
		ID      int    `json:"id"`
		Name    string `json:"name"`
		Focused bool   `json:"focused"`
	} `json:"container"`
}

type SwayWindow struct{}

func (w SwayWindow) Stream(s bar.Sink) {
	cmd := exec.Command("swaymsg", "-t", "subscribe", "-m", `["window"]`)
	stdout, err := cmd.StdoutPipe()
	if s.Error(err) {
		return
	}
	if s.Error(cmd.Start()) {
		return
	}
	decoder := json.NewDecoder(stdout)
	for {
		msg := &swayMsg{}
		if s.Error(decoder.Decode(msg)) {
			return
		}
		s.Output(bar.TextSegment(msg.Container.Name))
	}
}
