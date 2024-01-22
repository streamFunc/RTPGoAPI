// Copyright (C) 2011 Werner Dittmann
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
//
// Authors: Werner Dittmann <Werner.Dittmann@t-online.de>
//

package rtp

import "sync"

// For full reference of registered RTP parameters and payload types refer to:
// http://www.iana.org/assignments/rtp-parameters

// Registry:
// PT        encoding name   audio/video (A/V)  clock rate (Hz)  channels (audio)  Reference
// --------  --------------  -----------------  ---------------  ----------------  ---------
// 0         PCMU            A                  8000             1                 [RFC3551]
// 1         Reserved
// 2         Reserved
// 3         GSM             A                  8000             1                 [RFC3551]
// 4         G723            A                  8000             1                 [Kumar][RFC3551]
// 5         DVI4            A                  8000             1                 [RFC3551]
// 6         DVI4            A                  16000            1                 [RFC3551]
// 7         LPC             A                  8000             1                 [RFC3551]
// 8         PCMA            A                  8000             1                 [RFC3551]
// 9         G722            A                  8000             1                 [RFC3551]
// 10        L16             A                  44100            2                 [RFC3551]
// 11        L16             A                  44100            1                 [RFC3551]
// 12        QCELP           A                  8000             1                 [RFC3551]
// 13        CN              A                  8000             1                 [RFC3389]
// 14        MPA             A                  90000                              [RFC3551][RFC2250]
// 15        G728            A                  8000             1                 [RFC3551]
// 16        DVI4            A                  11025            1                 [DiPol]
// 17        DVI4            A                  22050            1                 [DiPol]
// 18        G729            A                  8000             1                 [RFC3551]
// 19        Reserved        A
// 20        Unassigned      A
// 21        Unassigned      A
// 22        Unassigned      A
// 23        Unassigned      A
// 24        Unassigned      V
// 25        CelB            V                  90000                              [RFC2029]
// 26        JPEG            V                  90000                              [RFC2435]
// 27        Unassigned      V
// 28        nv              V                  90000                              [RFC3551]
// 29        Unassigned      V
// 30        Unassigned      V
// 31        H261            V                  90000                              [RFC4587]
// 32        MPV             V                  90000                              [RFC2250]
// 33        MP2T            AV                 90000                              [RFC2250]
// 34        H263            V                  90000                              [Zhu]
// 35-71     Unassigned      ?
// 72-76     Reserved for RTCP conflict avoidance                                  [RFC3551]
// 77-95     Unassigned      ?
// 96-127    dynamic         ?                                                     [RFC3551]

const (
	Audio = 1
	Video = 2
)

// AVProfile holds RTP payload profiles.
//
// The global variable PayloadFormatMap holds the well known payload formats
// (see http://www.iana.org/assignments/rtp-parameters).
// Applications shall not alter these predefined formats.
//
// If an application needs additional payload formats it must create and populate
// PayloadFormat structures and insert them into PayloadFormatMap before setting
// up the RTP communication. The index (key) into the map must be the payload
// format number. For dynamic payload formats applications shall use payload
// format numbers between 96 and 127 only.
//
// dynamic profiles have TypeNumber 255, which is reset in stream's SetProfile method
var initOnceT sync.Once

type AVProfile struct {
	TypeNumber uint8

	MediaType,
	ClockRate,
	Channels int

	MimeType,
	ProfileName string
}
type profileDbType map[string]*AVProfile
type profileIndexType map[uint8]*AVProfile

var avProfileDb = make(profileDbType)       // query by mime or profile name
var avProfileIndex = make(profileIndexType) // query by type number, only for non-dynamic profiles

func (profile *AVProfile) Name() string {
	if profile.ProfileName != "" {
		return profile.ProfileName
	} else {
		return profile.MimeType
	}
}

func initConfigOnce() {
	initOnceT.Do(initPayLoad)
}

func initPayLoad() {
	GlobalCRtpSessionMap = make(map[*CRtpSessionContext]*Session)
	profiles := []*AVProfile{
		{0, Audio, 8000, 1, "PCMU", ""},
		{3, Audio, 8000, 1, "GSM", ""},
		{4, Audio, 8000, 1, "G723", ""},
		{5, Audio, 8000, 1, "DVI4", ""},
		{7, Audio, 8000, 1, "LPC", ""},
		{8, Audio, 8000, 1, "PCMA", ""},
		{9, Audio, 8000, 1, "G722", ""},
		{10, Audio, 44100, 2, "L16", "L16-stereo"},
		{11, Audio, 44100, 1, "L16", "L16-mono"},
		{12, Audio, 8000, 1, "QCELP", ""},
		{13, Audio, 8000, 1, "CN", ""},
		{14, Audio, 90000, 0, "MPA", ""},
		{15, Audio, 8000, 1, "G728", ""},
		{16, Audio, 11025, 1, "DVI4", "DVI4-11K"},
		{17, Audio, 22050, 1, "DVI4", "DVI4-22K"},
		{18, Audio, 8000, 1, "G729", ""},
		{25, Video, 90000, 0, "CelB", ""},
		{26, Video, 90000, 0, "JPEG", ""},
		{28, Video, 90000, 0, "nv", ""},
		{31, Video, 90000, 0, "H261", ""},
		{32, Video, 90000, 0, "MPV", ""},
		{33, Audio | Video, 90000, 0, "MP2T", ""},
		{34, Video, 90000, 0, "H263", ""},

		// dynamic profiles
		{255, Audio, 8000, 1, "AMR", ""},
		{255, Audio, 16000, 1, "AMR-WB", ""},
		{255, Audio, 24400, 1, "EVS", ""},
		{255, Audio, 16000, 1, "TELEPHONE-EVENT", ""},
		{255, Video, 90000, 0, "H264", ""},
		{255, Video, 90000, 0, "H265", ""},
	}

	for _, profile := range profiles {
		name := profile.Name()
		avProfileDb[name] = profile
		if profile.TypeNumber >= 0 {
			avProfileIndex[profile.TypeNumber] = profile
		}

	}
}
