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
	//fmt.Println("Receive payload len=", len, "seq=", C.GetSequenceNumber(user), " from ssrc=", C.GetSsrc(user), " marker=", marker, " user=", user, " pt=", C.GetPayloadType(user))

	handle := (*CRtpSessionContext)(user)
	length := int(dataLen)
	data := (*[1 << 30]byte)(unsafe.Pointer(buf))[:length:length]

	slice := make([]byte, length)
	copy(slice, data)

	GlobalCRtpSessionMapMutex.Lock()
	nUser := GlobalCRtpSessionMap[handle]
	GlobalCRtpSessionMapMutex.Unlock()

	// Parse RTP payload
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

	//nUser.HandleCallBackData(payload, flag)
	nUser.receiveRtpCache(rp)

	return dataLen

}

type (
	CRtpSessionContext  = C.struct_CRtpSessionManager
	CRtpSessionInitData = C.struct_CRtpSessionInitData
)

func creatRtpInitData(localIp, remoteIp string, localPort, remotePort, payloadType, clockRate int) *CRtpSessionInitData {
	l := C.CString(localIp)
	defer C.free(unsafe.Pointer(l))

	r := C.CString(remoteIp)
	defer C.free(unsafe.Pointer(r))

	return (*CRtpSessionInitData)(C.CreateRtpSessionInitData(l, r, (C.int)(localPort), (C.int)(remotePort), (C.int)(payloadType), (C.int)(clockRate)))
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

func (rtp *CRtpSessionContext) stopRtpSession() bool {
	if rtp == nil {
		return false
	}
	return bool(C.StopRtpSession(rtp))
}

func (rtp *CRtpSessionContext) sendDataRtpSession(payload []byte, len, marker int) int {
	if rtp == nil {
		return -1
	}
	res := int(C.SendDataRtpSession(rtp, (*C.uint8_t)(unsafe.Pointer(&payload[0])), (C.int)(len), (C.uint16_t)(marker)))
	return res
}

func (rtp *CRtpSessionContext) sendDataWithTsRtpSession(payload []byte, len int, pts uint32, marker int) int {
	if rtp == nil {
		return -1
	}
	res := int(C.SendDataWithTsRtpSession(rtp, (*C.uint8_t)(unsafe.Pointer(&payload[0])), (C.int)(len), (C.uint32_t)(pts), (C.uint16_t)(marker)))
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
