package main

import (
	"machine"
	"strconv"
	"time"

	"tinygo.org/x/drivers/servo"
)

// Robot is the Elegoo SmartCar v4 using the Elegoo Smartcar Shield v1.1 and an Arduino Uno R3.
type Robot struct {

	// These are kind of working. Need to do further tweaking
	// TB6612 Motor Driver pin
	standby    machine.Pin // STBY pin 3 - must be HIGH to work
	leftSpeed  machine.Pin // PWMA pin 5 - "Speed" of LEFT motors
	rightSpeed machine.Pin // PWMB pin 6 - "Speed" of RIGHT motors
	leftDir    machine.Pin // AIN1 pin 7 - Power to LEFT motors
	rightDir   machine.Pin // BIN1 pin 8 - Power to RIGHT motors

	// Ultrasonic sensor HC-SR04
	ultrasonicTrig machine.Pin // trig pin 13
	ultrasonicEcho machine.Pin // echo pin 12

	// ^ The above are confirmed working

	// HUNCH BUT NEED TO TEST
	modePin machine.Pin // mode_pin 2
	rgbPin  machine.Pin // RGB_pin 4
	irPin   machine.Pin // IR pin 9
	servo   servo.Servo
	led     machine.Pin
}

func NewRobot() *Robot {
	return &Robot{
		// Motor driver pins - These work
		standby:    machine.D3, // STBY
		leftSpeed:  machine.D5, // PWMA
		rightSpeed: machine.D6, // PWMB
		leftDir:    machine.D7, // AIN1
		rightDir:   machine.D8, // BIN1

		// Ultrasonic sensor - works
		ultrasonicTrig: machine.D13, // trig
		ultrasonicEcho: machine.D12, // echo

		// Hunches
		modePin: machine.D2, // mode_pin
		rgbPin:  machine.D4, // RGB_pin
		irPin:   machine.D9, // IR
		led:     machine.LED,
	}
}

func (r *Robot) Initialize() {
	machine.Serial.Configure(machine.UARTConfig{
		BaudRate: 115200,
	})

	s, err := servo.New(machine.Timer1, machine.D10)
	if err != nil {
		for {
			machine.Serial.Write([]byte("could not configure servo"))
			time.Sleep(time.Second)
		}
	}
	r.standby.Configure(machine.PinConfig{Mode: machine.PinOutput})
	r.leftSpeed.Configure(machine.PinConfig{Mode: machine.PinOutput})
	r.rightSpeed.Configure(machine.PinConfig{Mode: machine.PinOutput})
	r.leftDir.Configure(machine.PinConfig{Mode: machine.PinOutput})
	r.rightDir.Configure(machine.PinConfig{Mode: machine.PinOutput})

	r.ultrasonicTrig.Configure(machine.PinConfig{Mode: machine.PinOutput})
	r.ultrasonicEcho.Configure(machine.PinConfig{Mode: machine.PinInput})

	r.modePin.Configure(machine.PinConfig{Mode: machine.PinInput})
	r.rgbPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	r.irPin.Configure(machine.PinConfig{Mode: machine.PinInput})
	r.servo = s
	r.ResetServoPosition() // make sure this doesn't break any gears

	r.led.Configure(machine.PinConfig{Mode: machine.PinOutput})

	r.standby.High()

	r.Stop()
}

func (r *Robot) Stop() {
	r.leftSpeed.Low()
	r.rightSpeed.Low()
	r.leftDir.Low()
	r.rightDir.Low()
}

func (r *Robot) MoveForward() {
	r.standby.High()
	r.leftDir.High()
	r.rightDir.High()
	r.leftSpeed.High()
	r.rightSpeed.High()
}

func (r *Robot) MoveBackward() {
	r.standby.High()
	r.leftDir.Low()
	r.rightDir.Low()
	r.leftSpeed.High()
	r.rightSpeed.High()
}

func (r *Robot) TurnLeft() {
	r.standby.High()
	r.leftDir.Low()
	r.rightDir.High()
	r.leftSpeed.High()
	r.rightSpeed.High()
}

func (r *Robot) TurnRight() {
	r.standby.High()
	r.leftDir.High()
	r.rightDir.Low()
	r.leftSpeed.High()
	r.rightSpeed.High()
}
func (r *Robot) ResetServoPosition() {
	r.servo.SetMicroseconds(1500)
}
func (r *Robot) TurnServoLeft() {
	r.servo.SetMicroseconds(2000)
}
func (r *Robot) TurnServoRight() {
	r.servo.SetMicroseconds(1000)
}

// TODO this needs to be reworked. Getting sensor sound inputs but distance is off
func (r *Robot) GetDistance() float64 {
	r.ultrasonicTrig.Low()
	time.Sleep(2 * time.Microsecond)

	r.ultrasonicTrig.High()
	time.Sleep(10 * time.Microsecond)
	r.ultrasonicTrig.Low()

	timeout := time.Now().Add(30 * time.Millisecond)
	for !r.ultrasonicEcho.Get() {
		if time.Now().After(timeout) {
			return 0
		}
	}

	start := time.Now()
	timeout = start.Add(30 * time.Millisecond)
	for r.ultrasonicEcho.Get() {
		if time.Now().After(timeout) {
			return 0 // Timeout - pulse too long
		}
	}
	end := time.Now()

	pulseTime := end.Sub(start).Microseconds()

	// Use the same formula as Arduino: distance = 0.5 * pulse_time * 0.0343
	// This accounts for sound traveling to object and back (hence 0.5)
	// 0.0343 is speed of sound in cm/μs
	distance := (float64(pulseTime) * 0.0343) / 2
	writeDistanceMessage(distance)
	return distance
}

func writeDistanceMessage(distance float64) {
	message := "Test " +
		": Dist=" + strconv.Itoa(int(distance)) + "cm"
	machine.Serial.Write([]byte(message + "\n"))
}

func turnAndGetDistance(robot *Robot) {
	for {

		robot.TurnServoLeft()
		robot.GetDistance()
		time.Sleep(time.Millisecond * 2000)
		robot.ResetServoPosition()
		robot.GetDistance()
		time.Sleep(time.Millisecond * 2000)
		robot.TurnServoRight()
		robot.GetDistance()
		time.Sleep(time.Millisecond * 2000)
		robot.ResetServoPosition()
		robot.GetDistance()
		time.Sleep(time.Millisecond * 2000)

	}
}
func main() {
	robot := NewRobot()
	robot.Initialize()

	machine.Serial.Write([]byte("Elegoo SmartCar V4 - Correct Pins!\n"))
	time.Sleep(1000 * time.Millisecond)

	machine.Serial.Write([]byte("Testing STBY pin...\n"))
	robot.standby.Low()
	time.Sleep(500 * time.Millisecond)
	robot.standby.High()
	machine.Serial.Write([]byte("Motors enabled!\n"))

	machine.Serial.Write([]byte("Starting main loop...\n"))

	turnAndGetDistance(robot)

}

//https://github.com/antonioastro/Elegoo-SmartCar/blob/main/Car_separate_files_v3.ino
