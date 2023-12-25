//
// Created by sean on 2023/8/3.
//

#ifndef IRTP_RTPSESSIONMPL_H
#define IRTP_RTPSESSIONMPL_H

#include <atomic>
#include "ICommon.h"

#ifdef __cplusplus
extern "C" {
#endif


namespace iRtp {

    struct RtpSessionInitData {
        std::string localIp;
        std::string remoteIp;
        int localPort;
        int remotePort;
        int payloadType;
        int clockRate;  //h264=90000; audio=8000
    };

    struct RtpHeaderData{
        uint32_t ts;
        uint16_t seq;
        uint32_t ssrc;
        uint32_t csrc[16];
        uint16_t pt;
        bool     marker;
        uint8_t  version;
        bool     padding;
        bool     extension;
        uint8_t  cc;
    };


    typedef int (*RcvCb)(const uint8_t *buf, int len, int marker, void *user);


    class RtpSessionMpl {
    public:
        /*
         * finish initializing list
         */
        RtpSessionMpl() : m_bStopFlag(false) {}

        /*
         * it will do nothing. just to ensure that inherit object pointer or reference run destructor function
         * */
        virtual ~RtpSessionMpl() {}

        /*
         * initialize something such as ip,port ,payloadType and so on
         * */
        virtual bool Init(const RtpSessionInitData *pInitData) = 0;

        /*
         * it all depends on inherit object.may be not useful or just start tasks
         * */
        virtual bool Start() = 0;

        /*
         * it all depends on inherit object.may be not useful or just stop tasks
         * */
        virtual bool Stop() = 0;

        /*
         * send data
         * @param [in] buf:rtp payload data
         * @param [in] len:the len of payload data
         * @param [in] pts:present timestamp
         * @param [in] marker:a flag bit for rtp
         * @return the len of real send
         * */
        virtual int SendData(const uint8_t *buf, int len, uint16_t marker) = 0;

        /*
         * send data with ts
         * @param [in] buf:rtp payload data
         * @param [in] len:the len of payload data
         * @param [in] pts:present timestamp
         * @param [in] marker:a flag bit for rtp
         * @return the len of real send
         * */
        virtual int SendDataWithTs(const uint8_t *buf, int len, uint32_t pts, uint16_t marker) = 0;


        /*
         * receive data
         * &param [out] buf:the cache to store data.you should alloc memory by yourself before calling
         * &param [in] len:the len you expect
         * @param [in] rcvCb:user need to register callback function.
         * @param [in] user:user param
         * @return the len of real receiving one time
         */
        virtual int RcvData(uint8_t *buf, int len,RcvCb rcvCb, void *user) = 0;

        /*
         * receive data with ts
         * &param [out] buf:the cache to store data.you should alloc memory by yourself before calling
         * &param [in] len:the len you expect
         * @param [in] ts:expected timestamp
         * @param [in] rcvCb:user need to register callback function.
         * @param [in] user:user param
         * @return the len of real receiving one time
         */
        virtual int RcvDataWithTs(uint8_t *buf, int len, uint32_t ts, RcvCb rcvCb, void *user) = 0;

        /*
         * receive payload data
         * @param [out] buf:the cache to store data.you should alloc memory by yourself before calling
         * @param [in] len:the len you expect
         * @param [in] ts:expected timestamp
         * @param [in] rcvCb:user need to register callback function.
         * @param [in] user:user param
         * @return the len of real receiving one time
         */
        virtual int RcvPayloadData(uint8_t *buf, int len,RcvCb rcvCb, void *user)=0;


        /*
         * get current time rtpHeaderData
         */
        const RtpHeaderData& GetRtpHeaderData() const {return m_rtpHeaderData;}


    protected:
        std::atomic_bool m_bStopFlag;
        RtpHeaderData    m_rtpHeaderData;

    };


}//namespace iRtp

#ifdef __cplusplus
}
#endif

#endif //IRTP_RTPSESSIONMPL_H
