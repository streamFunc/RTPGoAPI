//
// Created by sean on 2023/8/3.
//

#ifndef IRTP_RTPSESSIONMPL_H
#define IRTP_RTPSESSIONMPL_H

#include <atomic>
#include <map>
#include <thread>
#include <condition_variable>
#include "RtcpPacket.h"
#include "ICommon.h"

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
    typedef RcvCb RtpRcvCb;
    typedef void (*RtcpRcvCb)(void* rtcpPacket,void* user);

    struct RtpRcvCbData{
        RcvCb cb{nullptr};
        void* user{nullptr};

        enum CallbackType{
            ONLY_PAYLOAD=0,
            WHOLE_PACKET,
            SIZE
        };
    };


    /*
     * rtcp receive callback struct
     */
    struct RtcpRcvCbData{
        RtcpRcvCb cb{nullptr};
        void*   user{nullptr};

        enum CallBackType{
            APP_PACKET=0,
            RECEIVER_REPORT,
            SENDER_REPORT,
            SDES_ITEM,
            SDES_PRIVATE_ITEM,
            BYE_PACKET,
            ORIGIN,
            UNKNOWN,
            SIZE
        };

    };

    static const int  RTCP_MAX_CALLBACK_ITEM_SIZE=RtcpRcvCbData::SIZE;
    static const int  RTP_MAX_CALLBACK_ITEM_SIZE=RtpRcvCbData::SIZE;

    class RtpSessionMpl {
    public:
        /*
         * finish initializing list
         */
        RtpSessionMpl() : m_bStopFlag(false),m_pThread(nullptr),m_isWaking(false){}

        /*
         * it will do nothing. just to ensure that inherit object pointer or reference run destructor function
         * */
        virtual ~RtpSessionMpl() {}

        /*
         * initialize something such as ip,port ,payloadType and so on
         * */
        virtual bool Init(const RtpSessionInitData *pInitData) = 0;


        /*
         * start session
         */
        virtual bool Start(){return true;}

        /*
         * initialize thread and enter loop which inherit decide specific action
         * notice:dont block caller thread and return immediately
         * */
        bool Loop() {
            if(m_pThread)Stop(); //if exist then stop and delete

            bool haveTask=false;
            for(int i=0;i<RTP_MAX_CALLBACK_ITEM_SIZE;i++){
                if(m_rtpRcvCbDataArr[i].cb!= nullptr){
                    haveTask=true;
                    break;
                }
            }

            if(!haveTask){
                std::cerr<<LOG_FIXED_HEADER()<<"There is not a rtp receive callback function in the array."<<std::endl;
                return false;
            }

            m_pThread=new std::thread(&RtpSessionMpl::loop,this);
            m_isWaking=true;

            return true;

        };


        /*
         * stop rtp schedule task and handle inherit stop function.
         * */
        bool Stop(){
            m_bStopFlag=true;
            tryToWakeUp();

            if(m_pThread){
                std::this_thread::sleep_for(std::chrono::nanoseconds (1)); //ns out of the piece of time
                if(m_pThread->joinable())m_pThread->join();
                delete m_pThread;
                m_pThread=nullptr;
            }

            return stop(); //caller thread should inherit stop
        }

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
         * @param [in] appDataLen:the data length of app packet.it must be a multiple of 32 bits long
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
            if(type>=RTCP_MAX_CALLBACK_ITEM_SIZE || type<0){
                std::cout<<"The type is invalid."<<std::endl;
                return false;
            }
            m_rtcpRcvCbDataArr[type].cb=cb;
            m_rtcpRcvCbDataArr[type].user=user;

            return true;
        }

        /*
       * Register rtp receive callback function.
       * @param [in] type:rtcp type
       * @param [in] cb:handler
       * @return true if success or false
       */
        inline bool RegisterRtpRcvCb(int type,RcvCb cb,void* user){
            if(type>=RTCP_MAX_CALLBACK_ITEM_SIZE || type<0){
                std::cout<<"The type is invalid."<<std::endl;
                return false;
            }
            m_rtpRcvCbDataArr[type].cb=cb;
            m_rtpRcvCbDataArr[type].user=user;

            return true;
        }

        /*
         * GetRtcpRcvCbData
         * @param [in] type:rtcp type
         * @return the callback function
         */
        RtcpRcvCbData* GetRtcpRcvCbData(int t){return t<RTCP_MAX_CALLBACK_ITEM_SIZE ? &(m_rtcpRcvCbDataArr[t]):nullptr;}

        /*
         * GetRtpRcvCbData
         * @param [in] type:rtp type
         * @return the callback function
         */
        RtcpRcvCbData* GetRtpRcvCbData(int t){return t<RTP_MAX_CALLBACK_ITEM_SIZE ? &(m_rtcpRcvCbDataArr[t]):nullptr;}

        /*
         * rtcp packet without unpacking
         * the user should unpack including different type by self
         */
        inline uint8_t* GetPacketData(RtcpPacket* p)const{
            return p ? p->data: nullptr;
        }
        inline int GetPacketDataLength(RtcpPacket* p)const{
            return p ? p->dataLen: 0;
        }
        inline uint32_t  GetSSRC(RtcpPacket* p)const{
            return p ? p->ssrc: 0;
        }

        /*
         * app packet.user can get different fields by call function as follow
         */
        inline uint8_t* GetAppData(RtcpPacket* rtcpPacket)const{
            RtcpAppPacket* p=static_cast<RtcpAppPacket*>(rtcpPacket);
            return p ? p->appData:nullptr;
        }
        inline int GetAppDataLength(RtcpPacket* rtcpPacket)const{
            RtcpAppPacket* p=static_cast<RtcpAppPacket*>(rtcpPacket);
            return p ? p->appDataLen:0;
        }
        inline uint8_t* GetAppName(RtcpPacket* rtcpPacket)const{
            RtcpAppPacket* p=static_cast<RtcpAppPacket*>(rtcpPacket);
            return p ? p->name: nullptr;
        }
        inline uint32_t GetAppSsrc(RtcpPacket* rtcpPacket)const{
            RtcpAppPacket* p=static_cast<RtcpAppPacket*>(rtcpPacket);
            return p ? p->ssrc: 0;
        }
        inline uint8_t GetAppSubType(RtcpPacket* rtcpPacket)const{
            RtcpAppPacket* p=static_cast<RtcpAppPacket*>(rtcpPacket);
            return p ? p->subType: 0;
        }

        /*
         * sdes item;user can get what they need for one item
         */
        inline uint8_t* GetSdesItemData(RtcpPacket* rp)const{
            RtcpSdesPacket* p=static_cast<RtcpSdesPacket*>(rp);
            return p ? p->itemData : nullptr;
        }
        inline int GetSdesItemDataLen(RtcpPacket* rp)const{
            RtcpSdesPacket* p=static_cast<RtcpSdesPacket*>(rp);
            return p ? p->itemDataLen : 0;
        }
        inline int GetSdesItemType(RtcpPacket* rp)const{
            RtcpSdesPacket* p=static_cast<RtcpSdesPacket*>(rp);
            return p ? p->itemType : 0;
        }


        /*
         * sdes private item
         */
        inline uint8_t* GetSdesPrivatePrefixData(RtcpPacket* rp)const{
            RtcpSdesPrivatePacket* p=static_cast<RtcpSdesPrivatePacket*>(rp);
            return p ? p->prefixData : nullptr;
        }
        inline int GetSdesPrivatePrefixDataLen(RtcpPacket* rp)const{
            RtcpSdesPrivatePacket* p=static_cast<RtcpSdesPrivatePacket*>(rp);
            return p ? p->prefixDataLength: 0;
        }
        inline uint8_t* GetSdesPrivateValueData(RtcpPacket* rp)const{
            RtcpSdesPrivatePacket* p=static_cast<RtcpSdesPrivatePacket*>(rp);
            return p ? p->valueData: nullptr;
        }
        inline int GetSdesPrivateValueDataLen(RtcpPacket* rp)const{
            RtcpSdesPrivatePacket* p=static_cast<RtcpSdesPrivatePacket*>(rp);
            return p ? p->valueDataLength: 0;
        }

        /*
         * Bye packet
         */
        inline uint8_t* GetByeReasonData(RtcpPacket* rp)const{
           RtcpByePacket* p=static_cast<RtcpByePacket*>(rp);
           return p ? p->reasonData: 0;
        }
        inline int GetByeReasonDataLen(RtcpPacket* rp)const{
            RtcpByePacket* p=static_cast<RtcpByePacket*>(rp);
            return p ? p->reasonDataLength: 0;
        }

        /*
         * unKnown packet
         */
        inline uint8_t  GetUnknownPacketType(RtcpPacket* rp)const{
            RtcpUnknownPacket* p=static_cast< RtcpUnknownPacket*>(rp);
            return p ? p->unKnownType: 0;
        }
        inline uint8_t* GetUnKnownRtcpPacketData(RtcpPacket* rp) const{
            return rp ? rp ->data :nullptr;
        }
        inline int GetUnKnownRtcpPacketDataLen(RtcpPacket* rp)const{
            return rp ? rp->dataLen : 0;
        }
        inline uint32_t GetUnKnownRtcpPacketSsrc(RtcpPacket* rp)const{
            return rp ? rp->ssrc : 0 ;
        }

        /*
         * RR packet
         */
        inline uint8_t GetRRFractionLost(RtcpPacket* rp)const{
            RtcpRRPacket* p=static_cast<RtcpRRPacket*>(rp);
            return p ? p->fractionLost: 0;
        }
        inline uint32_t GetRRLostPacketNumber(RtcpPacket* rp)const{
            RtcpRRPacket* p=static_cast<RtcpRRPacket*>(rp);
            return p ? p->lostPacketNumber: 0;
        }
        inline uint32_t GetRRExtendedHighestSequenceNumber(RtcpPacket* rp)const{
            RtcpRRPacket* p=static_cast<RtcpRRPacket*>(rp);
            return p ? p->extendedHighestSequenceNumber: 0;
        }
        inline uint32_t GetRRJitter(RtcpPacket* rp)const{
            RtcpRRPacket* p=static_cast<RtcpRRPacket*>(rp);
            return p ? p->jitter: 0;
        }
        inline uint32_t GetRRLastSR(RtcpPacket* rp)const{
            RtcpRRPacket* p=static_cast<RtcpRRPacket*>(rp);
            return p ? p->lastSR: 0;
        }
        inline uint32_t GetRRDelaySinceLastSR(RtcpPacket* rp)const{
            RtcpRRPacket* p=static_cast<RtcpRRPacket*>(rp);
            return p ? p->delaySinceLastSR: 0;
        }

        /*
         * SR packet
         */
        inline uint32_t GetSRNtpLSWTimeStamp(RtcpPacket* rp)const{
            RtcpSRPacket* p=static_cast<RtcpSRPacket*>(rp);
            return p ? p->ntpLSWTimeStamp: 0;
        }
        inline uint32_t GetSRNtpMSWTimeStamp(RtcpPacket* rp)const{
            RtcpSRPacket* p=static_cast<RtcpSRPacket*>(rp);
            return p ? p->ntpMSWTimeStamp: 0;
        }
        inline uint32_t GetSRRtpTimeStamp(RtcpPacket* rp)const{
            RtcpSRPacket* p=static_cast<RtcpSRPacket*>(rp);
            return p ? p->rtpTimeStamp: 0;
        }
        inline uint32_t GetSRSenderPacketCount(RtcpPacket* rp)const{
            RtcpSRPacket* p=static_cast<RtcpSRPacket*>(rp);
            return p ? p->senderPacketCount: 0;
        }
        inline uint32_t GetSRSenderOctetCount(RtcpPacket* rp)const{
            RtcpSRPacket* p=static_cast<RtcpSRPacket*>(rp);
            return p ? p->senderOctetCount: 0;
        }



    protected:
        virtual void loop()=0;
        virtual bool stop()=0;

        void tryToWakeUp(){
            if(m_isWaking)return;
            m_cv.notify_all();
        }

        void wait(){
            std::unique_lock<std::mutex> lock(m_mutex);
            m_isWaking=false;

            m_cv.wait(lock);
            m_isWaking=true;
        }



        RtpHeaderData       m_rtpHeaderData;

        RtcpRcvCbData       m_rtcpRcvCbDataArr[RTCP_MAX_CALLBACK_ITEM_SIZE];
        RtpRcvCbData        m_rtpRcvCbDataArr[RTP_MAX_CALLBACK_ITEM_SIZE];

        //rtp schedule
        std::atomic_bool        m_bStopFlag;
        std::thread*            m_pThread;
        std::condition_variable m_cv;
        std::mutex              m_mutex;
        std::atomic_bool        m_isWaking;

    };


}//namespace iRtp


//#ifdef __cplusplus
//}
//#endif

#endif //IRTP_RTPSESSIONMPL_H
