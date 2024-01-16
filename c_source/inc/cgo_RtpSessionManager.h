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

typedef int (*CRcvCb)(const uint8_t *buf, int len, int marker, void *user);
int RcvCb(uint8_t *buf, int len, int marker, void *user); //ensure the same function name when define with go language

typedef enum CRtpSessionType{
    CRtpSessionType_ORTP,
    CRtpSessionType_JRTP
}CRtpSessionType;

CRtpSessionManager* CreateRtpSession(CRtpSessionType t);
void DestroyRtpSession(CRtpSessionManager* p);
bool InitRtpSession(CRtpSessionManager* p,CRtpSessionInitData* pInitData);
bool StartRtpSession(CRtpSessionManager* p);
bool StopRtpSession(CRtpSessionManager* p);
int SendDataRtpSession(CRtpSessionManager* p,const uint8_t* buf,int len,uint16_t marker);
int RcvDataRtpSession(CRtpSessionManager* p,uint8_t* buf,int len,CRcvCb rcvCb,void* user);
int SendDataWithTsRtpSession(CRtpSessionManager* p,const uint8_t* buf,int len,uint32_t pts,uint16_t marker);
int RcvDataWithTsRtpSession(CRtpSessionManager* p,uint8_t* buf,int len,uint32_t ts,CRcvCb rcvCb,void* user);

CRtpSessionInitData*  CreateRtpSessionInitData(const char* localIp,const char* remoteIp,int localPort
                                               ,int remotePort,int payloadType,int clockRate);
void DestroyRtpSessionInitData(CRtpSessionInitData* pi);

//receive rtp header
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


//rtcp
//inline uint8_t* GetPacketData(void* rtcpPacket);
//inline int GetPacketDataLength(void* rtcpPacket);
//inline uint8_t* GetAppData(void* p);
//inline int GetAppDataLength(void* rtcpPacket);
//inline uint8_t* GetAppName(void* rtcpPacket);
//inline uint32_t GetAppSsrc(void* rtcpPacket);
//inline uint8_t GetAppSubType(void* rtcpPacket);


#ifdef __cplusplus
}
#endif

#endif //IRTP_RTPSESSIONMANAGER_H
