package jrtp

import (
	"encoding/binary"
	"fmt"
)

const (
	// The NAL unit type octet has the following format:
	//+---------------+
	//|0|1|2|3|4|5|6|7|
	//+-+-+-+-+-+-+-+-+
	//|F|NRI| TypeId    |
	//+---------------+

	CNalTypeStapa uint8 = 24
	CNalTypeFua   uint8 = 28
	CNalTypeSei   uint8 = 6
	CNalTypeSps   uint8 = 7
	CNalTypePps   uint8 = 8
	CNalTypeAu    uint8 = 9 // Access Unit Delimiter

	CBitmaskNalType uint8 = 0x1f
	CBitmaskRefIdc  uint8 = 0x60
	CBitmaskFuStart uint8 = 0x80
	CBitmaskFuEnd   uint8 = 0x40

	CDefaultMtu = 1400
)

type H264Packet struct {
	Payload []byte
	Pts     int64
}

// CPacketListFromH264Mode
// mtu is for payload, not including ip,udp headers
// nal of type stapA would not be created if disableStap set true
func CPacketListFromH264Mode(annexbPayload []byte, pts uint32, payloadType uint8, mtu int, disableStap bool) (pl *CRtpPacketList) {
	// packetization-mode == 1
	// Only single NAL unit packets, STAP-As, and FU-As MAY be used in this mode.
	nals := CExtractNals(annexbPayload)
	var bufferedNals [][]byte
	var rtpPayloadArray [][]byte
	bufferedSize := 1 // stapA header with 1 byte
	i := 0
	for i < len(nals) {
		nal := nals[i]
		size := len(nal)
		if disableStap {
			goto noStap
		}
		if size+2+bufferedSize <= mtu {
			// nal size with 2 bytes in stapA
			bufferedNals = append(bufferedNals, nal)
			bufferedSize += size + 2
			i++
			continue
		} else {
			// this nal can not be aggregated, just flush buffered nals if any
			if len(bufferedNals) > 0 {
				if len(bufferedNals) == 1 {
					// single nal, no aggregation
					rtpPayloadArray = append(rtpPayloadArray, bufferedNals[0])
				} else {
					stapA := makeStapA(bufferedNals...)
					rtpPayloadArray = append(rtpPayloadArray, stapA)
				}
				bufferedNals = nil
				bufferedSize = 1
			}
		}

	noStap:
		// check this nal again
		if size > mtu {
			rtpPayload := makeFuA(mtu, nal)
			rtpPayloadArray = append(rtpPayloadArray, rtpPayload...)
			i++
		} else if !disableStap && size+2+bufferedSize < mtu {
			// size < mtu - (2 + bufferedSize)
			// if this nal can be put into stapA after buffer flushed, do it again
			continue
		} else {
			// mtu - (2 + bufferedSize) <= size <= mtu
			// rare case, just send as it is
			rtpPayloadArray = append(rtpPayloadArray, nal)
			i++
		}

	}

	// check if buffered nals exist for the last time
	if !disableStap && len(bufferedNals) > 0 {
		if len(bufferedNals) == 1 {
			// single nal, no aggregation
			rtpPayloadArray = append(rtpPayloadArray, bufferedNals[0])
		} else {
			stapA := makeStapA(bufferedNals...)
			rtpPayloadArray = append(rtpPayloadArray, stapA)
		}
	}

	// all payloads are in order, make the packet list
	makePacketList(&pl, rtpPayloadArray, pts, payloadType)
	return
}

func makePacketList(pl **CRtpPacketList, rtpPayload [][]byte, pts uint32, payloadType uint8) {
	var packet *CRtpPacketList
	prev := *pl
	for _, payload := range rtpPayload {
		packet = &CRtpPacketList{
			Payload:     payload,
			Pts:         pts,
			PayloadType: payloadType,
		}
		if prev == nil {
			*pl = packet
		} else {
			prev.SetNext(packet)
		}
		prev = packet
	}
	if packet != nil {
		// set last packet mark bit if they are in the same access unit

		// TODO:
		//For aggregation packets (STAP and MTAP), the marker bit in the RTP
		//header MUST be set to the value that the marker bit of the last
		//NAL unit of the aggregation packet would have been if it were
		//transported in its own RTP packet
		packet.Marker = true
	}
}

