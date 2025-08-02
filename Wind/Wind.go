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

	var baseline_sum uint32 = 0
	machine.Serial.Write([]byte("Calibrating... keep magnet away\n"))
	for i := 0; i < 10; i++ {
		baseline_sum += uint32(sensor.Get())
		time.Sleep(time.Millisecond * 100)
	}
	baseline := baseline_sum / 10

	machine.Serial.Write([]byte("Baseline: " + strconv.Itoa(int(baseline)) + " (no magnet)\n"))

	for {
		raw := sensor.Get()
		sensor_value := int(raw)

		voltage := (float32(raw) * 5.0) / 65535.0

		diff := int(raw) - int(baseline)

		magnet_status := "No Magnet"
		polarity := ""

		if diff > 100 {
			polarity = "NORTH POLE"
			magnet_status = "STRONG " + polarity + " (+" + strconv.Itoa(diff) + ")"
		} else if diff < -100 {
			polarity = "SOUTH POLE"
			magnet_status = "STRONG " + polarity + " (" + strconv.Itoa(diff) + ")"
		} else if diff > 20 {
			polarity = "north pole"
			magnet_status = "weak " + polarity + " (+" + strconv.Itoa(diff) + ")"
		} else if diff < -20 {
			polarity = "south pole"
			magnet_status = "weak " + polarity + " (" + strconv.Itoa(diff) + ")"
		}

		voltage_int := int(voltage * 1000)
		voltage_str := strconv.Itoa(voltage_int/1000) + "." +
			strconv.Itoa((voltage_int%1000)/100) +
			strconv.Itoa((voltage_int%100)/10) +
			strconv.Itoa(voltage_int%10)

		machine.Serial.Write([]byte("Raw: " + strconv.Itoa(sensor_value) +
			" | Voltage: " + voltage_str + "V" +
			" | Diff: " + strconv.Itoa(diff) +
			" | " + magnet_status + "\n"))

		time.Sleep(time.Second)
	}
}
