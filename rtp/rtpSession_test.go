package rtp

import (
	"fmt"
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

	r := NewRtpSession("127.0.0.1", "127.0.0.1", 0, 11000, 11002, 96, 90000)
	if err := r.InitSession(); err != nil {
		fmt.Println("error:", err)
		return
	}
	if err := r.StartSession(); err != nil {
		fmt.Println("error:", err)
		return
	}

	fmt.Printf("seq:%v pt:%v\n", r.GetSequenceNumber(), r.GetPayloadType())

	const rcvLen int = 1024
	var rcvBuf []byte
	rcvBuf = make([]byte, 1024)
	go func() {
		for !gStopFlag.Load() {
			if err := r.ReceiveData(rcvBuf, rcvLen); err != nil {
				fmt.Println("received fail,error:", err)
			}
			time.Sleep(time.Millisecond)
		}
	}()

	payLoad := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9}

	if err := r.SendData(payLoad, 10, 0); err != nil {
		fmt.Println("send fail,error:", err)
	} else {
		fmt.Printf("send data success\n")
	}

	time.Sleep(time.Second * 30)
	gStopFlag.Store(true)

	r.StopSession()
	r.Destroy()

}
