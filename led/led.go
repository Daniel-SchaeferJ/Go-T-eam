package main

import (
	"machine"
	"strconv"
	"time"
)

func main() {
	led := machine.LED
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})

	machine.Serial.Configure(machine.UARTConfig{
		BaudRate: 115200,
	})
	info_count := 0
	for {
		led.Low()
		time.Sleep(time.Second)
		machine.Serial.Write([]byte("Hello World!\n"))
		info_count++
		machine.Serial.Write([]byte("Info count: " + strconv.Itoa(info_count) + "\n"))
		led.High()
		time.Sleep(time.Second)
	}
}
