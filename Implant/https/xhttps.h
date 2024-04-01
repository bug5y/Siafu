#ifndef xHTTP_H
#define xHTTP_H

#include <string>
#include <vector>
#include <cstdint>
#include <Winsock2.h>
#include <map>
#include <nlohmann/json.hpp>

namespace xhttp {
    using json = nlohmann::json;
    extern std::string host;
    extern std::string url;
    extern std::string port;
    extern std::vector<uint32_t> req_body;
    extern SOCKET ConnectSocket;
    extern std::string cmdstr;
    extern std::string cmd;
    extern int key;
    extern std::string cmdResponse;
    extern std::string cmdGroup;
    extern std::string cmdString;

    std::string http_get(const std::string& url);
    std::vector<char> receive_data(SOCKET ConnectSocket);
    SOCKET create_socket(const std::string& host, int port);
    std::string extractCMD(const std::vector<char>& data);
    std::string base64_encode(const std::string &in);
    std::string base64_decode(const std::string &in);

struct Queue {
    std::string Group;
    std::string String;
    std::string Response;

    // Default constructor
    Queue() = default;

    // Constructor with arguments
    Queue(const std::string& group, const std::string& str, const std::string& response)
        : Group(group), String(str), Response(response) {}

    // Assignment operator
    Queue& operator=(const Queue& other) {
        if (this != &other) {
            Group = other.Group;
            String = other.String;
            Response = other.Response;
        }
        return *this;
    }
};


    typedef std::map<int, Queue> CommandQueue;

void addToQueue(CommandQueue& queue, int key, const std::string& cmdGroup, const std::string& cmdString, const std::string& cmdResponse);
void removeFromQueue(CommandQueue& queue);

struct cmdStruct {
    std::string Group;
    std::string String;
    std::string Response;
};
extern cmdStruct msg;
}

#endif
