package main

import (
	"time"
)

func main() {
	// 1. Give the display a moment to settle after power-up
	time.Sleep(500 * time.Millisecond)

	/*// 2. Initialise I2C (nice!nano v2 defaults: SDA=P0.17, SCL=P0.20)
	machine.I2C0.Configure(machine.I2CConfig{
		Frequency: machine.TWI_FREQ_400KHZ,
	})

	display := ssd1306.NewI2C(machine.I2C0)
	display.Configure(ssd1306.Config{
		Address: 0x3C, // Standard address
		Width:   128,
		Height:  32,
	})*/

	ant_init()

	radarSendChan := make(chan RadarPkt)
	go processAntRadar(radarSendChan)

	//go simulateCars(&radarSendChan)
	//ctx, cancel := context.WithCancel(context.Background())
	//go FlashMsg(ctx, display, "Car Back!")

	radarRecvChan := make(chan RadarPkt)
	go func(radarRecvChan chan RadarPkt) {
		for {
			pollEvents(radarRecvChan)
		}
	}(radarRecvChan)

	go RebroadcastThreats(radarRecvChan, radarSendChan)
	//time.Sleep(30 * time.Second)
	//cancel()
	select {}
}

func simulateCars(radarChan *chan RadarPkt) {
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
	for i := 0; i < 500; i++ {
		//radarChan <- RadarPkt{}
		*radarChan <- radardata
		time.Sleep(50 * time.Millisecond)
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

	// clear down at end of sim
	*radarChan <- RadarPkt{}

}
