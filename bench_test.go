package main

import (
	"testing"

	"barista.run/bar"
)

func BenchmarkCheckInterface(b *testing.B) {
	fn := CheckInterface("wlan0")
	for i := 0; i < b.N; i++ {
		fn(func(output bar.Output) {})
	}
}

func BenchmarkWorldClock(b *testing.B) {
	fn := WorldClock()
	for i := 0; i < b.N; i++ {
		fn(func(output bar.Output) {})
	}
}
