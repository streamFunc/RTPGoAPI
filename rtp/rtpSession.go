package rtp

import "C"
import (
	"errors"
	"fmt"
	"os"
	"unsafe"
)

var GlobalCRtpSessionMap map[*CRtpSessionContext]*CRtpSession

type CRtpSession struct {
	LocalIp, RemoteIp            string
	LocalPort, RemotePort        int
	PayloadType, ClockRate, Mode int
	startFlag                    bool
	frameBufData                 []byte
	HandleC                      chan *RPacket
	fp                           *os.File
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
	dec.HandleC = make(chan *RPacket, 120)

	if dec.ctx == nil || dec.ini == nil {
		fmt.Printf("NewRtpSession error creat dec.ctx or dec.ini fail\n")
		return nil
	}

	GlobalCRtpSessionMap = make(map[*CRtpSessionContext]*CRtpSession)
	GlobalCRtpSessionMap[dec.ctx] = &dec

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
		n.fp, _ = os.Create("myReceived.264")
		if n.fp == nil {
			fmt.Printf("error creat myReceived.264")
		}
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
		delete(GlobalCRtpSessionMap, n.ctx)
		n.startFlag = false
		if res == false {
			fmt.Printf("StopSession fail,error:%v\n", res)
			return errors.New(fmt.Sprintf("StopSession fail"))
		} else {
			if n.fp != nil {
				n.fp.Close()
				n.fp = nil
			}
			fmt.Printf("StopSession success\n")
		}
		if n.HandleC != nil {
			// notify packet handler
			close(n.HandleC)
			n.HandleC = nil
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

func (n *CRtpSession) HandleCallBackData(data []byte, marker bool) {
	n.unPackRTP2H264(data, marker)

}

func (n *CRtpSession) Destroy() {
	if n.ctx != nil {
		n.ini.destroySessionInitData()
		n.ctx.destroyRtpSession()
		if n.fp != nil {
			n.fp.Close()
			n.fp = nil
		}
	}
}

func (n *CRtpSession) GetTimeStamp() uint32 {
	if n.ctx != nil {
		return n.ctx.getTimeStamp()
	} else {
		return 0
	}
}

func (n *CRtpSession) GetSequenceNumber() uint16 {
	if n.ctx != nil {
		return n.ctx.getSequenceNumber()
	} else {
		return 0
	}
}

func (n *CRtpSession) GetSsrc() uint32 {
	if n.ctx != nil {
		return n.ctx.getSsrc()
	} else {
		return 0
	}
}

func (n *CRtpSession) GetCSrc() []uint32 {
	if n.ctx != nil {
		return n.ctx.getCSrc()
	} else {
		return nil
	}
}

func (n *CRtpSession) GetPayloadType() uint16 {
	if n.ctx != nil {
		return n.ctx.getPayloadType()
	} else {
		return 0
	}
}

func (n *CRtpSession) GetMarker() bool {
	if n.ctx != nil {
		return n.ctx.getMarker()
	} else {
		return false
	}
}

func (n *CRtpSession) GetVersion() uint8 {
	if n.ctx != nil {
		return n.ctx.getVersion()
	} else {
		return 0
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

func (n *CRtpSession) GetCC() uint8 {
	if n.ctx != nil {
		return n.ctx.getCC()
	} else {
		return 0
	}
}

func (n *CRtpSession) receiveRtpCache(pl *RPacket) {
	n.HandleC <- pl
}

func (n *CRtpSession) unPackRTP2H264(rtpPayload []byte, marker bool) {
	fuIndicator := rtpPayload[0]                       //获取第一个字节 分片单元指识
	fuHeader := rtpPayload[1]                          //获取第二个字节 分片单元头
	naluType := fuIndicator & 0x1f                     //获取FU indicator的类型域
	flag := fuHeader & 0xe0                            //获取FU header的前三位，判断当前是分包的开始、中间或结束
	nalFua := (fuIndicator & 0xe0) | (fuHeader & 0x1f) //FU_A nal

	var FrameType string
	if naluType == 28 {
		if nalFua == 0x67 {
			FrameType = "SPS"
		} else if nalFua == 0x68 {
			FrameType = "PPS"
		} else if nalFua == 0x65 {
			FrameType = "IDR"
		} else if nalFua == 0x61 {
			FrameType = "B Frame"
		} else if nalFua == 0x41 {
			FrameType = "P Frame"
		} else if nalFua == 0x06 {
			FrameType = "sei"
		} else {
			FrameType = "other"
		}
	} else if naluType <= 23 {
		if naluType == 7 {
			FrameType = "SPS"
		} else if naluType == 8 {
			FrameType = "PPS"
		} else if naluType == 5 {
			FrameType = "IDR"
		} else if naluType == 6 {
			FrameType = "SEI"
		} else if naluType == 1 {
			FrameType = "P Frame"
		} else {
			FrameType = "B Frame"
		}
	}

	if naluType == 28 { //判断NAL的类型为0x1c=28，FU-A分片
		if flag == 0x80 { //first slice
			fmt.Printf("H264Rtp slice is first  fu-a %v\n", len(rtpPayload))
			o := make([]byte, len(rtpPayload)+4-2)
			o[0] = 0x00
			o[1] = 0x00
			o[2] = 0x01
			o[3] = nalFua
			copy(o[4:], rtpPayload[2:])
			n.frameBufData = append(n.frameBufData, o...)

		} else if flag == 0x40 { //last slice
			o := make([]byte, len(rtpPayload)-2)
			copy(o[0:], rtpPayload[2:])
			n.frameBufData = append(n.frameBufData, o...)
			if marker {
				if n.fp != nil {
					n.fp.Write(n.frameBufData)
				}
				n.frameBufData = nil
			}
		} else {
			o := make([]byte, len(rtpPayload)-2)
			copy(o[0:], rtpPayload[2:])
			n.frameBufData = append(n.frameBufData, o...)
		}
	} else if naluType == 29 { //fu-b
		fmt.Printf("H264Rtp slice is fu-b,ignore")
	} else if naluType == 24 { //stap-a
		stapADataLen := len(rtpPayload) - 1
		stapAData := rtpPayload[1:]

		if stapADataLen < 2 {
			return
		}
		for true {
			if stapADataLen < 2 {
				break
			}
			nalSize := int((stapAData[0] << 8) | stapAData[1])
			if stapADataLen < nalSize+2 {
				break
			}
			frame := make([]byte, nalSize+3)
			frame[0] = 0x00
			frame[1] = 0x00
			frame[2] = 0x01
			copy(frame[3:], stapAData[2:nalSize+2])
			if n.fp != nil {
				fmt.Printf("H264Rtp stap-a length:%v\n", len(frame))
				n.fp.Write(frame)
			}

			stapAData = stapAData[nalSize+2:]
			stapADataLen = stapADataLen - 2 - nalSize
		}
	} else if naluType == 25 { //stap-b
		fmt.Printf("H264Rtp slice is stap-b,ignore")
	} else if naluType == 26 { //MTAP16
		fmt.Println("H264Rtp slice is MTAP16,ignore")
	} else if naluType == 27 { //MTAP24
		fmt.Printf("H264Rtp slice is MTAP24,ignore FrameType:%s", FrameType)
	} else { //单一NAL 单元模式
		if marker {
			frame := make([]byte, len(rtpPayload)+3)
			frame[0] = 0x00
			frame[1] = 0x00
			frame[2] = 0x01
			copy(frame[3:], rtpPayload[0:])
			n.frameBufData = append(n.frameBufData, frame...)
			if n.fp != nil {
				n.fp.Write(n.frameBufData)
			}
			n.frameBufData = nil
		} else {
			frame := make([]byte, len(rtpPayload)+3)
			frame[0] = 0x00
			frame[1] = 0x00
			frame[2] = 0x01
			copy(frame[3:], rtpPayload[0:])
			n.frameBufData = append(n.frameBufData, frame...)

		}

	}
}
