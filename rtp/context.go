package rtp

import "C"

//#cgo CXXFLAGS: -std=c++11
//#cgo  CFLAGS:-I../c_source/inc
//#cgo LDFLAGS: -lm -L../c_source/lib -lIRtp-static  -lstdc++ -lm
//#include <stdio.h>
//#include <stdlib.h>
//#include <inttypes.h>
//#include <stdint.h>
//#include <string.h>
//#include "cgo_RtpSessionManager.h"
import "C"
import (
	"unsafe"
)

//export RcvCb
func RcvCb(buf *C.uint8_t, dataLen C.int, marker C.int, user unsafe.Pointer) C.int {
	if user == nil && marker == 1 || buf == nil {
		return -1
	}
	//fmt.Println( "Receive payload len=", dataLen, "seq=", C.GetSequenceNumber(user), " from ssrc=", C.GetSsrc(user),
	//" marker=", marker, " user=", user, " pt=", C.GetPayloadType(user))

	handle := (*CRtpSessionContext)(user)
	length := int(dataLen)
	data := (*[1 << 30]byte)(unsafe.Pointer(buf))[:length:length]

	slice := make([]byte, length)
	copy(slice, data)

	payload := parseRtpPayload(slice)

	var flag bool
	if marker == 0 {
		flag = false
	} else {
		flag = true
	}

	rp := &DataPacket{
		payload:     payload,
		rawBuffer:   slice,
		pts:         handle.getTimeStamp(),
		marker:      flag,
		payloadType: uint8(handle.getPayloadType()),
		ssrc:        handle.getSsrc(),
		csrc:        handle.getCSrc(),
		seq:         handle.getSequenceNumber(),
	}

	if val, ok := GlobalCRtpSessionMap.Load(handle); ok {
		if session, ok := val.(*Session); ok {
			session.receiveRtpCache(rp)
			//session.HandleCallBackData(payload, flag)
		} else {
			logger.Errorf("not find user,had destory\n")
		}
	} else {
		logger.Errorf("not find handle,return\n")
	}

	return dataLen

}

//export RtcpOriginPacketRcvCb
func RtcpOriginPacketRcvCb(rtcpPacket unsafe.Pointer, user unsafe.Pointer) {
	if user == nil {
		return
	}
	handle := (*CRtpSessionContext)(user)

	logger.Info("Receive rtcp origin dataLen=", handle.GetRtcpOriginDataLen(rtcpPacket), "ssrc=", handle.GetRtcpOriginSsrc(rtcpPacket))
}

//export RtcpAppPacketRcvCb
func RtcpAppPacketRcvCb(rtcpPacket unsafe.Pointer, user unsafe.Pointer) {
	if user == nil {
		return
	}
	handle := (*CRtpSessionContext)(user)

	//fmt.Println("Receive rtcp AppData name=", handle.GetAppName(rtcpPacket), "subType=", C.GetAppSubType(user, rtcpPacket))

	rp := &CtrlEvent{
		EventType: RtcpApp,
	}

	if val, ok := GlobalCRtpSessionMap.Load(handle); ok {
		if session, ok := val.(*Session); ok {
			session.receiveRtcpCache(rp)
		} else {
			logger.Errorf("RtcpAppPacketRcvCb cb not found user,had destory\n")
		}
	}

}

//export RtcpRRPacketRcvCb
func RtcpRRPacketRcvCb(rtcpPacket unsafe.Pointer, user unsafe.Pointer) {
	if user == nil {
		return
	}
	handle := (*CRtpSessionContext)(user)
	//fmt.Println("Receive rtcp RR lost packet=", handle.GetRRLostPacketNumber(rtcpPacket), "jitter=", handle.GetRRJitter(rtcpPacket))

	/*packetLen := 88
	// 将 rtcpPacket 转换为 []byte 切片
	rtcpData := make([]byte, packetLen)
	rtcpDataPtr := uintptr(rtcpPacket)
	for i := 0; i < packetLen; i++ {
		rtcpData[i] = *(*byte)(unsafe.Pointer(rtcpDataPtr + uintptr(i)))
	}*/

	rp := &CtrlEvent{
		EventType: RtcpRR,
		//buffer:    rtcpData,
	}

	if val, ok := GlobalCRtpSessionMap.Load(handle); ok {
		if session, ok := val.(*Session); ok {
			session.receiveRtcpCache(rp)
		} else {
			logger.Errorf("RtcpRRPacketRcvCb cb not found user,had destory\n")
		}
	}
}

