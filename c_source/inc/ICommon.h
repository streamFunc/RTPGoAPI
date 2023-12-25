//
// Created by sean on 2023/8/2.
//

#ifndef IRTP_ICOMMON_H
#define IRTP_ICOMMON_H
#include <iostream>
#include <chrono>

#define LOG_FIXED_HEADER() ({std::string strRet= \
        std::move("["+std::string(basename(const_cast<char*>(__FILE__)))+"-"+std::string(__FUNCTION__ )+"-"+std::to_string(__LINE__)+"]"); \
        strRet;})

namespace iRtp{

static inline std::string TimeStamp(){
    auto now=std::chrono::system_clock::now();
    std::time_t now_c=std::chrono::system_clock::to_time_t(now);
    thread_local std::tm timeInfo;
    localtime_r(&now_c,&timeInfo);
//   struct std::tm* timeInfo=std::localtime(&now_c);
    char buffer[80];
    std::strftime(buffer,80,"%Y-%m-%d %H:%M:%S",&timeInfo);
    return buffer;
}//TimeStamp

}//namespace iRtp


#endif //IRTP_ICOMMON_H
