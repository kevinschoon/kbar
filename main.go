package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"barista.run"
	"barista.run/bar"
	"barista.run/colors"
	"barista.run/modules/battery"
	"barista.run/modules/clock"
	"barista.run/modules/funcs"
	"barista.run/modules/sysinfo"
	"barista.run/outputs"
	"barista.run/pango"
)

const (
	dot = `●`
)

var (
	yellow = colors.Hex("#e4dd85")
	green  = colors.Hex("#66a461")
	red    = colors.Hex("#a46163")
)

type swayNode struct {
	ID      int        `json:"id"`
	Name    string     `json:"name"`
	Focused bool       `json:"focused"`
	Nodes   []swayNode `json:"nodes"`
}

func walk(root swayNode) string {
	if root.Focused {
		return root.Name
	}
	for _, node := range root.Nodes {
		if name := walk(node); name != "" {
			return name
		}
	}
	return ""
}

func maybe(err error) {
	if err != nil {
		panic(err)
	}
}

func getFocusedWindowTitle() (string, error) {
	// TODO: maybe there is a sway lib?
	// TODO: contribute to barista
	cmd := exec.Command("swaymsg", "-t", "get_tree")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	err = cmd.Start()
	if err != nil {
		return "", err
	}
	root := &swayNode{}
	err = json.NewDecoder(stdout).Decode(root)
	if err != nil {
		return "", err
	}
	return walk(*root), nil
}

func activeSwayWindow() funcs.Func {
	return func(sink bar.Sink) {
		window, err := getFocusedWindowTitle()
		maybe(err)
		sink.Output(bar.TextSegment(window))
	}
}

func checkInterface(iface string) funcs.Func {
	return func(sink bar.Sink) {
		raw, err := ioutil.ReadFile(fmt.Sprintf("/sys/class/net/%s/operstate", iface))
		maybe(err)
		if strings.ReplaceAll(string(raw), "\n", "") == "up" {
			sink.Output(pango.New(pango.Text(iface).Color(green)))
		} else {
			sink.Output(pango.New(pango.Text(iface).Color(red)))
		}
	}
}

func checkSystemdState(user bool) funcs.Func {
	// TODO: this should use dbus but that looks
	// awfully complicated..
	// https://www.freedesktop.org/wiki/Software/systemd/dbus/
	const systemctl = "systemctl"
	var args []string
	if user {
		args = append(args, "--user")
	}
	args = append(args, "status")
	return func(sink bar.Sink) {
		raw, err := exec.Command(systemctl, args...).CombinedOutput()
		maybe(err)
		if strings.Contains(string(raw), "State: running") {
			sink.Output(pango.New(pango.Text(dot)).Color(green))
		} else {
			sink.Output(pango.New(pango.Text(dot)).Color(red))
		}
	}
}

func worldClock() funcs.Func {
	return func(sink bar.Sink) {
		now := time.Now()
		buf := bytes.NewBuffer(nil)
		// San Francisco
		fmt.Fprintf(buf, "SF:%s", now.UTC().Add(-(7 * time.Hour)).Format("15"))
		// New York
		fmt.Fprintf(buf, "|NY:%s", now.UTC().Add(-(4 * time.Hour)).Format("15"))
		// London
		fmt.Fprintf(buf, "|LN:%s", now.UTC().Format("15"))
		// Beijing
		fmt.Fprintf(buf, "|BG:%s", now.UTC().Add((7 * time.Hour)).Format("15"))
		sink.Output(bar.TextSegment(buf.String()))
	}
}

func batteryFunc(info battery.Info) bar.Output {
	remaining := info.RemainingPct()
	var color colors.ColorfulColor
	switch {
	case remaining >= 75:
		color = green
	case remaining >= 25:
		color = yellow
	default:
		color = red
	}
	return outputs.Pango(
		pango.New(
			pango.Text(fmt.Sprintf("B:%d", remaining)),
		).Color(color))
}

func clockFunc(now time.Time) bar.Output {
	return bar.TextSegment(now.Format(time.UnixDate))
}

func loadToColor(value float64, nCpus uint16) colors.ColorfulColor {
	load := (value / float64(runtime.NumCPU())) * 100
	switch {
	case load > 50:
		return red
	case load > 25:
		return yellow
	default:
		return green
	}
}

func loadFunc(info sysinfo.Info) bar.Output {
	return pango.New(
		pango.Text("L:"),
		pango.Textf("%.2f|", info.Loads[0]).Color(loadToColor(info.Loads[0], info.Procs)),
		pango.Textf("%.2f|", info.Loads[1]).Color(loadToColor(info.Loads[1], info.Procs)),
		pango.Textf("%.2f", info.Loads[2]).Color(loadToColor(info.Loads[2], info.Procs)),
	)
}

func initModules() []bar.Module {
	return []bar.Module{
		funcs.Every(1*time.Second, activeSwayWindow()),
		clock.Local().Output(1*time.Second, clockFunc),
		battery.Named("BAT0").Output(batteryFunc),
		sysinfo.New().Output(loadFunc),
		funcs.Every(10*time.Second, checkInterface("wlan0")),
		funcs.Every(10*time.Second, checkInterface("wg0")),
	}
}

func main() {
	// typicons.Load()
	err := barista.Run(initModules()...)
	if err != nil {
		panic(err)
	}
}
