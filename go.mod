module github.com/kevinschoon/kbar

go 1.14

require (
	barista.run v0.0.0-20200711163523-206a19a855ae
	github.com/coreos/go-systemd/v22 v22.1.0
	github.com/dustin/go-humanize v1.0.0
	github.com/martinlindhe/unit v0.0.0-20190604142932-3b6be53d49af
)

replace barista.run => ../../soumya92/barista
