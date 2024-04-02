package jrtp

// CRtpPacketList is either received RTP data packet or generated packets by codecs that can be readily put to
// stack for transmission. audio data is usually one packet at a time as no pts is required, but video codecs can
// build multiple packets of the same pts. those packets can be linked and send to rtp stack as a whole.
type CRtpPacketList struct {
	Payload     []byte // rtp payload
	RawBuffer   []byte // rtp payload + rtp header
	PayloadType uint8
	Pts         uint32 // presentation timestamp
	PrevPts     uint32 // previous packet's pts
	Marker      bool   // should mark-bit in rtp header be set?
	Ssrc        uint32
	Csrc        []uint32

	next *CRtpPacketList // more RtpPacketList, if any
}

func CNewPacketListFromRtpPacket(packet *DataPacket) *CRtpPacketList {
	if packet.InUse() <= 0 || packet.Buffer() == nil {
		return nil
	}
	return &CRtpPacketList{
		Payload:     packet.Payload(),
		RawBuffer:   packet.Buffer()[:packet.InUse()],
		PayloadType: packet.PayloadType(),
		Pts:         packet.Timestamp(),
		Marker:      packet.Marker(),
		Ssrc:        packet.Ssrc(),
		Csrc:        packet.CsrcList(),
	}
}

func (pl *CRtpPacketList) Iterate(f func(p *CRtpPacketList)) {
	ppl := pl
	for ppl != nil {
		f(ppl)
		ppl = ppl.next
	}
}

func (pl CRtpPacketList) CloneSingle() *CRtpPacketList {
	return &CRtpPacketList{
		Payload:     pl.Payload,
		RawBuffer:   pl.RawBuffer,
		PayloadType: pl.PayloadType,
		Pts:         pl.Pts,
		Marker:      pl.Marker,
		Ssrc:        pl.Ssrc,
		Csrc:        pl.Csrc,
	}
}

func (pl *CRtpPacketList) Clone() *CRtpPacketList {
	var cloned, current *CRtpPacketList
	pl.Iterate(func(packet *CRtpPacketList) {
		newPacket := packet.CloneSingle()
		if cloned == nil {
			cloned = newPacket
		} else {
			current.next = newPacket
		}

		current = newPacket
	})
	return cloned
}

func (pl *CRtpPacketList) Next() *CRtpPacketList {
	return pl.next
}

func (pl *CRtpPacketList) SetNext(npl *CRtpPacketList) {
	pl.next = npl
}

func (pl *CRtpPacketList) GetLast() *CRtpPacketList {
	ppl := pl
	for ppl.next != nil {
		ppl = ppl.next
	}
	return ppl
}

func (pl *CRtpPacketList) Len() (length int) {
	pl.Iterate(func(ppl *CRtpPacketList) {
		length++
	})
	return
}