//export RtcpSRPacketRcvCb
func RtcpSRPacketRcvCb(rtcpPacket unsafe.Pointer, user unsafe.Pointer) {
	if user == nil {
		return
	}
	handle := (*CRtpSessionContext)(user)
	//fmt.Println("Receive rtcp SR sender packet count=", handle.GetSRSenderPacketCount(rtcpPacket))

	rp := &CtrlEvent{
		EventType: RtcpSR,
	}

	if val, ok := GlobalCRtpSessionMap.Load(handle); ok {
		if session, ok := val.(*Session); ok {
			session.receiveRtcpCache(rp)
		} else {
			logger.Errorf("RtcpSRPacketRcvCb cb not found user,had destory\n")
		}
	}
}

//export RtcpSdesItemRcvCb
func RtcpSdesItemRcvCb(rtcpPacket unsafe.Pointer, user unsafe.Pointer) {
	if user == nil {
		return
	}
	handle := (*CRtpSessionContext)(user)
	//fmt.Println("Receive rtcp sdes item packet len=", handle.GetSdesItemDataLen(rtcpPacket), "type=", handle.GetSdesItemType(rtcpPacket))

	rp := &CtrlEvent{
		EventType: RtcpSdes,
	}

	if val, ok := GlobalCRtpSessionMap.Load(handle); ok {
		if session, ok := val.(*Session); ok {
			session.receiveRtcpCache(rp)
		} else {
			logger.Errorf("RtcpSdesItemRcvCb cb not found user,had destory\n")
		}
	}

}

//export RtcpSdesPrivateItemRcvCb
func RtcpSdesPrivateItemRcvCb(rtcpPacket unsafe.Pointer, user unsafe.Pointer) {
	if user == nil {
		return
	}
	handle := (*CRtpSessionContext)(user)
	//fmt.Println("Receive rtcp SdesPrivateItem packet len=", handle.GetSdesPrivateValueDataLen(rtcpPacket))

	rp := &CtrlEvent{
		EventType: RtcpSdes,
	}

	if val, ok := GlobalCRtpSessionMap.Load(handle); ok {
		if session, ok := val.(*Session); ok {
			session.receiveRtcpCache(rp)
		} else {
			logger.Errorf("RtcpSdesPrivateItemRcvCb cb not found user,had destory\n")
		}
	}
}

//export RtcpByePacketRcvCb
func RtcpByePacketRcvCb(rtcpPacket unsafe.Pointer, user unsafe.Pointer) {
	if user == nil {
		return
	}
	handle := (*CRtpSessionContext)(user)
	//fmt.Println("Receive rtcp bye reason len=", handle.GetByeReasonDataLen(rtcpPacket))

	rp := &CtrlEvent{
		EventType: RtcpBye,
	}

	if val, ok := GlobalCRtpSessionMap.Load(handle); ok {
		if session, ok := val.(*Session); ok {
			session.receiveRtcpCache(rp)
		} else {
			logger.Errorf("RtcpByePacketRcvCb cb not found user,had destory\n")
		}
	}
}

//export RtcpUnKnownPacketRcvCb
func RtcpUnKnownPacketRcvCb(rtcpPacket unsafe.Pointer, user unsafe.Pointer) {
	if user == nil {
		return
	}
	handle := (*CRtpSessionContext)(user)
	//fmt.Println("Receive rtcp unKnow  len=", handle.GetUnKnownRtcpPacketDataLen(rtcpPacket))

	rp := &CtrlEvent{
		EventType: unKnown,
	}

	if val, ok := GlobalCRtpSessionMap.Load(handle); ok {
		if session, ok := val.(*Session); ok {
			session.receiveRtcpCache(rp)
		} else {
			logger.Errorf("RtcpUnKnownPacketRcvCb cb not found user,had destory\n")
		}
	}
}

type (
	CRtpSessionContext  = C.struct_CRtpSessionManager
	CRtpSessionInitData = C.struct_CRtpSessionInitData
)

