//
// Created by sean on 2023/10/16.
//

#ifndef IRTP_JRTPSESSION_H
#define IRTP_JRTPSESSION_H

#include "RtpSessionMpl.h"


namespace iRtp{

typedef class JRtpSessionImpl   JRtpSessionImpl;
typedef class JRTPSessionParams JRTPSessionParams;
typedef class JRTPUDPv4TransmissionParams JRTPUDPv4TransmissionParams;


class JRtpSession :public RtpSessionMpl{
public:
    JRtpSession();
    ~JRtpSession();

    virtual bool Init(const RtpSessionInitData* pInitData);

    //rtp
    virtual int SendData(const uint8_t *buf, int len, uint16_t marker,int pt);
    virtual int SendDataWithTs(const uint8_t *buf, int len, uint32_t pts, uint16_t marker,int pt);
    virtual int RcvData(uint8_t *buf, int len,RcvCb rcvCb, void *user);
    virtual int RcvDataWithTs(uint8_t *buf, int len, uint32_t ts, RcvCb rcvCb, void *user);
    virtual int RcvPayloadData(uint8_t *buf, int len,RcvCb rcvCb, void *user);

    //rtcp
    virtual int SendRtcpAppData(uint8_t subType,const uint8_t name[4],const void* appData,int appDataLen);
    virtual int SendRawData(uint8_t* data,int len,bool isRtp);

    virtual int SetSessionBandwidth(double bw);

    void TryToWakeUp(){
        if(m_pThread)tryToWakeUp();
    }

protected:
    virtual void loop();
    virtual bool stop();
    virtual void setDisableRtcp();

private:
    void __updateRtpHeaderData(void* p);


protected:


private:
    JRtpSessionImpl*                        m_pRtpSessionImpl;
    JRTPSessionParams*                      m_pSessParams;
    JRTPUDPv4TransmissionParams*            m_pTransParams;

    int                         m_nPayloadType;
    uint32_t                    m_nCurPts;
    uint32_t                    m_nSndIncTs;



};

}//namespace iRtp
#endif //IRTP_JRTPSESSION_H
