package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"time"

	"barista.run"
	"barista.run/bar"
	"barista.run/colors"
	"barista.run/modules/battery"
	"barista.run/modules/clock"
	"barista.run/modules/funcs"
	"barista.run/modules/meminfo"
	"barista.run/modules/netspeed"
	"barista.run/modules/sysinfo"
	"barista.run/modules/volume"
	"barista.run/modules/volume/pulseaudio"
	"barista.run/outputs"
	"barista.run/pango"

	"github.com/kevinschoon/kbar/systemd"
)

const (
	dot       = `●`
	up        = `↑`
	down      = `↓`
	connected = `⌁`
	moon      = "🌙"
)

var (
	yellow = colors.Hex("#e4dd85")
	green  = colors.Hex("#66a461")
	red    = colors.Hex("#a46163")
	white  = colors.Hex("#ffffff")
	gray   = colors.Hex("#b5b5b5")
	black  = colors.Hex("#000000")

	pctHigh = NewStyleizer(
		SetColor(black),
		Float(func(v float64) bool { return v > 25 }, SetColor(red)),
	)

	pctLow = NewStyleizer(
		SetColor(black),
		Float(func(v float64) bool { return v < 25 }, SetColor(red)),
	)

	// network speed (kb)
	speedChk = NewStyleizer(
		SetColor(black),
		Float(func(v float64) bool { return v > 100 }, SetColor(red)),
	)

	systemChk = NewStyleizer(
		SetColor(red),
		String(func(v string) bool { return strings.Contains(v, "running") }, SetColor(black)),
		String(func(v string) bool { return strings.Contains(v, "degraded") }, SetColor(red)),
	)

	clocks = NewStyleizer(
		SetColor(black), // night
		Float(func(v float64) bool { return (v >= 6 && v <= 8) }, SetColor(black)),   // dawn
		Float(func(v float64) bool { return (v > 8 && v < 18) }, SetColor(black)),    // daytime
		Float(func(v float64) bool { return (v >= 18 && v <= 20) }, SetColor(black)), // dusk
	)
)

func WorldClock() funcs.Func {
	return func(sink bar.Sink) {
		now := time.Now()
		sf := now.UTC().Add(-(7 * time.Hour))
		ny := now.UTC().Add(-(4 * time.Hour))
		ln := now.UTC()
		hk := now.UTC().Add((8 * time.Hour))
		sink.Output(pango.New(
			pango.Text("W:["),
			pango.Text("SF:"),
			clocks.Int(sf.Hour())(pango.Textf("%02d", sf.Hour())),
			pango.Text("|").Color(gray),
			pango.Text("NY:"),
			clocks.Int(ny.Hour())(pango.Textf("%02d", ny.Hour())),
			pango.Text("|").Color(gray),
			pango.Text("LN:"),
			clocks.Int(ln.Hour())(pango.Textf("%02d", ln.Hour())),
			pango.Text("|").Color(gray),
			pango.Text("HK:"),
			clocks.Int(hk.Hour())(pango.Textf("%02d", hk.Hour())),
			pango.Text("]"),
		))
	}
}

func initModules() []bar.Module {
	return []bar.Module{
		clock.Local().Output(1*time.Second, func(now time.Time) bar.Output {
			return pango.New(
				pango.Text("T:["),
				pango.Text(now.Format("15:04:05")).Color(black),
				pango.Text("|").Color(gray),
				pango.Text(now.Format("Mon Jan 2")).Color(black),
				pango.Text("]"),
			)
		}),
		funcs.Every(1*time.Minute, WorldClock()),
		volume.New(pulseaudio.DefaultSink()).Output(func(info volume.Volume) bar.Output {
			return pango.New(
				pango.Text("V:["),
				pango.Textf("%d", info.Pct()),
				pango.Text("]"),
			)
		}),
		battery.All().Output(func(info battery.Info) bar.Output {
			var symbol *pango.Node
			if info.Discharging() {
				symbol = pango.New(pango.Text(down))
			} else {
				symbol = pango.New(pango.Text(connected))
			}
			remaining := info.RemainingPct()
			return outputs.Pango(
				pango.New(
					pango.Text("B:["),
					symbol,
					pctLow.Int(remaining)(pango.Textf("%d", remaining)),
					pango.Text("]"),
				))
		}),
		sysinfo.New().Output(func(info sysinfo.Info) bar.Output {
			return pango.New(
				pango.Text("L:["),
				pctHigh.Float64((info.Loads[0] / float64(runtime.NumCPU()) * 100))(pango.Textf("%.2f", info.Loads[0])),
				pango.Text("|").Color(gray),
				pctHigh.Float64((info.Loads[1] / float64(runtime.NumCPU()) * 100))(pango.Textf("%.2f", info.Loads[1])),
				pango.Text("|").Color(gray),
				pctHigh.Float64((info.Loads[2] / float64(runtime.NumCPU()) * 100))(pango.Textf("%.2f", info.Loads[2])),
				pango.Text("]"),
			)
		}),
		meminfo.New().Output(func(info meminfo.Info) bar.Output {
			memUtilized := (info["MemTotal"].Gigabytes() - info.Available().Gigabytes())
			return pango.New(
				pango.Textf(
					"M:[%.2f]", memUtilized,
				),
			)
		}),
		netspeed.New("wlan0").Output(func(speeds netspeed.Speeds) bar.Output {
			if !speeds.Connected() {
				return pango.New(
					pango.Text("W:?").Color(red),
				)
			}
			tx := math.Round(speeds.Tx.BytesPerSecond() / 1000)
			rx := speeds.Rx.BytesPerSecond() / 1000
			return pango.New(
				pango.Text("N:["+up),
				speedChk.Float64(tx)(pango.Textf("%.2f", tx)),
				pango.Text("|").Color(gray),
				pango.Text(down),
				speedChk.Float64(tx)(pango.Textf("%.2f", rx)),
				pango.Text("]"),
			)
		}),
		systemd.New(5 * time.Second).Output(func(state systemd.SystemdState) bar.Output {
			user, system := pango.Text(
				strings.ToUpper(string(state.UserState[1])),
			), pango.Text(
				strings.ToUpper(string(state.SystemState[1])),
			)
			return pango.New(
				pango.Text("S:["),
				systemChk.String(state.UserState)(user),
				pango.Text("|").Color(gray),
				systemChk.String(state.SystemState)(system),
				pango.Text("]"),
			)
		}),
	}
}

func isSway() bool {
	return os.Getenv("SWAYSOCK") != ""
}

func main() {
	var (
		profile = flag.Bool("profile", false, "generate a pprof file")
	)
	flag.Parse()
	if *profile {
		go func() {
			log.Println(http.ListenAndServe("localhost:6060", nil))
		}()
	}
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, os.Interrupt)
	if isSway() {
		barista.SetErrorHandler(func(err bar.ErrorEvent) {
			exec.Command("swaynag", "-m", err.Error.Error()).Run()
		})
	}
	go func() {
		err := barista.Run(initModules()...)
		if err != nil {
			panic(err)
		}
	}()
	sig := <-sigCh
	fmt.Printf("caught signal (%s)\n", sig)
}