func creatRtpInitData(localIp, remoteIp string, localPort, remotePort, payloadType, clockRate, fps int) *CRtpSessionInitData {
	l := C.CString(localIp)
	defer C.free(unsafe.Pointer(l))

	r := C.CString(remoteIp)
	defer C.free(unsafe.Pointer(r))

	return (*CRtpSessionInitData)(C.CreateRtpSessionInitData(l, r, (C.int)(localPort), (C.int)(remotePort), (C.int)(payloadType), (C.int)(clockRate), (C.int)(fps)))
}

func parseRtpPayload(buf []byte) []byte {
	// RTP header is usually 12 bytes
	headerSize := 12
	payload := buf[headerSize:]
	return payload
}

func (init *CRtpSessionInitData) destroySessionInitData() {
	if init == nil {
		return
	}
	C.DestroyRtpSessionInitData(init)
}

func newSessionContext(mode int) *CRtpSessionContext {
	var t C.CRtpSessionType
	if mode == 0 {
		t = C.CRtpSessionType_ORTP
	} else {
		t = C.CRtpSessionType_JRTP
	}
	return (*CRtpSessionContext)(C.CreateRtpSession(t))
}

func (rtp *CRtpSessionContext) destroyRtpSession() {
	if rtp == nil {
		return
	}
	C.DestroyRtpSession(rtp)
}

func (rtp *CRtpSessionContext) initRtpSession(v *CRtpSessionInitData) bool {
	if rtp == nil {
		return false
	}
	return bool(C.InitRtpSession(rtp, v))
}

func (rtp *CRtpSessionContext) startRtpSession() bool {
	if rtp == nil {
		return false
	}
	return bool(C.StartRtpSession(rtp))
}

func (rtp *CRtpSessionContext) loopRtpSession() bool {
	if rtp == nil {
		return false
	}
	return bool(C.LoopRtpSession(rtp))
}

func (rtp *CRtpSessionContext) stopRtpSession() bool {
	if rtp == nil {
		return false
	}
	return bool(C.StopRtpSession(rtp))
}

func (rtp *CRtpSessionContext) sendDataRtpSession(payload []byte, len, marker, payloadType int) int {
	if rtp == nil {
		return -1
	}
	res := int(C.SendDataRtpSession(rtp, (*C.uint8_t)(unsafe.Pointer(&payload[0])), (C.int)(len), (C.uint16_t)(marker), (C.int)(payloadType)))
	return res
}

func (rtp *CRtpSessionContext) sendDataWithTsRtpSession(payload []byte, len int, pts uint32, marker, payloadType int) int {
	if rtp == nil {
		return -1
	}
	res := int(C.SendDataWithTsRtpSession(rtp, (*C.uint8_t)(unsafe.Pointer(&payload[0])), (C.int)(len), (C.uint32_t)(pts), (C.uint16_t)(marker), (C.int)(payloadType)))
	return res
}

func (rtp *CRtpSessionContext) rcvDataRtpSession(buffer []byte, len int, user unsafe.Pointer) int {
	if rtp == nil {
		return -1
	}
	return int(C.RcvDataRtpSession(rtp, (*C.uint8_t)(unsafe.Pointer(&buffer[0])), (C.int)(len), C.CRcvCb(C.RcvCb), user))
}

func (rtp *CRtpSessionContext) rcvDataWithTsRtpSession(buffer []byte, len int, pts uint32, rcvCb C.CRcvCb, user unsafe.Pointer) int {
	if rtp == nil {
		return -1
	}
	return int(C.RcvDataWithTsRtpSession(rtp, (*C.uint8_t)(unsafe.Pointer(&buffer[0])), (C.int)(len), (C.uint32_t)(pts), rcvCb, user))
}

func (rtp *CRtpSessionContext) getTimeStamp() uint32 {
	if rtp == nil {
		return 0
	}
	t := C.GetTimeStamp(unsafe.Pointer(rtp))
	return uint32(t)
}

func (rtp *CRtpSessionContext) getSequenceNumber() uint16 {
	if rtp == nil {
		return 0
	}
	t := C.GetSequenceNumber(unsafe.Pointer(rtp))
	return uint16(t)
}

func (rtp *CRtpSessionContext) getSsrc() uint32 {
	if rtp == nil {
		return 0
	}
	t := C.GetSsrc(unsafe.Pointer(rtp))
	return uint32(t)
}

