package main

/*
#cgo CFLAGS: -I./lib/s340/include
#cgo CFLAGS: -I./lib/nrf_headers/device
#cgo CFLAGS: -I./lib/nrf_headers/cmsis
#cgo CFLAGS: -I./lib/nrf_headers/mdk

#cgo CFLAGS: -DNRF52840_XXAA
#include "ant_glue.h"
*/
import "C"
import (
	"math/rand"
	"time"
)

type Threat struct {
	ThreatLevel  int     // 0=none, 1=approaching, 2=fast approach
	ThreatSides  int     // 0=direct behind 1=left 2=right
	Distance     float32 // meters
	ClosingSpeed float32 // m/s
}

type RadarPkt struct {
	Threats [4]Threat
}

func (data RadarPkt) ToBytes() [8]byte {
	//			payload = [8]byte{0x30, 0x01, 0x00, 0x0A, 0x00, 0x00, 0x50, 0x00} // Page 48: Target
	pkt := [8]byte{}

	// Byte 0 (radar packet type)
	pkt[0] = 0x30

	// Byte 1 (Threat levels)
	numBits := 2
	mask := (1 << numBits) - 1
	for i := 0; i < len(data.Threats); i++ {
		pkt[1] |= (byte(data.Threats[i].ThreatLevel&mask) << (numBits * i))
	}

	// Byte 2 (Sides)
	numBits = 2
	mask = (1 << numBits) - 1
	for i := 0; i < len(data.Threats); i++ {
		pkt[2] |= (byte(data.Threats[i].ThreatSides&mask) << (numBits * i))
	}

	// Bytes 3 - 5 (Distance)
	var packedRange uint32
	mask = (1 << 6) - 1 // 6 bits

	for i := 0; i < len(data.Threats); i++ {
		val := uint32(int(data.Threats[i].Distance/3.125) & mask)
		packedRange |= (val << (i * 6))
	}
	pkt[3] = byte(packedRange)
	pkt[4] = byte(packedRange >> 8)
	pkt[5] = byte(packedRange >> 16)

	// Bytes 6 - 7 (closing speed)
	numBits = 4
	mask = (1 << numBits) - 1
	for i := 0; i < len(data.Threats); i++ {
		pktnum := int(i/2) + 6
		pkt[pktnum] |= (byte(int(data.Threats[i].ClosingSpeed/3.04)&mask) << (numBits * (i % 2)))
	}

	return pkt
}

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

	println("Setting up RADAR ANT+")

	res := C.enable_softdevice()
	if res != 0 {
		time.Sleep(time.Second)
		println("FAILED: enable softdevice returned error code: 0x", uint32(res))
	} else {
		println("Setting up soft device success")
	}

	e := C.setup_radar_channel()
	if e != 0 {
		println("FAILED: Setup returned error code: 0x", uint32(e))
	} else {
		println("Setting up radar chan success")
	}

	radarChan := make(chan RadarPkt)
	go processAntRadar(radarChan)

	simulateCars(&radarChan)
	//ctx, cancel := context.WithCancel(context.Background())
	//go FlashMsg(ctx, display, "Car Back!")

	time.Sleep(30 * time.Second)

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
				ClosingSpeed: 10,
			},
			{
				ThreatLevel:  1,
				ThreatSides:  0,
				Distance:     20,
				ClosingSpeed: 10,
			},
			{
				ThreatLevel:  1,
				ThreatSides:  0,
				Distance:     30,
				ClosingSpeed: 10,
			},
		},
	}
	for i := 0; i < 500; i++ {
		//radarChan <- RadarPkt{}
		*radarChan <- radardata
		time.Sleep(200 * time.Millisecond)
		for i := range radardata.Threats {
			if radardata.Threats[i].Distance > 10 {
				radardata.Threats[i].Distance -= 10
			} else {
				radardata.Threats[i].Distance += 100
			}

			if rand.Intn(20) == 0 {
				radardata.Threats[i].ThreatLevel = 2
			} else if rand.Intn(30) == 0 {
				*radarChan <- RadarPkt{}
			} else {
				radardata.Threats[i].ThreatLevel = 1
			}
		}
	}
}

func processAntRadar(radar chan RadarPkt) {
	count := 0

	for {
		// Page 1, 1 Target, ID 0, 20m away, 10m/s, Threat Level 2 (High)
		var payload [8]byte
		// Every 65 messages, send a background page
		if count%195 == 64 {
			payload = [8]byte{0x50, 0xFF, 0xFF, 123, 0x00, 0x01, 0x00, 0x01} // Page 80: Mfg ID
		} else if count%195 == 129 {
			payload = [8]byte{0x51, 0xFF, 0x01, 0x00, 0x00, 0x00, 0x00, 0x01} // Page 81: Product
		} else if count%195 == 194 {
			payload = [8]byte{0x52, 0xFF, 0xFF, 0xFF, 0x00, 0x00, 0x00, 0x04} // Page 82: Battery OK
		} else {
			radarData := <-radar
			if len(radarData.Threats) > 0 {
				payload = radarData.ToBytes()
			} else {
				payload = [8]byte{0x30, 0x01, 0x00, 0x0A, 0x00, 0x00, 0x50, 0x00} // Page 48: Target
			}
		}
		err := C.send_radar_update((*C.uint8_t)(&payload[0]))

		msg := ""
		for _, i := range payload {
			msg = msg + string(uint8(i))
		}
		if err != 0 {
			println("Error sending update: ", err)
		} //println("sent radar update")
		C.process_ant_events()
		count++
		time.Sleep(125 * time.Millisecond)
	}
}
