package rtp

import (
	"net"
)

type TransportRecv interface {
}

type TransportWrite interface {
}

type TransportUDP struct {
	mode                        int
	clockRate                   int
	payloadType                 int
	localAddrRtp, localAddrRtcp *net.UDPAddr
}

func NewTransportUDP(addr *net.IPAddr, port int, zone string) (*TransportUDP, error) {
	tp := new(TransportUDP)
	tp.mode = 1
	tp.clockRate = 90000
	tp.payloadType = 123
	tp.localAddrRtp = &net.UDPAddr{IP: addr.IP, Port: port, Zone: zone}
	tp.localAddrRtcp = &net.UDPAddr{IP: addr.IP, Port: port + 1, Zone: zone}
	return tp, nil
}

func (n *TransportUDP) SetRtpMode(mode int) {
	n.mode = mode
}

func (n *TransportUDP) SetClockRate(clockRate int) {
	n.clockRate = clockRate
}

func (n *TransportUDP) SetPayLoadType(pt int) {
	n.payloadType = pt
}
