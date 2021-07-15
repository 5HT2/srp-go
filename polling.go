package main

import (
	"time"
)

// StartPolling will run a few functions every second
func StartPolling() {
	for {
		time.Sleep(1 * time.Second)
		go UpdateCurrentImage()
		go UpdateCurrentCss()
	}
}
