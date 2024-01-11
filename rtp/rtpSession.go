package rtp

import "C"
import (
	"errors"
	"fmt"
	"net"
	"time"
	"unsafe"
)

var GlobalCRtpSessionMap map[*CRtpSessionContext]*Session

const (
	dataReceiveChanLen = 3
	ctrlEventChanLen   = 3
)

const (
	maxNumberOutStreams = 5
	maxNumberInStreams  = 30
)

type Session struct {
	LocalIp, RemoteIp             string
	LocalPort, RemotePort         int
	localCtrlPort, RemoteCtrlPort int
	PayloadType, ClockRate, Mode  int
	streamOutIndex                uint32
	profile                       string
	startFlag                     bool
	frameBufData                  []byte
	streamsOut                    streamOutMap
	HandleC                       chan *DataPacket
	ctrlEventChan                 CtrlEventChan
	//fp                           *os.File
	ctx *CRtpSessionContext
	ini *CRtpSessionInitData
}

type Address struct {
	IPAddr             net.IP
	DataPort, CtrlPort int
	Zone               string
}

func NewSession(rp *TransportUDP, tv TransportRecv) *Session {
	dec := Session{
		startFlag:      false,
		LocalIp:        rp.localAddrRtp.IP.String(),
		LocalPort:      rp.localAddrRtp.Port,
		PayloadType:    rp.payloadType,
		ClockRate:      rp.clockRate,
		Mode:           rp.mode,
		streamOutIndex: 0,
	}
	dec.streamsOut = make(streamOutMap, maxNumberOutStreams)
	dec.ctx = newSessionContext(dec.Mode)
	dec.HandleC = make(chan *DataPacket, 120)

	initConfigOnce()

	if dec.ctx == nil {
		fmt.Printf("NewRtpSession error creat dec.ctx or dec.ini fail\n")
		return nil
	}

	GlobalCRtpSessionMap[dec.ctx] = &dec

	fmt.Printf("localIp:%v port:%v  payloadType:%v\n", dec.LocalIp, dec.LocalPort, dec.PayloadType)

	return &dec
}

func (n *Session) AddRemote(remote *Address) (index uint32, err error) {
	if (remote.DataPort & 0x1) == 0x1 {
		return 0, errors.New(fmt.Sprintf("RTP data port number is not an even number."))
	}
	n.RemoteIp = remote.IPAddr.String()
	n.RemotePort = remote.DataPort
	n.localCtrlPort = remote.DataPort + 1

	return
}

func (n *Session) NewDataPacket(stamp uint32) *DataPacket {
	rp := &DataPacket{
		pts: stamp,
	}
	return rp
}

func (n *Session) CreateDataReceiveChan() chan *DataPacket {
	return n.HandleC
}

func (n *Session) RemoveDataReceiveChan() {
	if n.HandleC != nil {
		close(n.HandleC)
	}
	n.HandleC = nil
}

func (n *Session) CreateCtrlEventChan() CtrlEventChan {
	n.ctrlEventChan = make(CtrlEventChan, ctrlEventChanLen)
	return n.ctrlEventChan
}

// RemoveCtrlEventChan deletes the control event channel.
func (n *Session) RemoveCtrlEventChan() {
	n.ctrlEventChan = nil
}

func (n *Session) StartSession() error {
	if n.ctx != nil {
		//	n.fp, _ = os.Create("myReceived.264")
		//if n.fp == nil {
		//	fmt.Printf("error creat myReceived.264")
		//}
		n.initSession()

		res := n.ctx.startRtpSession()
		n.startFlag = true
		if res == false {
			fmt.Printf("StartSession fail,error:%v", res)
			return errors.New(fmt.Sprintf("StartSession  fail"))
		} else {
			fmt.Printf("StartSession success\n")
			go n.receivePacketLoop()
		}
	} else {
		fmt.Printf("StartSession fail,ctx nil\n")
		return errors.New(fmt.Sprintf("ctx nil StartSession fail"))
	}
	return nil
}

