//
// Created by sean on 2023/8/3.
//

#ifndef IRTP_ORTPSESSION_H
#define IRTP_ORTPSESSION_H
#include <ortp/ortp.h>
#include "RtpSessionMpl.h"


namespace iRtp{

class ORtpSession :public RtpSessionMpl{
public:
    ORtpSession();
    virtual ~ORtpSession();
    virtual bool Init(const RtpSessionInitData* pInitData);

    virtual int SendData(const uint8_t *buf, int len, uint16_t marker,int pt);
    virtual int SendDataWithTs(const uint8_t *buf, int len, uint32_t pts, uint16_t marker,int pt);
    virtual int RcvData(uint8_t *buf, int len,RcvCb rcvCb, void *user);
    virtual int RcvDataWithTs(uint8_t *buf, int len, uint32_t ts, RcvCb rcvCb, void *user);
    virtual int RcvPayloadData(uint8_t *buf, int len,RcvCb rcvCb, void *user);

    static void StaticInit();
    static void StaticUnInit();

protected:
    virtual void loop();
    virtual bool stop();

private:
    void __updateRtpHeaderData(mblk_t* mp);
//    int __rcvDataWithoutJitter(uint8_t *buf, int len,RcvCb rcvCb, void *user);
//    int __rcvDataWithJitter(uint8_t *buf, int len,RcvCb rcvCb, void *user);

private:
    RtpSession*                 m_pRtpSession;
    std::string                 m_strRemoteIp;
    int                         m_nRemotePort;
    bool                        m_bIsFirst;

    uint32_t                    m_nSndPreviousTs;
    uint32_t                    m_nSndIncTs;
    uint32_t                    m_nRcvPreviousTs;
    uint32_t                    m_nRcvIncTs;
    uint16_t                    m_nRcvSeq;


};

}//namespace namespace
#endif //IRTP_ORTPSESSION_H
