package rtp

type SsrcStream struct {
	sequenceNumber    uint16
	ssrc              uint32
	profile           *AVProfile
	payloadTypeNumber uint8
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

type streamOutMap map[uint32]*SsrcStream

type Error string

func (s Error) Error() string {
	return string(s)
}