func (n *Session) CloseSession() error {
	if n.ctx != nil && n.startFlag {
		n.startFlag = false
		res := n.ctx.stopRtpSession()
		delete(GlobalCRtpSessionMap, n.ctx)
		if res == false {
			fmt.Printf("StopSession fail,error:%v\n", res)
			return errors.New(fmt.Sprintf("StopSession fail"))
		} else {
			//if n.fp != nil {
			//n.fp.Close()
			//n.fp = nil
			//}
			fmt.Printf("StopSession success\n")
		}
		if n.HandleC != nil {
			// notify packet handler
			close(n.HandleC)
			n.HandleC = nil
		}
		n.destroy()
	} else {
		return errors.New(fmt.Sprintf("ctx nil StopSession fail"))
	}
	return nil
}

func (n *Session) WriteData(rp *DataPacket) (k int, err error) {
	if n.ctx != nil && n.startFlag {
		if rp.marker {
			n.ctx.sendDataRtpSession(rp.payload, len(rp.payload), 1)
		} else {
			n.ctx.sendDataRtpSession(rp.payload, len(rp.payload), 0)
		}

	} else {
		return 0, Error("WriteData fail.")
	}
	return 0, nil
}

func (n *Session) initSession() error {
	if n.ctx != nil {
		if len(n.streamsOut) > 0 {
			n.ClockRate = n.streamsOut[0].profile.ClockRate
			n.PayloadType = int(n.streamsOut[0].payloadTypeNumber)
			n.profile = n.streamsOut[0].profile.ProfileName
			fmt.Printf("InitSession ClockRate:%v  PayloadType:%v type:%v\n",
				n.streamsOut[0].profile.ClockRate, n.streamsOut[0].payloadTypeNumber, n.streamsOut[0].profile.MimeType)
		}

		n.ini = creatRtpInitData(n.LocalIp, n.RemoteIp, n.LocalPort, n.RemotePort, n.PayloadType, n.ClockRate)
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

func (n *Session) receiveData(buffer []byte, len int) error {
	if n.ctx != nil && n.startFlag {
		n.ctx.rcvDataRtpSession(buffer, len, unsafe.Pointer(n.ctx))
	} else {
		return errors.New(fmt.Sprintf("ctx nil ReceiveData fail"))
	}
	return nil
}

func (n *Session) receivePacketLoop() {
	var buffer []byte
	buffer = make([]byte, 1500)

	if n.ctx != nil {
		ticker := time.Tick(time.Millisecond)

		for range ticker {
			if !n.startFlag || n.ctx == nil {
				fmt.Printf("stop receivePacket...")
				return
			}
			n.receiveData(buffer, 1500)
		}
	}
}

func (n *Session) HandleCallBackData(data []byte, marker bool) {
	n.unPackRTP2H264(data, marker)

}

func (n *Session) destroy() {
	if n.ctx != nil {
		n.ini.destroySessionInitData()
		n.ctx.destroyRtpSession()
		n.ctx = nil
	}
}

func (n *Session) receiveRtpCache(pl *DataPacket) {
	n.HandleC <- pl
}

func (n *Session) unPackRTP2H264(rtpPayload []byte, marker bool) {
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
				//if n.fp != nil {
				//n.fp.Write(n.frameBufData)
				//}
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
			//	if n.fp != nil {
			//fmt.Printf("H264Rtp stap-a length:%v\n", len(frame))
			//n.fp.Write(frame)
			//}

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
			//if n.fp != nil {
			//	n.fp.Write(n.frameBufData)
			//}
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

// SsrcStreamOutForIndex not support now
func (n *Session) SsrcStreamOutForIndex(streamIndex uint32) *SsrcStream {
	return n.streamsOut[streamIndex]
}

func (n *Session) NewSsrcStreamOut(own *Address, ssrc uint32, sequenceNo uint16) (index uint32, err Error) {
	str := &SsrcStream{
		sequenceNumber: sequenceNo,
		ssrc:           ssrc,
	}
	n.streamsOut[n.streamOutIndex] = str
	index = n.streamOutIndex
	n.streamOutIndex++
	return index, ""
}
