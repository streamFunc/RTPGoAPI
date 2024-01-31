//
// Created by sean on 2024/1/11.
//

#ifndef IRTP_RTCPPACKET_H
#define IRTP_RTCPPACKET_H
#include "ICommon.h"
//#include <string.h>
//#include <list>

//#include "rtcpsdespacket.h"

namespace iRtp{



class RtcpPacket{
public:
    RtcpPacket():data(nullptr),dataLen(0),ssrc(0){}
    uint8_t* data; //reference.
    int dataLen;
    uint32_t ssrc;
};

class RtcpAppPacket:public RtcpPacket{
public:
    RtcpAppPacket():appData(nullptr),appDataLen(0),name(nullptr),subType(0){}
    uint8_t* appData;
    int appDataLen;
    uint8_t* name;
    uint8_t subType;
};

class RtcpSdesPacket:public RtcpPacket { //just for one item but can be any item type
public:
   RtcpSdesPacket():itemData(nullptr),itemType(0),itemDataLen(0){}
    uint8_t* itemData;
    int itemType;
    int itemDataLen;

}; //RtcpSdesPacket

class RtcpSdesPrivatePacket:public RtcpSdesPacket{
public:
    RtcpSdesPrivatePacket():prefixData(nullptr),prefixDataLength(0),valueData(nullptr),valueDataLength(0){}
    uint8_t* prefixData;
    int prefixDataLength;
    uint8_t* valueData;
    int valueDataLength;
};

class RtcpRRPacket:public RtcpPacket{
public:
    RtcpRRPacket():fractionLost(0),lostPacketNumber(0),extendedHighestSequenceNumber(0)
        ,jitter(0),lastSR(0),delaySinceLastSR(0){}
    uint8_t fractionLost;
    uint32_t lostPacketNumber;
    uint32_t extendedHighestSequenceNumber;
    uint32_t jitter;
    uint32_t lastSR;
    uint32_t delaySinceLastSR;
};
class RtcpSRPacket:public RtcpPacket{ //one report block
public:
    RtcpSRPacket():ntpLSWTimeStamp(0),ntpMSWTimeStamp(0),rtpTimeStamp(0),senderPacketCount(0),senderOctetCount(0){}
    uint32_t ntpLSWTimeStamp;
    uint32_t ntpMSWTimeStamp;
    uint32_t rtpTimeStamp;
    uint32_t senderPacketCount;
    uint32_t senderOctetCount;
    // the report block is the same as the RR

};
class RtcpByePacket:public RtcpPacket{
public:
    RtcpByePacket():reasonData(nullptr),reasonDataLength(0){}
    uint8_t* reasonData;
    uint8_t  reasonDataLength;
};
class RtcpUnknownPacket:public RtcpPacket{ //origin packet.one or more
public:
    RtcpUnknownPacket():unKnownType(0){}
    uint8_t unKnownType;

};



/*
 * further specific when necessary
 */
//class RtcpSdesPacket:public RtcpPacket{ //just for one item but can be any item type
//public:
//    RtcpSdesPacket(): RtcpPacket(RTCP_PACKET_SDES){}
//    ~RtcpSdesPacket(){
//        for(auto itr=privItems.begin();itr!=privItems.end();++itr){
//            if((*itr)!=nullptr){
//                delete (*itr);
//                (*itr)==nullptr;
//            }
//        }
//        if(!privItems.empty())privItems.clear();
//
//    }
//
//    typedef size_t ISize; //redefine
//    static const uint8_t NUMBER_ITEMS_NON_PRIVATE=7; //total items
//    static const uint8_t MAX_ITEM_TEXT_LENGTH=255; //RFC3550 page38
//
//    /** Identifies the type of an SDES item. */
//    enum ItemType
//    {
//        None=0,	/**< Used when the iteration over the items has finished. */
//        CNAME,	/**< Used for a CNAME (canonical name) item. */
//        NAME,	/**< Used for a NAME item. */
//        EMAIL,	/**< Used for an EMAIL item. */
//        PHONE,	/**< Used for a PHONE item. */
//        LOC,	/**< Used for a LOC (location) item. */
//        TOOL,	/**< Used for a TOOL item. */
//        NOTE,	/**< Used for a NOTE item. */
//        PRIV,	/**< Used for a PRIV item. */
//        Unknown /**< Used when there is an item present, but the type is not recognized. */
//    };
//
//    /*
//     * non private item interface of setting
//     */
//    inline ISize SetCNAME(const uint8_t* s,ISize len){return setNonPrivateItem(CNAME,s,len);}
//    inline ISize SetName(const uint8_t* s,ISize len){return setNonPrivateItem(NAME,s,len);}
//    inline ISize SetEmail(const uint8_t* s,ISize len){return setNonPrivateItem(EMAIL,s,len);}
//    inline ISize SetPhone(const uint8_t* s,ISize len){return setNonPrivateItem(PHONE,s,len);}
//    inline ISize SetLoc(const uint8_t* s,ISize len){return setNonPrivateItem(LOC,s,len);}
//    inline ISize SetTool(const uint8_t* s,ISize len){return setNonPrivateItem(TOOL,s,len);}
//    inline ISize SetNote(const uint8_t* s,ISize len){return setNonPrivateItem(NOTE,s,len);}
//
//    /*
//     * non private item interface of getting
//     */
//    inline uint8_t* GetCNAME(ISize* len)const{return getNonPrivateItem(CNAME,len);}
//    inline uint8_t* GetName(ISize* len)const{return getNonPrivateItem(NAME,len);}
//    inline uint8_t* GetEmail(ISize* len)const{return getNonPrivateItem(EMAIL,len);}
//    inline uint8_t* GetPhone(ISize* len)const{return getNonPrivateItem(PHONE,len);}
//    inline uint8_t* GetLoc(ISize* len)const{return getNonPrivateItem(LOC,len);}
//    inline uint8_t* GetTool(ISize* len)const{return getNonPrivateItem(TOOL,len);}
//    inline uint8_t* GetCNote(ISize* len)const{return getNonPrivateItem(NOTE,len);}
//
//private:
//    inline ISize setNonPrivateItem(int itemNo,const uint8_t* s,ISize len){
//        if(itemNo>NUMBER_ITEMS_NON_PRIVATE){
//            std::cout<<LOG_FIXED_HEADER()<<"There is out of array"<<std::endl;
//            return 0;
//        }
//
//        return nonPrivateItems[itemNo-1].SetInfo(s,len);
//    }
//    inline uint8_t* getNonPrivateItem(int itemNo,ISize* len) const{
//        if(itemNo>NUMBER_ITEMS_NON_PRIVATE){
//            std::cout<<LOG_FIXED_HEADER()<<"There is out of array"<<std::endl;
//            return nullptr;
//        }
//        return nonPrivateItems[itemNo-1].GetInfo(len);
//    }
//
//    struct SdesItem{
//        SdesItem():str(nullptr),length(0){}
//        ~SdesItem(){
//            if(str){
//                free(str);
//                str= nullptr;
//            }
//        }
//        inline uint8_t* GetInfo(ISize* len) const{
//            *len=length;
//            return str;
//        }
//
//        inline ISize SetInfo(const uint8_t* s,ISize len){return setString(&str,&length,s,len);}
//
//    protected:
//        inline ISize setString(uint8_t** dest,ISize* destLen,const uint8_t* src,ISize srcLen){
//            srcLen= srcLen>MAX_ITEM_TEXT_LENGTH ? MAX_ITEM_TEXT_LENGTH:srcLen;
//
//            uint8_t* temp=(uint8_t*)malloc(sizeof(uint8_t)*srcLen);
//            if(temp==nullptr){
//                std::cout<<LOG_FIXED_HEADER()<<"There is out of memory"<<std::endl;
//                return 0;
//            }
//
//            memcpy(temp,src,srcLen);
//
//            if(*dest)free((*dest));
//
//            *dest=temp;
//            *destLen=srcLen;
//
//            return srcLen;
//
//        }
//
//
//
//    private:
//        uint8_t* str;
//        ISize length;
//
//    };
//
//    SdesItem nonPrivateItems[NUMBER_ITEMS_NON_PRIVATE];
//
//
//    struct SdesPrivateItem:public SdesItem{
//        SdesPrivateItem():prefix(nullptr),prefixLen(0){}
//        ~SdesPrivateItem(){
//            if(prefix){
//                free(prefix);
//                prefix=nullptr;
//            }
//        }
//        inline ISize SetPrefix(const uint8_t* s,ISize len){return setString(&prefix,&prefixLen,s,len);}
//        inline uint8_t* GetPrefix(ISize* len) const{
//            *len=prefixLen;
//            return prefix;
//        }
//
//    private:
//        uint8_t* prefix;
//        ISize prefixLen;
//    };
//
//    std::list<SdesPrivateItem*> privItems;
//    std::list<SdesPrivateItem*>::const_iterator curPrivItem;
//
//
//
//};




}//namespace iRtp
#endif //IRTP_RTCPPACKET_H
