package main

import "time"

func simulateCars(radarChan *chan RadarPkt, changeInterval time.Duration) {
	radardata := RadarPkt{
		[4]Threat{
			{
				ThreatLevel:  1,
				ThreatSides:  0,
				Distance:     10,
				ClosingSpeed: 40,
			},
			{
				ThreatLevel:  1,
				ThreatSides:  0,
				Distance:     20,
				ClosingSpeed: 40,
			},
			{
				ThreatLevel:  1,
				ThreatSides:  0,
				Distance:     30,
				ClosingSpeed: 40,
			},
		},
	}
	for {
		//radarChan <- RadarPkt{}
		*radarChan <- radardata
		time.Sleep(changeInterval)
		for i := range radardata.Threats {
			if radardata.Threats[i].Distance > 1 {
				radardata.Threats[i].Distance -= 1
			} else {
				radardata.Threats[i].Distance += 100
			}
			/*
				if rand.Intn(20) == 0 {
					radardata.Threats[i].ThreatLevel = 2
				} else if rand.Intn(30) == 0 {
					*radarChan <- RadarPkt{}
				} else {
					radardata.Threats[i].ThreatLevel = 1
					}*/
		}
	}
}
