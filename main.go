package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
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
	dot  = `●`
	up   = `⬆️`
	down = `⬇️`
)

var (
	yellow = colors.Hex("#e4dd85")
	green  = colors.Hex("#66a461")
	red    = colors.Hex("#a46163")
)

func CheckWireguard(iface string) funcs.Func {
	return func(sink bar.Sink) {
		if _, err := os.Stat(fmt.Sprintf("/sys/class/net/%s/carrier", iface)); err != nil {
			sink.Output(
				pango.New(
					pango.Text(iface),
					pango.Text(down).Color(red)),
			)
			return
		}
		sink.Output(
			pango.New(pango.Text(iface),
				pango.Text(up).Color(green)),
		)
	}
}

func CheckInterface(iface string) funcs.Func {
	return func(sink bar.Sink) {
		raw, err := ioutil.ReadFile(fmt.Sprintf("/sys/class/net/%s/operstate", iface))
		if sink.Error(err) {
			return
		}
		if strings.Contains(string(raw), "up") {
			sink.Output(pango.New(pango.Text(iface).Color(green)))
		} else {
			sink.Output(pango.New(pango.Text(iface).Color(red)))
		}
	}
}

func WorldClock() funcs.Func {
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
	return bar.TextSegment(now.Format("(Mon)01-02|15:04:05|MST"))
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
		SwayWindow{},
		clock.Local().Output(1*time.Second, clockFunc),
		battery.Named("BAT0").Output(batteryFunc),
		sysinfo.New().Output(loadFunc),
		funcs.Every(10*time.Second, CheckInterface("wlan0")),
		funcs.Every(10*time.Second, CheckWireguard("wg0")),
		SystemdMonitor{true},
		SystemdMonitor{},
	}
}

func main() {
	barista.SetErrorHandler(func(err bar.ErrorEvent) {
		exec.Command("swaynag", "-m", err.Error.Error()).Run()
	})
	err := barista.Run(initModules()...)
	if err != nil {
		panic(err)
	}
}
