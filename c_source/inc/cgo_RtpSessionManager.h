//
// Created by sean on 2023/10/23.
//

#ifndef CGO_IRTP_RTPSESSIONMANAGER_H
#define CGO_IRTP_RTPSESSIONMANAGER_H
#include <stdio.h>
#include <stdbool.h>
#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

typedef struct CRtpSessionManager CRtpSessionManager;
typedef struct CRtpSessionInitData CRtpSessionInitData;

//rtp callback function
typedef int (*CRcvCb)(const uint8_t *buf, int len, int marker, void *user);
int RcvCb(uint8_t *buf, int len, int marker, void *user); //ensure the same function name when define with go language
typedef CRcvCb CRtpRcvCb;
int RtcpPayloadRcvCb(uint8_t *buf, int len, int marker, void *user);
int RtcpPacketRcvCb(uint8_t *buf, int len, int marker, void *user);

//rtcp callback function
typedef void (*CRtcpRcvCb)(void* rtcpPacket,void* user);
void RtcpOriginPacketRcvCb(void* rtcpPacket,void* user);
void RtcpAppPacketRcvCb(void* rtcpPacket,void* user);
void RtcpRRPacketRcvCb(void* rtcpPacket,void* user);
void RtcpSRPacketRcvCb(void* rtcpPacket,void* user);
void RtcpSdesItemRcvCb(void* rtcpPacket,void* user);
void RtcpSdesPrivateItemRcvCb(void* rtcpPacket,void* user);
void RtcpByePacketRcvCb(void* rtcpPacket,void* user);
void RtcpUnKnownPacketRcvCb(void* rtcpPacket,void* user);  //origin packet.one or more

//typedef struct CRtcpPacketType CRtcpPacketType;
typedef enum CRtpSessionType{
    CRtpSessionType_ORTP,
    CRtpSessionType_JRTP
}CRtpSessionType;

/*
 * Rtp interface.
 */
CRtpSessionManager* CreateRtpSession(CRtpSessionType t);
void DestroyRtpSession(CRtpSessionManager* p);
bool InitRtpSession(CRtpSessionManager* p,CRtpSessionInitData* pInitData);
bool StartRtpSession(CRtpSessionManager* p);
bool LoopRtpSession(CRtpSessionManager* p);
bool StopRtpSession(CRtpSessionManager* p);
int SendDataRtpSession(CRtpSessionManager* p,const uint8_t* buf,int len,uint16_t marker);
int RcvDataRtpSession(CRtpSessionManager* p,uint8_t* buf,int len,CRcvCb rcvCb,void* user);
int SendDataWithTsRtpSession(CRtpSessionManager* p,const uint8_t* buf,int len,uint32_t pts,uint16_t marker);
int RcvDataWithTsRtpSession(CRtpSessionManager* p,uint8_t* buf,int len,uint32_t ts,CRcvCb rcvCb,void* user);

/*
 * send rtcp app packet
 */
int SendRtcpAppData(CRtpSessionManager* p,uint8_t subType,const uint8_t name[4],const void* appData,int appDataLen);


/*
 * send origin rtp or rtcp packet
 */
int SendRtpOrRtcpRawData(CRtpSessionManager* p,uint8_t* data,int len,bool isRtp);

/*
 * disable rtcp.default enable
 */
void SetRtcpDisable(CRtpSessionManager* p,int disableRtcp);


/*
 * rtcp register
 * @type:
 * @cb:it should be CRtcpRcv type or occur a error
 */
bool RegisterRtpRcvCb(CRtpSessionManager* p,int type,void* cb,void* user);
/*
 * register specific rtp callback interface
 */
bool RegisterRtpOnlyPayloadRcvCb(CRtpSessionManager* p,void* cb,void* user);
bool RegisterRtpPacketRcvCb(CRtpSessionManager* p,void* cb,void* user);

/*
 * rtcp initialized interface
 * @type:
 * @cb:it should be CRtcpRcv type or occur a error
 */
bool RegisterRtcpRcvCb(CRtpSessionManager* p,int type,void* cb,void* user);
/*
 * register specific rtcp packet callback interface
 */
bool RegisterOriginPacketRcvCb(CRtpSessionManager* p,void* cb,void* user); //but we dont recommend use this interface,because there is low performance
bool RegisterAppPacketRcvCb(CRtpSessionManager* p,void* cb,void* user);
bool RegisterRRPacketRcvCb(CRtpSessionManager* p,void* cb,void* user);
bool RegisterSRPacketRcvCb(CRtpSessionManager* p,void* cb,void* user);
bool RegisterSdesItemRcvCb(CRtpSessionManager* p,void* cb,void* user);
bool RegisterSdesPrivateItemRcvCb(CRtpSessionManager* p,void* cb,void* user);
bool RegisterByePacketRcvCb(CRtpSessionManager* p,void* cb,void* user);
bool RegisterUnKnownPacketRcvCb(CRtpSessionManager* p,void* cb,void* user);




