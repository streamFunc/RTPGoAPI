package jrtp

import "C"
import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"net"
	"sync"
	"time"
	"unsafe"
)

var GlobalCRtpSessionMap sync.Map

const (
	dataReceiveChanLen  = 160
	ctrlEventChanLen    = 32
	maxNumberOutStreams = 5
)

type Session struct {
	LocalIp, RemoteIp             string
	LocalPort, RemotePort         int
	localCtrlPort, RemoteCtrlPort int
	PayloadType, ClockRate, Mode  int
	fps                           int
	streamOutIndex                uint32
	profile                       string
	startFlag                     bool
	frameBufData                  []byte
	streamId                      uint32
	streamsOut                    streamOutMap
	dataReceiveChan               chan *DataPacket
	ctrlEvArr                     []*CtrlEvent
	ctrlEventChan                 CtrlEventChan
	ctx                           *CRtpSessionContext
	ini                           *CRtpSessionInitData
}

type Address struct {
	IPAddr             net.IP
	DataPort, CtrlPort int
	Zone               string
}

var logger *logrus.Entry

func init() {
	gl := logrus.New()
	// Initialize all packages logger
	InitServerLogger(gl)
}

// InitServerLogger can be called multiple times before server starts to override default logger
func InitServerLogger(gl *logrus.Logger) {
	logger = gl.WithFields(logrus.Fields{"module": "rtp"})
}

func NewSession(rp *TransportUDP, tv TransportRecv) *Session {
	initConfigOnce()

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
	dec.dataReceiveChan = make(chan *DataPacket, dataReceiveChanLen)
	dec.ctrlEventChan = make(CtrlEventChan, ctrlEventChanLen)
	dec.ctrlEvArr = make([]*CtrlEvent, 0, 10)

	if dec.ctx == nil {
		logger.Errorf("NewRtpSession error creat dec.ctx or dec.ini fail\n")
		return nil
	}

	GlobalCRtpSessionMap.Store(dec.ctx, &dec)

	logger.Infof("session localIp:%v port:%v  \n", dec.LocalIp, dec.LocalPort)

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
		pts: stamp + n.streamsOut[0].initialStamp,
	}
	return rp
}

func (n *Session) CreateDataReceiveChan() chan *DataPacket {
	return n.dataReceiveChan
}

func (n *Session) RemoveDataReceiveChan() {
	if n.dataReceiveChan != nil {
		close(n.dataReceiveChan)
	}
	n.dataReceiveChan = nil
}

func (n *Session) CreateCtrlEventChan() CtrlEventChan {
	return n.ctrlEventChan
}

// RemoveCtrlEventChan deletes the control event channel.
func (n *Session) RemoveCtrlEventChan() {
	if n.ctrlEventChan != nil {
		close(n.ctrlEventChan)
	}
	n.ctrlEventChan = nil
}

func (n *Session) StartSession() error {
	if n.ctx != nil {

		if n.initSession() != nil {
			return errors.New(fmt.Sprintf("StartSession fail,initSession error..."))
		}
		n.RegisterRtpPacketRcvCb()

		// just for test
		//n.SetRtcpDisable(true)
		n.RegisterAllTypeRtcpPacketRcvCb()
		//n.RegisterRtcpOriginPacketRcvCb()

		res := n.ctx.loopRtpSession()
		n.startFlag = true
		if res == false {
			logger.Errorf("StartSession fail,error:%v", res)
			return errors.New(fmt.Sprintf("StartSession  fail"))
		} else {
			logger.Infof("StartSession success streamId:%v\n", n.streamId)
			//go n.receivePacketLoop()
		}
	} else {
		logger.Error("StartSession fail,ctx nil\n")
		return errors.New(fmt.Sprintf("ctx nil StartSession fail"))
	}
	return nil
}

func (n *Session) CloseSession() error {
	if n.ctx != nil && n.startFlag {
		n.startFlag = false
		res := n.ctx.stopRtpSession()
		if res == false {
			logger.Error("StopSession fail,error:%v\n", res)
			return errors.New(fmt.Sprintf("StopSession fail"))
		} else {
			logger.Infof("StopSession success streamId:%v\n", n.streamId)
		}
		n.destroy()
	} else {
		return errors.New(fmt.Sprintf("ctx nil StopSession fail"))
	}
	return nil
}

