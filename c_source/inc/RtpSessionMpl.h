//
// Created by sean on 2023/8/3.
//

#ifndef IRTP_RTPSESSIONMPL_H
#define IRTP_RTPSESSIONMPL_H

#include <atomic>
#include "ICommon.h"
#include <map>
#include "RtcpPacket.h"

//#ifdef __cplusplus
//extern "C" {
//#endif


namespace iRtp {

    struct RtpSessionInitData {
        RtpSessionInitData(){}
        RtpSessionInitData(const std::string& lip,const std::string& rip,int lport,int rport,int pt,int cr)
        :localIp(lip),remoteIp(rip),localPort(lport),remotePort(rport),payloadType(pt),clockRate(cr){}
        ~RtpSessionInitData(){
            if(!extraParams.empty())extraParams.clear();
        }
        void AddPairsParam(std::string k,std::string v){
            extraParams[k]=v;
        }

        const std::map<std::string,std::string>& GetExtraParamsMap()const {return extraParams;}

        std::string localIp;
        std::string remoteIp;
        int localPort;
        int remotePort;
        int payloadType;
        int clockRate;  //h264=90000; audio=8000
        int fps{25};
    private:
        std::map<std::string,std::string> extraParams;
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

    /*
     * define rtp or rtcp receive callback function
     */
    typedef int (*RcvCb)(const uint8_t *buf, int len, int marker, void *user);
    typedef void (*RtcpRcvCb)(void* rtcpPacket,void* user);

    /*
     * rtcp receive callback struct
     */
    struct RtcpRcvCbData{
        RtcpRcvCb cb{nullptr};
        void*   user{nullptr};
    };


    class RtpSessionMpl {
    public:
        /*
         * finish initializing list
         */
        RtpSessionMpl() : m_bStopFlag(false){}

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


        /*
         * Send origin rtcp data.provide default function for disable rtcp.
         * the user should pack the rtcp packet by self
         * @param [in] buf:the cache to store data.you should alloc memory by yourself before calling
         * @param [in] len:the len you expect
         * @return the len of real send
         */
//        virtual int SendRtcpData(const uint8_t* buf,int len){return 0;}

        /*
         * Send rtcp app data.provide default function for disable rtcp
         * @param [in] subType:the subType of app packet
         * @param [in] name:the name of app packet
         * @param [in] appData:the data of app packet
         * @param [in] appDataLen:the data length of app packet
         *  @return the len of real send
         */
        virtual int SendRtcpAppData(uint8_t subType,const uint8_t name[4],const void* appData,int appDataLen){return 0;}

        /*
         * Register rtcp receive callback function.
         * @param [in] type:rtcp type
         * @param [in] cb:handler
         * @return true if success or false
         */
        inline bool RegisterRtcpRcvCb(int type,RtcpRcvCb cb,void* user){
            if(type>=RTCP_PACKET_SIZE || type<0){
                std::cout<<"The type is invalid."<<std::endl;
                return false;
            }
            m_rtcpRcvCbDataArr[type].cb=cb;
            m_rtcpRcvCbDataArr[type].user=user;

            return true;
        }

        /*
         * GetRtcpRcvCbData
         * @param [in] type:rtcp type
         * @return the callback function
         */
        RtcpRcvCbData* GetRtcpRcvCbData(int t){return t<RTCP_PACKET_SIZE? &(m_rtcpRcvCbDataArr[t]): nullptr;}

        /*
         * rtcp packet without unpacking
         * the user should unpack including different type by self
         */
        inline uint8_t* GetPacketData(void* rtcpPacket){
            RtcpPacket* p=static_cast<RtcpPacket*>(rtcpPacket);
            return p ? p->data: nullptr;
        }
        inline int GetPacketDataLength(void* rtcpPacket){
            RtcpPacket* p=static_cast<RtcpPacket*>(rtcpPacket);
            return p ? p->dataLen: 0;
        }

        /*
         * app packet.user can get different fields by call function as follow
         */
        inline uint8_t* GetAppData(void* rtcpPacket){
            RtcpAppPacket* p=static_cast<RtcpAppPacket*>(rtcpPacket);
            return p ? p->appData:nullptr;
        }
        inline int GetAppDataLength(void* rtcpPacket){
            RtcpAppPacket* p=static_cast<RtcpAppPacket*>(rtcpPacket);
            return p ? p->appDataLen:0;
        }
        inline uint8_t* GetAppName(void* rtcpPacket) {
            RtcpAppPacket* p=static_cast<RtcpAppPacket*>(rtcpPacket);
            return p ? p->name: nullptr;
        }
        inline uint32_t GetAppSsrc(void* rtcpPacket){
            RtcpAppPacket* p=static_cast<RtcpAppPacket*>(rtcpPacket);
            return p ? p->ssrc: 0;
        }
        inline uint8_t GetAppSubType(void* rtcpPacket){
            RtcpAppPacket* p=static_cast<RtcpAppPacket*>(rtcpPacket);
            return p ? p->subType: 0;
        }




    protected:
        std::atomic_bool    m_bStopFlag;
        RtpHeaderData       m_rtpHeaderData;

        RtcpRcvCbData       m_rtcpRcvCbDataArr[RTCP_PACKET_SIZE];

    };


}//namespace iRtp


//#ifdef __cplusplus
//}
//#endif

#endif //IRTP_RTPSESSIONMPL_H
