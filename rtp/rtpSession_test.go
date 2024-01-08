package rtp

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

var gStopFlag atomic.Bool

func registerSignal() {
	sig := make(chan os.Signal, 8)
	signal.Notify(sig,
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGSEGV,
		syscall.SIGABRT,
	)
	go onSignal(sig)
}

func onSignal(sig chan os.Signal) {
	for {
		s := <-sig
		fmt.Println("receive signal:", s.String())
		gStopFlag.Store(true)
	}
}

func TestRtp(t *testing.T) {
	registerSignal()

	var tpLocal *TransportUDP
	var local, _ = net.ResolveIPAddr("ip", "127.0.0.1")
	var remote, _ = net.ResolveIPAddr("ip", "127.0.0.1")

	tpLocal, _ = NewTransportUDP(local, 11000, "")

	r := NewSession(tpLocal, tpLocal)

	strLocalIdx, _ := r.NewSsrcStreamOut(&Address{
		IPAddr:   local.IP,
		DataPort: 11000,
		CtrlPort: 1 + 11000,
		Zone:     "",
	}, 0, 0)

	fmt.Printf("strLocalIdx:%v\n", strLocalIdx)

	ok := r.SsrcStreamOutForIndex(strLocalIdx).SetProfile("AMR", byte(123))
	if ok {
		fmt.Printf("SsrcStreamOutForIndex success\n")
	}

	r.AddRemote(&Address{
		IPAddr:   remote.IP,
		DataPort: 11002,
		CtrlPort: 1 + 11002,
		Zone:     "",
	})

	if err := r.StartSession(); err != nil {
		fmt.Println("error:", err)
		return
	}

	go startRtpReceiveLoop(r)

	fmt.Printf("seq:%v pt:%v\n", r.GetSequenceNumber(), r.GetPayloadType())

	payLoad := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9}

	rp := r.NewDataPacket(0)
	rp.SetPayload(payLoad)
	rp.SetPayloadType(123)
	rp.SetMarker(true)

	if err := r.WriteData(rp); err != nil {
		fmt.Println("send fail,error:", err)
	} else {
		fmt.Printf("send data success\n")
	}

	time.Sleep(time.Second * 30)
	gStopFlag.Store(true)

	r.CloseSession()

}

func startRtpReceiveLoop(rp *Session) {
	for {
		if !rp.startFlag {
			break
		}
		select {
		case pl, more := <-rp.HandleC:
			if !more {
				return
			}
			fmt.Printf("startRtpReceiveLoop seq:%v payload:%v PayloadType:%v marker:%v", pl.seq, len(pl.payload), pl.payloadType, pl.marker)
			return
		}
	}
}