/*
 * rtp session initialized param
 */
CRtpSessionInitData*  CreateRtpSessionInitData(const char* localIp,const char* remoteIp,int localPort
                                               ,int remotePort,int payloadType,int clockRate);
void DestroyRtpSessionInitData(CRtpSessionInitData* pi);

CRtpSessionInitData* SetLocalIp(CRtpSessionInitData* p,const char* localIp);
CRtpSessionInitData* SetRemoteIp(CRtpSessionInitData* p,const char* remoteIp);
CRtpSessionInitData* SetLocalPort(CRtpSessionInitData* p,int localPort);
CRtpSessionInitData* SetRemotePort(CRtpSessionInitData* p,int remotePort);
CRtpSessionInitData* SetPayloadType(CRtpSessionInitData* p,int pt);
CRtpSessionInitData* SetClockRate(CRtpSessionInitData* p,int cr);

/*
 * extension params
 * support k-v:
 * 1.receiveBufferSize:10000 (byte)
 */
CRtpSessionInitData* AddPairsParams(CRtpSessionInitData* p,const char* key,const char* value);



//receive rtp header massage
uint32_t GetTimeStamp(void* p);
uint16_t GetSequenceNumber(void* p);
uint32_t GetSsrc(void* p);
uint32_t* GetCsrc(void* p);
uint16_t GetPayloadType(void* p);
bool     GetMarker(void* p);
uint8_t  GetVersion(void* p);
bool     GetPadding(void* p);
bool     GetExtension(void* p);
uint8_t  GetCC(void* p);


/*
 * rtcp origin packet interface
 */
uint8_t* GetRtcpPacketData(void* p,void* rtcpPacket);
int GetPacketDataLength(void* p,void* rtcpPacket);
uint32_t GetSSRC(void* p,void* rtcpPacket);

/*
 * rtcp app packet
 */
uint8_t* GetAppData(void* p,void*rtcpPacket);
int GetAppDataLength(void* p,void* rtcpPacket);
uint8_t* GetAppName(void* p,void* rtcpPacket);
uint32_t GetAppSsrc(void* p,void* rtcpPacket);
uint8_t GetAppSubType(void* p,void* rtcpPacket);


/*
 * rtcp sdes item
 */
uint8_t* GetSdesItemData(void* p,void* rtcpPacket);
int GetSdesItemDataLen(void* p,void* rtcpPacket);
int GetSdesItemType(void* p,void* rtcpPacket);

/*
 * rtcp sdes private item
 */
uint8_t* GetSdesPrivatePrefixData(void* p,void* rtcpPacket);
int GetSdesPrivatePrefixDataLen(void* p,void* rtcpPacket);
uint8_t* GetSdesPrivateValueData(void* p,void* rtcpPacket);
int GetSdesPrivateValueDataLen(void* p,void* rtcpPacket);

/*
 * unKnown packet
 */
uint8_t  GetUnknownPacketType(void* p,void* rtcpPacket);
uint8_t* GetUnKnownRtcpPacketData(void* p,void* rtcpPacket);
int GetUnKnownRtcpPacketDataLen(void* p,void* rtcpPacket);
uint32_t GetUnKnownRtcpPacketSsrc(void* p,void* rtcpPacket);


/*
 * RR or SR Packet
 */
uint8_t GetRRFractionLost(void* p,void* rtcpPacket);
uint32_t GetRRLostPacketNumber(void* p,void* rtcpPacket);
uint32_t GetRRExtendedHighestSequenceNumber(void* p,void* rtcpPacket);
uint32_t GetRRJitter(void* p,void* rtcpPacket);
uint32_t GetRRLastSR(void* p,void* rtcpPacket);
uint32_t GetRRDelaySinceLastSR(void* p,void* rtcpPacket);

/*
 * SR report packet
 */
uint32_t GetSRNtpLSWTimeStamp(void* p,void* rtcpPacket);
uint32_t GetSRNtpMSWTimeStamp(void* p,void* rtcpPacket);
uint32_t GetSRRtpTimeStamp(void* p,void* rtcpPacket);
uint32_t GetSRSenderPacketCount(void* p,void* rtcpPacket);
uint32_t GetSRSenderOctetCount(void* p,void* rtcpPacket);

/*
 * bye packet
 */
uint8_t* GetByeReasonData(void* p,void* rtcpPacket);
int GetByeReasonDataLen(void* p,void* rtcpPacket);

#ifdef __cplusplus
}
#endif

#endif //IRTP_RTPSESSIONMANAGER_H