func (rtp *CRtpSessionContext) getCSrc() []uint32 {
	if rtp == nil {
		return nil
	}
	t := C.GetCsrc(unsafe.Pointer(rtp))
	dataLen := 16
	csSrc := (*[1 << 30]C.uint32_t)(unsafe.Pointer(t))[:dataLen:dataLen] // 使用切片将指针转换为Go的uint32切片
	return *(*[]uint32)(unsafe.Pointer(&csSrc))
}

func (rtp *CRtpSessionContext) getPayloadType() uint16 {
	if rtp == nil {
		return 0
	}
	t := C.GetPayloadType(unsafe.Pointer(rtp))
	return uint16(t)
}

func (rtp *CRtpSessionContext) getMarker() bool {
	if rtp == nil {
		return false
	}
	t := C.GetMarker(unsafe.Pointer(rtp))
	return bool(t)
}

func (rtp *CRtpSessionContext) getVersion() uint8 {
	if rtp == nil {
		return 0
	}
	t := C.GetVersion(unsafe.Pointer(rtp))
	return uint8(t)
}

func (rtp *CRtpSessionContext) getPadding() bool {
	if rtp == nil {
		return false
	}
	t := C.GetPadding(unsafe.Pointer(rtp))
	return bool(t)
}

func (rtp *CRtpSessionContext) getExtension() bool {
	if rtp == nil {
		return false
	}
	t := C.GetExtension(unsafe.Pointer(rtp))
	return bool(t)
}

func (rtp *CRtpSessionContext) getCC() uint8 {
	if rtp == nil {
		return 0
	}
	t := C.GetCC(unsafe.Pointer(rtp))
	return uint8(t)
}

// rtcp register

func (rtp *CRtpSessionContext) RegisterOriginPacketRcvCb(user unsafe.Pointer) bool {
	return bool(C.RegisterOriginPacketRcvCb(rtp, unsafe.Pointer(C.CRtcpRcvCb(C.RtcpOriginPacketRcvCb)), user))
}

func (rtp *CRtpSessionContext) RegisterAppPacketRcvCb(user unsafe.Pointer) bool {
	return bool(C.RegisterAppPacketRcvCb(rtp, unsafe.Pointer(C.CRtcpRcvCb(C.RtcpAppPacketRcvCb)), user))
}

func (rtp *CRtpSessionContext) RegisterRRPacketRcvCb(user unsafe.Pointer) bool {
	return bool(C.RegisterRRPacketRcvCb(rtp, unsafe.Pointer(C.CRtcpRcvCb(C.RtcpRRPacketRcvCb)), user))
}

func (rtp *CRtpSessionContext) RegisterSRPacketRcvCb(user unsafe.Pointer) bool {
	return bool(C.RegisterSRPacketRcvCb(rtp, unsafe.Pointer(C.CRtcpRcvCb(C.RtcpSRPacketRcvCb)), user))
}

func (rtp *CRtpSessionContext) RegisterSdesItemRcvCb(user unsafe.Pointer) bool {
	return bool(C.RegisterSdesItemRcvCb(rtp, unsafe.Pointer(C.CRtcpRcvCb(C.RtcpSdesItemRcvCb)), user))
}

func (rtp *CRtpSessionContext) RegisterSdesPrivateItemRcvCb(user unsafe.Pointer) bool {
	return bool(C.RegisterSdesPrivateItemRcvCb(rtp, unsafe.Pointer(C.CRtcpRcvCb(C.RtcpSdesPrivateItemRcvCb)), user))
}

func (rtp *CRtpSessionContext) RegisterByePacketRcvCb(user unsafe.Pointer) bool {
	return bool(C.RegisterByePacketRcvCb(rtp, unsafe.Pointer(C.CRtcpRcvCb(C.RtcpByePacketRcvCb)), user))
}

func (rtp *CRtpSessionContext) RegisterUnKnownPacketRcvCb(user unsafe.Pointer) bool {
	return bool(C.RegisterUnKnownPacketRcvCb(rtp, unsafe.Pointer(C.CRtcpRcvCb(C.RtcpUnKnownPacketRcvCb)), user))
}

//rtcp origin data

