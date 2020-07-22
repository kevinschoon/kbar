package systemd

import (
	"time"

	"barista.run/bar"
	"github.com/coreos/go-systemd/v22/dbus"
)

type SystemdState struct {
	UserState   string
	SystemState string
}

type SystemdStateMonitor struct {
	duration time.Duration
	fn       func(SystemdState) bar.Output
}

func New(duration time.Duration) *SystemdStateMonitor {
	return &SystemdStateMonitor{duration, nil}
}

func (m *SystemdStateMonitor) Output(fn func(SystemdState) bar.Output) *SystemdStateMonitor {
	m.fn = fn
	return m
}

func (m *SystemdStateMonitor) Stream(s bar.Sink) {
	userConn, err := dbus.NewUserConnection()
	if s.Error(err) {
		return
	}
	defer userConn.Close()
	sysConn, err := dbus.NewSystemConnection()
	if s.Error(err) {
		return
	}
	defer sysConn.Close()
	state := &SystemdState{}
	ticker := time.NewTicker(m.duration)
	for {
		<-ticker.C
		userState, err := userConn.SystemState()
		if s.Error(err) {
			return
		}
		state.UserState = userState.Value.String()
		systemState, err := sysConn.SystemState()
		if s.Error(err) {
			return
		}
		state.SystemState = systemState.Value.String()
		s.Output(m.fn(*state))
	}
}