func (n *Session) WriteData(rp *DataPacket) (k int, err error) {
	if n.ctx != nil && n.startFlag {
		//	logger.Infof("WriteData len:%v pt:%v pts:%v marker:%v \n", len(rp.payload), rp.payloadType, rp.pts, rp.marker)
		if rp.marker {
			//n.ctx.sendDataRtpSession(rp.payload, len(rp.payload), 1)
			n.ctx.sendDataWithTsRtpSession(rp.payload, len(rp.payload), rp.pts, 1, int(rp.payloadType))
		} else {
			//n.ctx.sendDataRtpSession(rp.payload, len(rp.payload), 0)
			n.ctx.sendDataWithTsRtpSession(rp.payload, len(rp.payload), rp.pts, 0, int(rp.payloadType))
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
			logger.Infof("InitSession id:%v ClockRate:%v  PayloadType:%v type:%v\n", n.streamsOut[0].initialStamp,
				n.streamsOut[0].profile.ClockRate, n.streamsOut[0].payloadTypeNumber, n.streamsOut[0].profile.MimeType)
		}

		if n.fps <= 0 {
			if n.streamsOut[0].profile.MediaType == Audio {
				n.fps = 40
				logger.Infof("audio seesion fps is default :%v", n.fps)
			} else if n.streamsOut[0].profile.MediaType == Video {
				n.fps = 25
				logger.Infof("video session fps is default :%v", n.fps)
			} else {
				logger.Errorf("not audio or video, error")
				n.fps = -1
			}
		}

		n.ini = creatRtpInitData(n.LocalIp, n.RemoteIp, n.LocalPort, n.RemotePort, n.PayloadType, n.ClockRate, n.fps)
		res := n.ctx.initRtpSession(n.ini)
		if res == false {
			return errors.New(fmt.Sprintf("InitSession  fail"))
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
				return
			}
			n.receiveData(buffer, 1500)
		}
	}
}

func (n *Session) HandleCallBackData(data []byte, marker bool) {
	if n.streamsOut[0].profile.MimeType == "H264" {
		n.unPackRTPToH264(data, marker)
	} else if n.streamsOut[0].profile.MimeType == "PCMA" {
		//fmt.Printf("audio pcma len:%v\n", len(data))
	}
}

func (n *Session) destroy() {
	if n.ctx != nil {
		time.Sleep(time.Second * 3)
		n.ini.destroySessionInitData()
		n.ctx.destroyRtpSession()
		GlobalCRtpSessionMap.Delete(n.ctx)
		n.ctx = nil
		logger.Infof("Session destroy %v\n", n.streamId)
	}
	n.RemoveDataReceiveChan()
	n.RemoveCtrlEventChan()
}

// maybe panic when dataReceiveChan closed
func (n *Session) receiveRtpCache(pl *DataPacket) {
	select {
	case n.dataReceiveChan <- pl:
	default:
		logger.Errorf("rtp Channel is closed or blocked,discard it")
	}
	//n.dataReceiveChan <- pl
}

func (n *Session) receiveRtcpCache(pl *CtrlEvent) {
	n.ctrlEvArr = append(n.ctrlEvArr, pl)
	select {
	case n.ctrlEventChan <- n.ctrlEvArr:
	default:
		logger.Errorf("rtcp Channel is closed or blocked,discard it")
	}

}

func (n *Session) SsrcStreamOutForIndex(streamIndex uint32) *SsrcStream {
	return n.streamsOut[streamIndex]
}

func (n *Session) NewSsrcStreamOut(own *Address, ssrc uint32, sequenceNo uint16) (index uint32, err Error) {
	str := &SsrcStream{
		sequenceNumber: sequenceNo,
		ssrc:           ssrc,
	}
	if ssrc == 0 {
		str.newSsrc()
	}

	if sequenceNo == 0 {
		str.newSequence()
	}
	str.newInitialTimestamp()
	n.streamId = str.initialStamp
	n.streamsOut[n.streamOutIndex] = str
	index = n.streamOutIndex
	n.streamOutIndex++
	return index, ""
}