func (rtp *CRtpSessionContext) GetRtcpOriginData(rtcpPacket unsafe.Pointer) []uint8 {
	if rtp == nil {
		return nil
	}
	cData := C.GetRtcpPacketData(unsafe.Pointer(rtp), rtcpPacket)
	defer C.free(unsafe.Pointer(cData))

	dataLen := rtp.GetRtcpOriginDataLen(rtcpPacket)
	return C.GoBytes(unsafe.Pointer(cData), C.int(dataLen))

}

func (rtp *CRtpSessionContext) GetRtcpOriginDataLen(rtcpPacket unsafe.Pointer) int {
	if rtp == nil {
		return 0
	}
	return int(C.GetPacketDataLength(unsafe.Pointer(rtp), rtcpPacket))
}

func (rtp *CRtpSessionContext) GetRtcpOriginSsrc(rtcpPacket unsafe.Pointer) uint32 {
	if rtp == nil {
		return 0
	}
	return uint32(C.GetSSRC(unsafe.Pointer(rtp), rtcpPacket))
}

//  rtcp app packet

func (rtp *CRtpSessionContext) GetAppData(rtcpPacket unsafe.Pointer) []uint8 {
	if rtp == nil {
		return nil
	}
	cData := C.GetAppData(unsafe.Pointer(rtp), rtcpPacket)
	defer C.free(unsafe.Pointer(cData))

	dataLen := rtp.GetAppDataLength(rtcpPacket)
	return C.GoBytes(unsafe.Pointer(cData), C.int(dataLen))

}

func (rtp *CRtpSessionContext) GetAppDataLength(rtcpPacket unsafe.Pointer) int {
	if rtp == nil {
		return -1
	}
	return int(C.GetAppDataLength(unsafe.Pointer(rtp), rtcpPacket))
}

func (rtp *CRtpSessionContext) GetAppName(rtcpPacket unsafe.Pointer) []uint8 {
	if rtp == nil {
		return nil
	}
	cData := C.GetAppName(unsafe.Pointer(rtp), rtcpPacket)
	defer C.free(unsafe.Pointer(cData))

	dataLen := rtp.GetAppDataLength(rtcpPacket)
	return C.GoBytes(unsafe.Pointer(cData), C.int(dataLen))

}

func (rtp *CRtpSessionContext) GetAppSsrc(rtcpPacket unsafe.Pointer) uint32 {
	if rtp == nil {
		return 0
	}
	return uint32(C.GetAppSsrc(unsafe.Pointer(rtp), rtcpPacket))
}

func (rtp *CRtpSessionContext) GetAppSubType(rtcpPacket unsafe.Pointer) uint8 {
	if rtp == nil {
		return 0
	}
	return uint8(C.GetAppSubType(unsafe.Pointer(rtp), rtcpPacket))
}

//rtcp sdes item

func (rtp *CRtpSessionContext) GetSdesItemData(rtcpPacket unsafe.Pointer) []uint8 {
	if rtp == nil {
		return nil
	}
	cData := C.GetSdesItemData(unsafe.Pointer(rtp), rtcpPacket)
	defer C.free(unsafe.Pointer(cData))

	dataLen := rtp.GetSdesItemDataLen(rtcpPacket)
	return C.GoBytes(unsafe.Pointer(cData), C.int(dataLen))
}

func (rtp *CRtpSessionContext) GetSdesItemDataLen(rtcpPacket unsafe.Pointer) int {
	if rtp == nil {
		return -1
	}
	return int(C.GetSdesItemDataLen(unsafe.Pointer(rtp), rtcpPacket))
}

func (rtp *CRtpSessionContext) GetSdesItemType(rtcpPacket unsafe.Pointer) int {
	if rtp == nil {
		return -1
	}
	return int(C.GetSdesItemType(unsafe.Pointer(rtp), rtcpPacket))
}

// rtcp sdes private item

func (rtp *CRtpSessionContext) GetSdesPrivatePrefixData(rtcpPacket unsafe.Pointer) []uint8 {
	if rtp == nil {
		return nil
	}
	cData := C.GetSdesPrivatePrefixData(unsafe.Pointer(rtp), rtcpPacket)
	defer C.free(unsafe.Pointer(cData))

	dataLen := rtp.GetSdesPrivatePrefixDataLen(rtcpPacket)
	return C.GoBytes(unsafe.Pointer(cData), C.int(dataLen))
}

