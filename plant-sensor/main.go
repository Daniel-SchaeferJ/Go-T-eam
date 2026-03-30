package main

import (
	"machine"
	"time"
)

const (
	minLux         = 1000
	maxLux         = 3000
	windowSize     = 10
	sampleInterval = 500 * time.Millisecond
)

type RollingWindow struct {
	values []float64
	next   int
	isFull bool
}

func NewRollingWindow(size int) *RollingWindow {
	return &RollingWindow{
		values: make([]float64, size),
	}
}

func (rw *RollingWindow) Add(val float64) {
	rw.values[rw.next] = val
	rw.next = (rw.next + 1) % len(rw.values)
	if rw.next == 0 {
		rw.isFull = true
	}
}

func (rw *RollingWindow) Average() float64 {
	count := len(rw.values)
	if !rw.isFull {
		count = rw.next
	}
	if count == 0 {
		return 0
	}

	var sum float64
	for i := range count {
		sum += rw.values[i]
	}
	return sum / float64(count)
}

func main() {
	machine.InitADC()
	sensor := machine.ADC{Pin: machine.ADC0}
	sensor.Configure(machine.ADCConfig{})

	rw := NewRollingWindow(windowSize)

	for {
		raw := sensor.Get()

		// Convert raw value to Lux.
		// NOTE: This conversion depends on the specific photoresistor and circuit.
		// For the KY-018 Photo-resistor Module, it's typically a voltage divider
		// with a 10k resistor.
		// Raw values from machine.ADC are 0-65535.
		// Higher raw values usually mean more light (if 'S' is connected to ADC and '+' to Vcc).
		// For common LDRs, lux ≈ (constant / resistance)^exponent.
		// We'll use a slightly more informed linear approximation for demonstration.
		// 0 Lux (dark) -> 0; ~3000 Lux (bright indirect) -> ~20000; ~10000 Lux (full sun) -> 65535.
		lux := float64(raw) * 10000.0 / 65535.0

		rw.Add(lux)
		avgLux := rw.Average()

		status := "Optimal"
		if avgLux < minLux {
			status = "Too Little"
		} else if avgLux > maxLux {
			status = "Too Much"
		}

		print("Avg Lux: ")
		print(int(avgLux))
		print(" - Status: ")
		println(status)

		time.Sleep(sampleInterval)
	}
}
