package main

import (
	"fmt"
	"log"

	"github.com/Naikelin/atom-he68-pro-configurator/desktop/internal/hid"
)

func main() {
	var dev hid.Device
	if err := dev.Connect(); err != nil {
		log.Fatalf("connect failed: %v", err)
	}
	defer dev.Disconnect()

	resp, err := dev.HIDCommand(0x02, []byte{0x06})
	if err != nil {
		log.Fatalf("get keycodes version failed: %v", err)
	}
	fmt.Printf("Keycodes version response: %v\n", resp)
}
