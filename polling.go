package main

import (
	"time"
)

// StartPolling will run a few functions every half second
func StartPolling() {
	for { // 0.5 seconds
		time.Sleep(500 * time.Millisecond)
		go UpdateCurrentImage()
		go UpdateCurrentCss()
	}
}
