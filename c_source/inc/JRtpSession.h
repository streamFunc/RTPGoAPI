//
// Created by sean on 2023/10/16.
//

#ifndef IRTP_JRTPSESSION_H
#define IRTP_JRTPSESSION_H

#include "RtpSessionMpl.h"
#include "rtpsession.h"
#include "rtpudpv4transmitter.h"
#include "rtpsessionparams.h"
#include "rtpipv4address.h"
#include "rtppacket.h"
#include "rtptimeutilities.h"

namespace iRtp{


class JRtpSession :public RtpSessionMpl{
public:
    JRtpSession();
    ~JRtpSession();

    virtual bool Init(const RtpSessionInitData* pInitData);
    virtual bool Start();
    virtual bool Stop();

    virtual int SendData(const uint8_t *buf, int len, uint16_t marker);
    virtual int SendDataWithTs(const uint8_t *buf, int len, uint32_t pts, uint16_t marker);
    virtual int RcvData(uint8_t *buf, int len,RcvCb rcvCb, void *user);
    virtual int RcvDataWithTs(uint8_t *buf, int len, uint32_t ts, RcvCb rcvCb, void *user);
    virtual int RcvPayloadData(uint8_t *buf, int len,RcvCb rcvCb, void *user);


private:
    void __updateRtpHeaderData(jrtplib::RTPPacket* p);

protected:


private:
    jrtplib::RTPSession                     m_rtpSession;
    jrtplib::RTPSessionParams               m_sessParams;
    jrtplib::RTPUDPv4TransmissionParams     m_transParams;

    int                         m_nPayloadType;
    uint32_t                    m_nCurPts;
    uint32_t                    m_nSndIncTs;



};

}//namespace iRtp
#endif //IRTP_JRTPSESSION_H
