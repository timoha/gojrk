package main

import (
	"flag"
	"log"

	"github.com/timoha/gojrk"
)

var target = flag.Int("target", -1, "target to set the motor at (0-4095)")
var device = flag.String("device", "", "filepath to device")

func main() {
	flag.Parse()

	if *target < 0 || *target > 4095 {
		log.Fatalln("target has to be in range 0 to 4095")
	}

	j, err := gojrk.NewJRK(*device)
	if err != nil {
		log.Fatalln("failed to connect to jrk controller:", err)
	}
	defer j.Close()

	if f, err := j.Feedback(); err != nil {
		log.Fatalln("failed to get current feedback value", err)
	} else {
		log.Println("Current feedback:", f)
	}

	if t, err := j.Target(); err != nil {
		log.Fatalln("failed to get current target value", err)
	} else {
		log.Println("Current target:", t)
	}

	if err := j.SetTarget(*target); err != nil {
		log.Fatalln("failed to set new target value:", err)
	}
}
