package main

import (
	"machine"
	"strconv"
	"time"
)

func main() {
	machine.Serial.Configure(machine.UARTConfig{
		BaudRate: 115200,
	})

	machine.InitADC()
	sensor := machine.ADC{Pin: machine.ADC5}
	sensor.Configure(machine.ADCConfig{})

	var baselineSum uint32 = 0
	_, err := machine.Serial.Write([]byte("Calibrating... keep magnet away\n"))
	if err != nil {
		return
	}
	for i := 0; i < 10; i++ {
		baselineSum += uint32(sensor.Get())
		time.Sleep(time.Millisecond * 100)
	}
	baseline := baselineSum / 10

	_, err = machine.Serial.Write([]byte("Baseline: " + strconv.Itoa(int(baseline)) + " (no magnet)\n"))
	if err != nil {
		return
	}

	for {
		raw := sensor.Get()
		sensorValue := int(raw)

		voltage := (float32(raw) * 5.0) / 65535.0

		diff := int(raw) - int(baseline)

		magnetStatus := "No Magnet"
		polarity := ""

		if diff > 100 {
			polarity = "NORTH POLE"
			magnetStatus = "STRONG " + polarity + " (+" + strconv.Itoa(diff) + ")"
		} else if diff < -100 {
			polarity = "SOUTH POLE"
			magnetStatus = "STRONG " + polarity + " (" + strconv.Itoa(diff) + ")"
		} else if diff > 20 {
			polarity = "north pole"
			magnetStatus = "weak " + polarity + " (+" + strconv.Itoa(diff) + ")"
		} else if diff < -20 {
			polarity = "south pole"
			magnetStatus = "weak " + polarity + " (" + strconv.Itoa(diff) + ")"
		}

		voltageInt := int(voltage * 1000)
		voltageStr := strconv.Itoa(voltageInt/1000) + "." +
			strconv.Itoa((voltageInt%1000)/100) +
			strconv.Itoa((voltageInt%100)/10) +
			strconv.Itoa(voltageInt%10)

		_, err := machine.Serial.Write([]byte("Raw: " + strconv.Itoa(sensorValue) +
			" | Voltage: " + voltageStr + "V" +
			" | Diff: " + strconv.Itoa(diff) +
			" | " + magnetStatus + "\n"))
		if err != nil {
			return
		}

		time.Sleep(time.Second)
	}
}
