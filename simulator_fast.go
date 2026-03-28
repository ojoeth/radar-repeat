//go:build simulator_fast

package main

import "time"

func initialise_ant() {
	ant_init(12346)
}

func RebroadcastThreats(liveRadarChan chan RadarPkt, sendRadarChan chan RadarPkt) {
	simulateCars(&sendRadarChan, (20 * time.Millisecond))
}
