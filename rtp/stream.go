package rtp

import "crypto/rand"

type SsrcStream struct {
	sequenceNumber    uint16
	ssrc              uint32
	profile           *AVProfile
	payloadTypeNumber uint8
	initialStamp      uint32
}

func (ss *SsrcStream) SetProfile(profileName string, dynamicTypeNumber uint8) bool {
	var profile *AVProfile
	var ok bool
	if profile, ok = avProfileDb[profileName]; !ok {
		return false
	}
	profileCopy := *profile
	if profile.TypeNumber < 96 {
		ss.payloadTypeNumber = profile.TypeNumber
	} else {
		ss.payloadTypeNumber = dynamicTypeNumber
	}
	ss.profile = &profileCopy
	ss.profile.TypeNumber = ss.payloadTypeNumber
	return true
}

// newInitialTimestamp creates a random initiali timestamp for outgoing packets
func (ss *SsrcStream) newInitialTimestamp() {
	var randBuf [4]byte
	rand.Read(randBuf[:])
	tmp := uint32(randBuf[0])
	tmp |= uint32(randBuf[1]) << 8
	tmp |= uint32(randBuf[2]) << 16
	tmp |= uint32(randBuf[3]) << 24
	ss.initialStamp = (tmp & 0xFFFFFFF)
}

type streamOutMap map[uint32]*SsrcStream

type Error string

func (s Error) Error() string {
	return string(s)
}
