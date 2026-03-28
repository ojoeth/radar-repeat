//go:build !simulator_fast && !simulator_slow

package main

func initialise_ant() {
	ant_init(12345)
}

func RebroadcastThreats(liveRadarChan chan RadarPkt, sendRadarChan chan RadarPkt) {
	packets := []RadarPkt{}
	for {
		receivedPkt := <-liveRadarChan
		packets = append(packets, receivedPkt)
		if len(packets) == 4 {
			toSend := ConsolidateThreats(packets)
			sendRadarChan <- toSend
			packets = []RadarPkt{}
		}
	}
}
