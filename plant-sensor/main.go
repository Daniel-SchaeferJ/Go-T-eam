package main

import (
	"machine"
	"strconv"
	"time"
)

const (
	minLux         = 1000
	maxLux         = 5000
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
	machine.Serial.Configure(machine.UARTConfig{
		BaudRate: 115200,
	})

	machine.InitADC()
	sensor := machine.ADC{Pin: machine.ADC0}
	sensor.Configure(machine.ADCConfig{})

	rw := NewRollingWindow(windowSize)

	machine.Serial.Write([]byte("Plant sensor started\n"))

	for {
		raw := sensor.Get()

		// Convert raw value to Lux.
		// NOTE: This conversion depends on the specific photoresistor and circuit.
		// For the KY-018 Photo-resistor Module, it's typically a voltage divider
		// with a 10k resistor. According to specs, Signal (S) should be HIGH in light
		// and LOW in dark when using standard wiring (+ to VCC, - to GND).
		// Raw values from machine.ADC are 0-65535.
		// Normalize value (1.0 to 0.0) based on 16-bit ADC (0-65535).
		// Higher raw values = DARKER.
		normalized := 1.0 - (float64(raw) / 65535.0)
		// Lux ≈ normalized^2 * 10000.0 (quadratic approximation fits better for LDRs).
		lux := normalized * normalized * 10000.0

		rw.Add(lux)
		avgLux := rw.Average()

		status := "Optimal"
		if avgLux < minLux {
			status = "Too Little"
		} else if avgLux > maxLux {
			status = "Too Much"
		}

		message := "Raw: " + strconv.Itoa(int(raw)) + " - Avg Lux: " + strconv.Itoa(int(avgLux)) + " - Status: " + status + "\n"
		machine.Serial.Write([]byte(message))

		time.Sleep(sampleInterval)
	}
}