func (rtp *CRtpSessionContext) GetSdesPrivatePrefixDataLen(rtcpPacket unsafe.Pointer) int {
	if rtp == nil {
		return -1
	}
	return int(C.GetSdesPrivatePrefixDataLen(unsafe.Pointer(rtp), rtcpPacket))
}

func (rtp *CRtpSessionContext) GetSdesPrivateValueData(rtcpPacket unsafe.Pointer) []uint8 {
	if rtp == nil {
		return nil
	}
	cData := C.GetSdesPrivateValueData(unsafe.Pointer(rtp), rtcpPacket)
	defer C.free(unsafe.Pointer(cData))

	dataLen := rtp.GetSdesPrivateValueDataLen(rtcpPacket)
	return C.GoBytes(unsafe.Pointer(cData), C.int(dataLen))

}

func (rtp *CRtpSessionContext) GetSdesPrivateValueDataLen(rtcpPacket unsafe.Pointer) int {
	if rtp == nil {
		return -1
	}
	return int(C.GetSdesPrivateValueDataLen(unsafe.Pointer(rtp), rtcpPacket))
}

// rtcp unKnown packet

func (rtp *CRtpSessionContext) GetUnknownPacketType(rtcpPacket unsafe.Pointer) uint8 {
	if rtp == nil {
		return 0
	}
	return uint8(C.GetUnknownPacketType(unsafe.Pointer(rtp), rtcpPacket))
}

func (rtp *CRtpSessionContext) GetUnKnownRtcpPacketData(rtcpPacket unsafe.Pointer) []uint8 {
	if rtp == nil {
		return nil
	}
	cData := C.GetUnKnownRtcpPacketData(unsafe.Pointer(rtp), rtcpPacket)
	defer C.free(unsafe.Pointer(cData))

	dataLen := rtp.GetUnKnownRtcpPacketDataLen(rtcpPacket)
	return C.GoBytes(unsafe.Pointer(cData), C.int(dataLen))

}

func (rtp *CRtpSessionContext) GetUnKnownRtcpPacketDataLen(rtcpPacket unsafe.Pointer) int {
	if rtp == nil {
		return -1
	}
	return int(C.GetUnKnownRtcpPacketDataLen(unsafe.Pointer(rtp), rtcpPacket))
}

func (rtp *CRtpSessionContext) GetUnKnownRtcpPacketSsrc(rtcpPacket unsafe.Pointer) uint32 {
	if rtp == nil {
		return 0
	}
	return uint32(C.GetUnKnownRtcpPacketSsrc(unsafe.Pointer(rtp), rtcpPacket))
}

//rtcp RR or SR Packet

func (rtp *CRtpSessionContext) GetRRFractionLost(rtcpPacket unsafe.Pointer) uint8 {
	if rtp == nil {
		return 0
	}
	return uint8(C.GetRRFractionLost(unsafe.Pointer(rtp), rtcpPacket))
}

func (rtp *CRtpSessionContext) GetRRLostPacketNumber(rtcpPacket unsafe.Pointer) uint32 {
	if rtp == nil {
		return 0
	}
	return uint32(C.GetRRLostPacketNumber(unsafe.Pointer(rtp), rtcpPacket))
}

func (rtp *CRtpSessionContext) GetRRExtendedHighestSequenceNumber(rtcpPacket unsafe.Pointer) uint32 {
	if rtp == nil {
		return 0
	}
	return uint32(C.GetRRExtendedHighestSequenceNumber(unsafe.Pointer(rtp), rtcpPacket))
}

func (rtp *CRtpSessionContext) GetRRJitter(rtcpPacket unsafe.Pointer) uint32 {
	if rtp == nil {
		return 0
	}
	return uint32(C.GetRRJitter(unsafe.Pointer(rtp), rtcpPacket))
}

func (rtp *CRtpSessionContext) GetRRLastSR(rtcpPacket unsafe.Pointer) uint32 {
	if rtp == nil {
		return 0
	}
	return uint32(C.GetRRLastSR(unsafe.Pointer(rtp), rtcpPacket))
}

