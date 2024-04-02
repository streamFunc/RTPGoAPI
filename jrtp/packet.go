package jrtp

type DataPacket struct {
	payload     []byte // rtp payload
	rawBuffer   []byte // rtp payload + rtp header
	payloadType uint8
	pts         uint32 // presentation timestamp
	marker      bool   // should mark-bit in rtp header be set?
	ssrc        uint32
	csrc        []uint32
	seq         uint16
}

func (rp *DataPacket) Payload() []byte {
	return rp.payload
}

func (rp *DataPacket) PayloadType() uint8 {
	return rp.payloadType
}

func (rp *DataPacket) Buffer() []byte {
	return rp.rawBuffer
}

func (rp *DataPacket) InUse() int {
	return len(rp.rawBuffer)
}

func (rp *DataPacket) Timestamp() uint32 {
	return rp.pts
}

func (rp *DataPacket) Marker() bool {
	return rp.marker
}

func (rp *DataPacket) Ssrc() uint32 {
	return rp.ssrc
}

func (rp *DataPacket) CsrcList() []uint32 {
	return rp.csrc
}

func (rp *DataPacket) Seq() uint16 {
	return rp.seq
}

func (rp *DataPacket) SetMarker(m bool) {
	rp.marker = m
}

func (rp *DataPacket) SetPayload(payload []byte) {
	rp.payload = payload
}

func (rp *DataPacket) SetPayloadType(pt byte) {
	rp.payloadType = pt
}
