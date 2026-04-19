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
	"sync"
	"sync/atomic"
	"time"
)

const NUMBER_OF_CHANS = 8 // Hardcoded at 8. Do not change.
var channelRestarting [8]atomic.Bool
var antInitOnce sync.Once

type Threat struct {
	ThreatLevel  int     // 0=none, 1=approaching, 2=fast approach
	ThreatSides  int     // 0=direct behind 1=right 2=left
	Distance     float32 // meters
	ClosingSpeed float32 // m/s
}

type RadarPkt struct {
	Threats [4]Threat
}

func checkAndRestartDeadChannels() {
	for chanNum := 1; chanNum < NUMBER_OF_CHANS; chanNum++ {
		status := uint8(C.get_channel_status(C.uint8_t(uint8(chanNum))))
		if status == 0x00 || status == 0x01 { // unassigned or assigned-but-not-open
			if channelRestarting[chanNum].CompareAndSwap(false, true) {
				go restartChannelWithDelay(uint8(chanNum), 0)
			}
		}
	}
}

func RadarPktFromBytes(pkt [8]byte) RadarPkt {
	radardata := RadarPkt{}
	// Byte 1 (Threat levels)
	numBits := 2
	mask := (1 << numBits) - 1
	for i := 0; i < 4; i++ {
		rawbits := (pkt[1] >> (numBits * i)) & byte(mask)
		if rawbits != 0 {
			radardata.Threats[i].ThreatLevel = int(rawbits)
		}
	}

	// Byte 2 (Threat sides)
	for i := 0; i < 4; i++ {
		rawbits := (pkt[2] >> (numBits * i)) & byte(mask)
		if rawbits != 0 {
			radardata.Threats[i].ThreatSides = int(rawbits)
		}
	}

	// Bytes 3-5 (distance)
	numBits = 6
	mask = (1 << numBits) - 1
	var packed uint32 = uint32(pkt[3]) | (uint32(pkt[4]) << 8) | (uint32(pkt[5]) << 16)
	for i := 0; i < 4; i++ {
		rawbits := (packed >> (numBits * i)) & uint32(mask)
		if rawbits != 0 {
			radardata.Threats[i].Distance = float32(rawbits) * 3.125
		}
	}

	// Bytes 6 - 7 (closing speed)
	numBits = 4
	mask = (1 << numBits) - 1
	for i := 0; i < 4; i++ {
		pktnum := int(i/2) + 6

		rawbits := (pkt[pktnum] >> (numBits * (i % 2))) & byte(mask)
		if rawbits != 0 {
			radardata.Threats[i].ClosingSpeed = float32(rawbits) * 3.04
		}
	}

	return radardata
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

func ant_init(radarDeviceNum int) {
	antInitOnce.Do(func() { // Ensure this only gets called once!
		debug("Setting up RADAR ANT+")

		res := C.enable_softdevice()
		if res != 0 {
			time.Sleep(time.Second)
			debug("FAILED: enable softdevice returned error code: 0x", uint32(res))
		} else {
			debug("Setting up soft device success")
		}

		e := C.setup_radar_channel(uint16(radarDeviceNum))
		if e != 0 {
			debug("FAILED: Setup returned error code: 0x", uint32(e))
		} else {
			debug("Setting up radar chan success")
		}

		for i := 1; i <= NUMBER_OF_CHANS-1; i++ {
			e = C.setup_radar_receive_channel(uint32(i))
			if e != 0 {
				debug("FAILED: Setup of channel", i, "returned error code: 0x", uint32(e))
			} else {
				debug("Setting up radar scan channel", i, "success")
			}
		}

		go func() {
			for {
				time.Sleep(time.Minute)
				checkAndRestartDeadChannels()
			}
		}()
	})
}

func pollEvents(radarChan chan RadarPkt) {
	var event uint8
	var channel uint8
	var data [8]byte

	for {
		err := C.wait_for_ant_event(
			(*C.uint8_t)(&channel),
			(*C.uint8_t)(&event),
			(*C.uint8_t)(&data[0]),
		)

		if err == 0 {
			handleEvent(radarChan, channel, event, data)
		}
	}
}

func handleEvent(radarChan chan RadarPkt, channel uint8, event uint8, data [8]byte) {
	switch event {
	case 0x4E: // MESG_BROADCAST_DATA_ID
		//println("Radar Data Received: ", data)
	case 0x40: // MESG_RESPONSE_EVENT_ID
		eventCode := data[0]
		if eventCode == 3 {
			// Just a successful transmit, maybe ignore this.
		} else if eventCode == 1 {
			debug("Search timed out. No radars nearby.")
		}
	case 1: // timed out
		if channel > 0 && int(channel) < len(channelRestarting) {
			if channelRestarting[channel].CompareAndSwap(false, true) {
				go restartChannelWithDelay(channel, 30*time.Second)
			}
		}
	case 0x07, 0x08, 0x0B: // chan closed
		if channel > 0 && int(channel) < len(channelRestarting) {
			if channelRestarting[channel].CompareAndSwap(false, true) {
				go restartChannelWithDelay(channel, 3*time.Second)
			}
		}
	case 0x35:
		checkAndRestartDeadChannels()
	case 0x80: // actual packet
		if data[0] == 0x30 { // radar pkt
			radarChan <- RadarPktFromBytes(data)
		}
	}
}

func restartChannelWithDelay(channel uint8, delay time.Duration) {
	defer channelRestarting[channel].Store(false)
	debug("Channel ", channel, "dropped. Restarting...")
	time.Sleep(delay)
	C.setup_radar_receive_channel(uint32(channel))
}

func processAntRadar(radar chan RadarPkt) {
	count := 0
	var lastRadarData RadarPkt

	for {
		var payload [8]byte
		if count%195 == 64 {
			payload = [8]byte{0x50, 0xFF, 0xFF, 123, 0x00, 0x01, 0x00, 0x01} // Page 80: Mfg ID
		} else if count%195 == 129 {
			payload = [8]byte{0x51, 0xFF, 0x01, 0x00, 0x00, 0x00, 0x00, 0x01} // Page 81: Product
		} else if count%195 == 194 {
			payload = [8]byte{0x52, 0xFF, 0x01, 0x00, 0x00, 0x00, 0x00, 0x13} // Page 82: Battery OK
		} else {
			select {
			case radarData := <-radar:
				lastRadarData = radarData
			default:
			}
			if len(lastRadarData.Threats) > 0 {
				payload = lastRadarData.ToBytes()
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
			debug("Error sending update: ", err)
		}
		//C.process_ant_events()
		count++
		time.Sleep(125 * time.Millisecond)
	}
}

func ConsolidateThreats(packets []RadarPkt) RadarPkt {
	threatsPerPacket := make([]int, len(packets))
	for i, pkt := range packets {
		for _, threat := range pkt.Threats {
			if threat.ThreatLevel != 0 {
				threatsPerPacket[i]++
			}
		}
	}

	var realThreatNum int
	for _, threatnum := range threatsPerPacket {
		if realThreatNum < threatnum {
			realThreatNum = threatnum
		}
	}

	threatlevelstotal := make([]int, realThreatNum)
	//var threatsides [][]int
	distancestotal := make([]float32, realThreatNum)
	closingspeedstotal := make([]float32, realThreatNum)
	threatCounts := make([]int, realThreatNum)

	for _, pkt := range packets {
		for i := 0; i < realThreatNum; i++ {
			if pkt.Threats[i].ThreatLevel > 0 {
				threatCounts[i]++
				threatlevelstotal[i] += pkt.Threats[i].ThreatLevel
				distancestotal[i] += pkt.Threats[i].Distance
				closingspeedstotal[i] += pkt.Threats[i].ClosingSpeed
			}
		}
	}

	avgPacket := RadarPkt{}
	for i := 0; i < realThreatNum; i++ {
		avgPacket.Threats[i].ThreatLevel = threatlevelstotal[i] / threatCounts[i]
		avgPacket.Threats[i].Distance = distancestotal[i] / float32(threatCounts[i])
		avgPacket.Threats[i].ClosingSpeed = closingspeedstotal[i] / float32(threatCounts[i])
	}

	return avgPacket
}
