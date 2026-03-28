//go:build simulator_slow

package main

import "time"

func initialise_ant() {
	ant_init(12347)
}

func RebroadcastThreats(liveRadarChan chan RadarPkt, sendRadarChan chan RadarPkt) {
	simulateCars(&sendRadarChan, (100 * time.Millisecond))
}
