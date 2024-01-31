package rtp

// RTCP packet types
const (
	RtcpSR    = 200 // SR         sender report          [RFC3550]
	RtcpRR    = 201 // RR         receiver report        [RFC3550]
	RtcpSdes  = 202 // SDES       source description     [RFC3550]
	RtcpBye   = 203 // BYE        goodbye                [RFC3550]
	RtcpApp   = 204 // APP        application-defined    [RFC3550]
	RtcpRtpfb = 205 // RTPFB      Generic RTP Feedback   [RFC4585]
	RtcpPsfb  = 206 // PSFB       Payload-specific       [RFC4585]
	RtcpXr    = 207 // XR         extended report        [RFC3611]
	unKnown   = 208
)

// RTCP SDES item types
const (
	SdesEnd       = iota // END          end of SDES list                    [RFC3550]
	SdesCname            // CNAME        canonical name                      [RFC3550]
	SdesName             // NAME         user name                           [RFC3550]
	SdesEmail            // EMAIL        user's electronic mail address      [RFC3550]
	SdesPhone            // PHONE        user's phone number                 [RFC3550]
	SdesLoc              // LOC          geographic user location            [RFC3550]
	SdesTool             // TOOL         name of application or tool         [RFC3550]
	SdesNote             // NOTE         notice about the source             [RFC3550]
	SdesPriv             // PRIV         private extensions                  [RFC3550]
	SdesH323Caddr        // H323-CADDR   H.323 callable address              [Kumar]
	sdesMax
)

type CtrlEvent struct {
	EventType int    // Either a Stream event or a Rtcp* packet type event, e.g. RtcpSR, RtcpRR, RtcpSdes, RtcpBye
	Ssrc      uint32 // the input stream's SSRC
	Index     uint32 // and its index
	Reason    string // Resaon string if it was available, empty otherwise
}

type CtrlEventChan chan []*CtrlEvent