// CExtractNals splits annexb payload by start code
func CExtractNals(annexbPayload []byte) (nals [][]byte) {
	// start code can be of 4 bytes: 0x00,0x00,0x00,0x01 (sps,pps,first slice)
	// or 3 bytes: 0x00,0x00,0x01
	zeros := 0
	prevStart := 0
	totalLen := len(annexbPayload)
	for i, b := range annexbPayload {
		switch b {
		case 0x00:
			zeros++
			continue
		case 0x01:
			if zeros == 2 || zeros == 3 {
				// found a start code
				if i-zeros > prevStart {
					nal := annexbPayload[prevStart : i-zeros]
					nals = append(nals, nal)
				}
				prevStart = i + 1
				if prevStart >= totalLen {
					return
				}
			}
		}
		zeros = 0
	}
	if totalLen > prevStart {
		nals = append(nals, annexbPayload[prevStart:])
	}
	return
}

func CPrintNal(nal []byte) {
	nalType := nal[0] & CBitmaskNalType
	nalRefIdc := nal[0] & CBitmaskRefIdc
	switch nalType {
	case CNalTypeSps:
		fmt.Printf("type sps")
	case CNalTypePps:
		fmt.Printf("type pps")
	case CNalTypeSei:
		fmt.Printf("type sei")
	case CNalTypeFua:
		fmt.Printf("type fua")
	case CNalTypeStapa:
		fmt.Printf("type stapa")
	default:
		fmt.Printf("type: %d", nalType)
	}
	fmt.Printf("nal len is %d,refIdc is %d", len(nal), nalRefIdc)
}

func makeStapA(nals ...[]byte) (rtpPayload []byte) {
	// 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	//+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	//|                     RTP Header                                |
	//+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	//| STAP-A NAL HDR|        NALU 1 Size            |  NALU 1 HDR   |
	//+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	//|             NALU 1 Data                                       |
	//:                                                               :
	//+               +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	//|               |     NALU 2 Size               | NALU 2 HDR    |
	//+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	//|                        NALU 2 Data                            |
	//:                                                               :
	//|                               +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	//|                               :...OPTIONAL RTP padding        |
	//+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	//
	// The value of NRI MUST be the maximum of all the NAL units carried
	// in the aggregation packet.
	rtpPayload = []byte{0x00} // header placeholder, set it later
	size := make([]byte, 2)
	var maxNri uint8
	for _, nal := range nals {
		binary.BigEndian.PutUint16(size, uint16(len(nal)))
		rtpPayload = append(rtpPayload, size...)
		rtpPayload = append(rtpPayload, nal...)
		nri := nal[0] & CBitmaskRefIdc
		if maxNri < nri {
			maxNri = nri
		}
	}
	rtpPayload[0] = maxNri | CNalTypeStapa
	return
}

func makeFuA(mtu int, nal []byte) (rtpPayload [][]byte) {
	// 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	//+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	//| FU indicator |  FU header     |                               |
	//+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+                               |
	//|                                                               |
	//|                    FU payload                                 |
	//|                                                               |
	//|                               +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	//|                               :...OPTIONAL RTP padding        |
	//+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	//
	// The FU header has the following format:
	//+---------------+
	//|0|1|2|3|4|5|6|7|
	//+-+-+-+-+-+-+-+-+
	//|S|E|R| TypeId    |
	//+---------------+
	nri := nal[0] & CBitmaskRefIdc
	nalType := nal[0] & CBitmaskNalType
	indicator := nri | CNalTypeFua
	isFirstFragment := true
	startPtr := 1 // skip the nal header
	remainingPayloadSize := len(nal) - startPtr
	maxPayloadSize := mtu - 2 // 2 = fu indicator + fu header

	for remainingPayloadSize > 0 {
		fragmentPayloadSize := remainingPayloadSize
		if fragmentPayloadSize > maxPayloadSize {
			fragmentPayloadSize = maxPayloadSize
		}
		payload := make([]byte, fragmentPayloadSize+2)
		payload[0] = indicator
		header := nalType
		if isFirstFragment {
			header |= CBitmaskFuStart // set start bit
			isFirstFragment = false
		} else if fragmentPayloadSize == remainingPayloadSize {
			header |= CBitmaskFuEnd // set end bit
		}
		payload[1] = header
		copy(payload[2:], nal[startPtr:startPtr+fragmentPayloadSize])
		rtpPayload = append(rtpPayload, payload)
		startPtr += fragmentPayloadSize
		remainingPayloadSize -= fragmentPayloadSize
	}
	return
}
