package main

import (
	"strings"
	"time"

	"barista.run/bar"
	"barista.run/pango"
	"github.com/coreos/go-systemd/v22/dbus"
)

type SystemdMonitor struct {
	user bool
}

func (m SystemdMonitor) Stream(s bar.Sink) {
	var conn *dbus.Conn
	if m.user {
		c, err := dbus.NewUserConnection()
		if s.Error(err) {
			return
		}
		conn = c
	} else {
		c, err := dbus.NewSystemConnection()
		if s.Error(err) {
			return
		}
		conn = c
	}
	defer conn.Close()
	uCh, errCh := conn.SubscribeUnits(1 * time.Second)
	for {
		select {
		case <-uCh:
			stat, err := conn.SystemState()
			if s.Error(err) {
				return
			}
			if strings.Contains(stat.Value.String(), "running") {
				s.Output(pango.New(pango.Text(dot).Color(green)))
			} else {
				s.Output(pango.New(pango.Text(dot).Color(red)))
			}
		case err := <-errCh:
			if s.Error(err) {
				return
			}
		}
	}
}
