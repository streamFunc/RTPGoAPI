package rtp

import "C"
import (
	"errors"
	"fmt"
	"unsafe"
)

type CRtpSession struct {
	LocalIp, RemoteIp            string
	LocalPort, RemotePort        int
	PayloadType, ClockRate, Mode int
	startFlag                    bool
	ctx                          *CRtpSessionContext
	ini                          *CRtpSessionInitData
}

// NewRtpSession mode 0 ORTP 1 JRTP
func NewRtpSession(localIp, remoteIp string, mode, localPort, remotePort, payloadType, clockRate int) *CRtpSession {
	dec := CRtpSession{
		startFlag:   false,
		LocalIp:     localIp,
		RemoteIp:    remoteIp,
		LocalPort:   localPort,
		RemotePort:  remotePort,
		PayloadType: payloadType,
		ClockRate:   clockRate,
		Mode:        mode,
	}
	dec.ctx = newSessionContext(dec.Mode)
	dec.ini = creatRtpInitData(localIp, remoteIp, localPort, remotePort, payloadType, clockRate)

	if dec.ctx == nil || dec.ini == nil {
		fmt.Printf("NewRtpSession error creat dec.ctx or dec.ini fail\n")
		return nil
	}

	fmt.Printf("localIp:%v port:%v remoteIp:%v port:%v payloadType:%v\n", localIp, localPort, remoteIp, remotePort, payloadType)

	return &dec
}

func (n *CRtpSession) InitSession() error {
	if n.ctx != nil {
		res := n.ctx.initRtpSession(n.ini)
		if res == false {
			fmt.Printf("InitSession fail\n")
			return errors.New(fmt.Sprintf("InitSession  fail"))
		} else {
			fmt.Printf("InitSession success\n")
		}
	} else {
		return errors.New(fmt.Sprintf("ctx nil InitSession  fail"))
	}
	return nil
}

func (n *CRtpSession) StartSession() error {
	if n.ctx != nil {
		res := n.ctx.startRtpSession()
		n.startFlag = true
		if res == false {
			fmt.Printf("StartSession fail,error:%v", res)
			return errors.New(fmt.Sprintf("StartSession  fail"))
		} else {
			fmt.Printf("StartSession success\n")
		}
	} else {
		fmt.Printf("StartSession fail,ctx nil\n")
		return errors.New(fmt.Sprintf("ctx nil StartSession fail"))
	}
	return nil
}

func (n *CRtpSession) StopSession() error {
	if n.ctx != nil {
		res := n.ctx.stopRtpSession()
		n.startFlag = false
		if res == false {
			fmt.Printf("StopSession fail,error:%v\n", res)
			return errors.New(fmt.Sprintf("StopSession fail"))
		} else {
			fmt.Printf("StopSession success\n")
		}
	} else {
		return errors.New(fmt.Sprintf("ctx nil StopSession fail"))
	}
	return nil
}

func (n *CRtpSession) SendData(payload []byte, len, marker int) error {
	if n.ctx != nil && n.startFlag {
		n.ctx.sendDataRtpSession(payload, len, marker)
	} else {
		return errors.New(fmt.Sprintf("ctx nil SendData fail"))
	}
	return nil
}

// ReceiveData for all time you can allow it until stop
func (n *CRtpSession) ReceiveData(buffer []byte, len int) error {
	if n.ctx != nil && n.startFlag {
		n.ctx.rcvDataRtpSession(buffer, len, unsafe.Pointer(n.ctx))
	} else {
		return errors.New(fmt.Sprintf("ctx nil ReceiveData fail"))
	}
	return nil
}

func (n *CRtpSession) Destroy() {
	if n.ctx != nil {
		n.ini.destroySessionInitData()
		n.ctx.destroyRtpSession()
	}
}

func (n *CRtpSession) GetTimeStamp() int {
	if n.ctx != nil {
		return n.ctx.getTimeStamp()
	} else {
		return -1
	}
}

func (n *CRtpSession) GetSequenceNumber() int {
	if n.ctx != nil {
		return n.ctx.getSequenceNumber()
	} else {
		return -1
	}
}

func (n *CRtpSession) GetSsrc() int {
	if n.ctx != nil {
		return n.ctx.getSsrc()
	} else {
		return -1
	}
}

func (n *CRtpSession) GetCSrc() []uint32 {
	if n.ctx != nil {
		return n.ctx.getCSrc()
	} else {
		return nil
	}
}

func (n *CRtpSession) GetPayloadType() int {
	if n.ctx != nil {
		return n.ctx.getPayloadType()
	} else {
		return -1
	}
}

func (n *CRtpSession) GetMarker() bool {
	if n.ctx != nil {
		return n.ctx.getMarker()
	} else {
		return false
	}
}

func (n *CRtpSession) GetVersion() int {
	if n.ctx != nil {
		return n.ctx.getVersion()
	} else {
		return -1
	}
}

func (n *CRtpSession) GetPadding() bool {
	if n.ctx != nil {
		return n.ctx.getPadding()
	} else {
		return false
	}
}

func (n *CRtpSession) GetExtension() bool {
	if n.ctx != nil {
		return n.ctx.getExtension()
	} else {
		return false
	}
}

func (n *CRtpSession) GetCC() int {
	if n.ctx != nil {
		return n.ctx.getCC()
	} else {
		return -1
	}
}