func (n *Session) PacketH264ToRtpAndSend(annexbPayload []byte, pts uint32, payloadType uint8) {
	packetList := CPacketListFromH264Mode(annexbPayload, pts, payloadType, 1200, false)

	packetList.Iterate(func(p *CRtpPacketList) {
		payload, pt, pts1, mark := p.Payload, p.PayloadType, p.Pts, p.Marker
		if payload != nil {
			packet := n.NewDataPacket(pts1)
			packet.SetMarker(mark)
			packet.SetPayload(payload)
			packet.SetPayloadType(pt)

			if _, err := n.WriteData(packet); err != nil {
				logger.Errorf(" PacketH264ToRtpAndSend WriteData fail...\n")
			}

		}
	})
}

func (n *Session) unPackRTPToH264(rtpPayload []byte, marker bool) {
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
			//fmt.Printf("H264Rtp slice is first  fu-a %v\n", len(rtpPayload))
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

func (n *Session) RegisterAllTypeRtcpPacketRcvCb() {
	n.RegisterRRPacketRcvCb()
	n.RegisterSRPacketRcvCb()
	n.RegisterAppPacketRcvCb()
	n.RegisterSdesPrivateItemRcvCb()
	n.RegisterSdesItemRcvCb()
	n.RegisterByePacketRcvCb()
	n.RegisterUnKnownPacketRcvCb()
}

func (n *Session) RegisterRtpPacketRcvCb() bool {
	if n.ctx != nil {
		return n.ctx.RegisterRtpPacketRcvCb(unsafe.Pointer(n.ctx))
	} else {
		return false
	}
}

func (n *Session) RegisterRtcpOriginPacketRcvCb() bool {
	if n.ctx != nil {
		return n.ctx.RegisterOriginPacketRcvCb(unsafe.Pointer(n.ctx))
	} else {
		return false
	}
}

func (n *Session) RegisterAppPacketRcvCb() bool {
	if n.ctx != nil {
		return n.ctx.RegisterAppPacketRcvCb(unsafe.Pointer(n.ctx))
	} else {
		return false
	}
}

func (n *Session) RegisterRRPacketRcvCb() bool {
	if n.ctx != nil {
		return n.ctx.RegisterRRPacketRcvCb(unsafe.Pointer(n.ctx))
	} else {
		return false
	}
}

func (n *Session) RegisterSRPacketRcvCb() bool {
	if n.ctx != nil {
		return n.ctx.RegisterSRPacketRcvCb(unsafe.Pointer(n.ctx))
	} else {
		return false
	}
}

func (n *Session) RegisterSdesItemRcvCb() bool {
	if n.ctx != nil {
		return n.ctx.RegisterSdesItemRcvCb(unsafe.Pointer(n.ctx))
	} else {
		return false
	}
}

func (n *Session) RegisterSdesPrivateItemRcvCb() bool {
	if n.ctx != nil {
		return n.ctx.RegisterSdesItemRcvCb(unsafe.Pointer(n.ctx))
	} else {
		return false
	}
}

func (n *Session) RegisterByePacketRcvCb() bool {
	if n.ctx != nil {
		return n.ctx.RegisterByePacketRcvCb(unsafe.Pointer(n.ctx))
	} else {
		return false
	}
}

func (n *Session) RegisterUnKnownPacketRcvCb() bool {
	if n.ctx != nil {
		return n.ctx.RegisterByePacketRcvCb(unsafe.Pointer(n.ctx))
	} else {
		return false
	}
}

func (n *Session) SendRtcpAppData(subType uint8, name [4]byte, appData []byte) int {
	if n.ctx != nil {
		return n.ctx.SendRtcpAppData(subType, name, appData)
	} else {
		return -1
	}
}

func (n *Session) SendRtpOrRtcpRawData(data []byte, isRtp bool) int {
	if n.ctx != nil {
		return n.ctx.SendRtpOrRtcpRawData(data, isRtp)
	} else {
		return -1
	}
}

func (n *Session) SetRtcpDisable(disableRtcp bool) {
	if n.ctx == nil {
		return
	}
	if disableRtcp {
		n.ctx.SetRtcpDisable(1)
	} else {
		n.ctx.SetRtcpDisable(0)
	}
}

func (n *Session) SetSessionFps(fps int) {
	if fps < 0 {
		logger.Errorf("SetSessionFps fps is error ,not set")
		return
	}
	n.fps = fps
}
