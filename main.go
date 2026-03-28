package main

import (
	"time"
)

func main() {
	// 1. Give the display a moment to settle after power-up
	time.Sleep(500 * time.Millisecond)

	initialise_ant()
	radarSendChan := make(chan RadarPkt)
	go processAntRadar(radarSendChan)

	radarRecvChan := make(chan RadarPkt)
	go func(radarRecvChan chan RadarPkt) {
		for {
			pollEvents(radarRecvChan)
		}
	}(radarRecvChan)

	go RebroadcastThreats(radarRecvChan, radarSendChan)

	select {}
}