func (rtp *CRtpSessionContext) GetRRDelaySinceLastSR(rtcpPacket unsafe.Pointer) uint32 {
	if rtp == nil {
		return 0
	}
	return uint32(C.GetRRDelaySinceLastSR(unsafe.Pointer(rtp), rtcpPacket))
}

// rtcp SR report packet

func (rtp *CRtpSessionContext) GetSRNtpLSWTimeStamp(rtcpPacket unsafe.Pointer) uint32 {
	if rtp == nil {
		return 0
	}
	return uint32(C.GetSRNtpLSWTimeStamp(unsafe.Pointer(rtp), rtcpPacket))
}

func (rtp *CRtpSessionContext) GetSRNtpMSWTimeStamp(rtcpPacket unsafe.Pointer) uint32 {
	if rtp == nil {
		return 0
	}
	return uint32(C.GetSRNtpMSWTimeStamp(unsafe.Pointer(rtp), rtcpPacket))
}

func (rtp *CRtpSessionContext) GetSRRtpTimeStamp(rtcpPacket unsafe.Pointer) uint32 {
	if rtp == nil {
		return 0
	}
	return uint32(C.GetSRRtpTimeStamp(unsafe.Pointer(rtp), rtcpPacket))
}

func (rtp *CRtpSessionContext) GetSRSenderPacketCount(rtcpPacket unsafe.Pointer) uint32 {
	if rtp == nil {
		return 0
	}
	return uint32(C.GetSRSenderPacketCount(unsafe.Pointer(rtp), rtcpPacket))
}

func (rtp *CRtpSessionContext) GetSRSenderOctetCount(rtcpPacket unsafe.Pointer) uint32 {
	if rtp == nil {
		return 0
	}
	return uint32(C.GetSRSenderOctetCount(unsafe.Pointer(rtp), rtcpPacket))
}

//rtcp bye packet

func (rtp *CRtpSessionContext) GetByeReasonData(rtcpPacket unsafe.Pointer) []uint8 {
	if rtp == nil {
		return nil
	}

	cData := C.GetByeReasonData(unsafe.Pointer(rtp), rtcpPacket)
	defer C.free(unsafe.Pointer(cData))

	// 转换 C 字节切片为 Go 的字节切片
	dataLen := rtp.GetByeReasonDataLen(rtcpPacket)
	return C.GoBytes(unsafe.Pointer(cData), C.int(dataLen))

}

func (rtp *CRtpSessionContext) GetByeReasonDataLen(rtcpPacket unsafe.Pointer) int {
	if rtp == nil {
		return -1
	}
	return int(C.GetByeReasonDataLen(unsafe.Pointer(rtp), rtcpPacket))
}

func (rtp *CRtpSessionContext) SendRtcpAppData(subType uint8, name [4]byte, appData []byte) int {
	if appData == nil {
		return 0
	}
	cAppData := unsafe.Pointer(nil)
	if len(appData) > 0 {
		cAppData = unsafe.Pointer(&appData[0])
	}

	return int(C.SendRtcpAppData(rtp, C.uint8_t(subType), (*C.uint8_t)(&name[0]), cAppData, C.int(len(appData))))
}

func (rtp *CRtpSessionContext) SendRtpOrRtcpRawData(data []byte, isRtp bool) int {
	if data == nil {
		return 0
	}
	cData := unsafe.Pointer(nil)
	if len(data) > 0 {
		cData = unsafe.Pointer(&data[0])
	}
	return int(C.SendRtpOrRtcpRawData(rtp, (*C.uint8_t)(cData), C.int(len(data)), C.bool(isRtp)))
}

func (rtp *CRtpSessionContext) SetRtcpDisable(disableRtcp int) {
	C.SetRtcpDisable(rtp, C.int(disableRtcp))
}

func (rtp *CRtpSessionContext) RegisterRtpPacketRcvCb(user unsafe.Pointer) bool {
	return bool(C.RegisterRtpPacketRcvCb(rtp, unsafe.Pointer(C.CRcvCb(C.RcvCb)), user))
}

// RegisterRtpOnlyPayloadRcvCb not used
func (rtp *CRtpSessionContext) RegisterRtpOnlyPayloadRcvCb(user unsafe.Pointer) bool {
	return bool(C.RegisterRtpOnlyPayloadRcvCb(rtp, unsafe.Pointer(C.CRcvCb(C.RcvCb)), user))
}
