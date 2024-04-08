#include <vector>
#include <winsock2.h>
#include <windows.h>
#include <iostream>
#include "misc.h"
#include <string>
#include <iphlpapi.h>
#include <ws2tcpip.h>
#include <random>
#include "zlib/zlib.h"
#include <securitybaseapi.h>
#include <random>
#include "base64.hpp"
#include <picosha2.h>
#include <codecvt>
#include <locale>
#include <iomanip>
#pragma comment(lib, "Advapi32.lib")
#pragma comment(lib, "Kernel32.lib")
#pragma comment(lib, "iphlpapi.lib")
#pragma comment(lib, "ws2_32.lib")

namespace misc {
// Define OSINFO structure
struct OSINFO {
    DWORD version;
    DWORD sp;
    DWORD build;
    WORD architecture;
};

std::string uid;
std::string truncatedHash;
DWORD ver;
DWORD version;

DWORD GetWindowsVersion()
{
    OSVERSIONINFOEX osvi;
    DWORD_PTR dwlConditionMask = 0;
    BYTE op = VER_GREATER_EQUAL;

    ZeroMemory(&osvi, sizeof(OSVERSIONINFOEX));
    osvi.dwOSVersionInfoSize = sizeof(OSVERSIONINFOEX);
    osvi.dwMajorVersion = 5;

    VER_SET_CONDITION(dwlConditionMask, VER_MAJORVERSION, op);

    DWORD dwlTypeBitsMask = VER_NT_WORKSTATION;
    BOOL bIsWindowsClient = VerifyVersionInfo(&osvi, VER_MAJORVERSION, dwlConditionMask);

    if (bIsWindowsClient)
    {
        if (osvi.dwMajorVersion == 10 && osvi.dwMinorVersion == 0 && osvi.dwBuildNumber >= 22000)
            return 0x1B00; // Windows 11
        else if (osvi.dwMajorVersion == 10 && osvi.dwMinorVersion == 0)
            return 0x1A00; // Windows 10
        else if (osvi.dwMajorVersion == 6 && osvi.dwMinorVersion == 3)
            return 0x1603; // Windows 8.1
        else if (osvi.dwMajorVersion == 6 && osvi.dwMinorVersion == 2)
            return 0x1602; // Windows 8
        else if (osvi.dwMajorVersion == 6 && osvi.dwMinorVersion == 1)
            return 0x1601; // Windows 7
        else if (osvi.dwMajorVersion == 6 && osvi.dwMinorVersion == 0)
            return 0x1600; // Windows Vista
        else if (osvi.dwMajorVersion == 5 && osvi.dwMinorVersion == 2)
            return 0x1502; // Windows XP Professional x64 Edition
        else if (osvi.dwMajorVersion == 5 && osvi.dwMinorVersion == 1)
            return 0x1501; // Windows XP
        else if (osvi.dwMajorVersion == 5 && osvi.dwMinorVersion == 0)
            return 0x1500; // Windows 2000
    }
    else
    {
        dwlTypeBitsMask = VER_NT_SERVER;
        if (osvi.dwMajorVersion == 10 && osvi.dwMinorVersion == 0 && osvi.dwBuildNumber >= 20348)
            return 0x2A01; // Windows Server 2022
        else if (osvi.dwMajorVersion == 10 && osvi.dwMinorVersion == 0 && osvi.dwBuildNumber >= 17763)
            return 0x2A02; // Windows Server 2019
        else if (osvi.dwMajorVersion == 10 && osvi.dwMinorVersion == 0 && osvi.dwBuildNumber >= 14393)
            return 0x2A03; // Windows Server 2016
        else if (osvi.dwMajorVersion == 6 && osvi.dwMinorVersion == 3)
            return 0x2603; // Windows Server 2012 R2
        else if (osvi.dwMajorVersion == 6 && osvi.dwMinorVersion == 2)
            return 0x2602; // Windows Server 2012
        else if (osvi.dwMajorVersion == 6 && osvi.dwMinorVersion == 1)
            return 0x2601; // Windows Server 2008 R2
        else if (osvi.dwMajorVersion == 6 && osvi.dwMinorVersion == 0)
            return 0x2600; // Windows Server 2008
        else if (osvi.dwMajorVersion == 5 && osvi.dwMinorVersion == 2)
            return 0x2502; // Windows Server 2003 R2
    }

    return 0; // Unknown version
}

std::string truncatedSHA256(const std::string& uid) {
    std::vector<unsigned char> hash(picosha2::k_digest_size);
    picosha2::hash256(uid.begin(), uid.end(), hash.begin(), hash.end());

    std::string hex_str = picosha2::bytes_to_hex_string(hash.begin(), hash.end());
    std::string trunc_str = hex_str.substr(0, 8);

    misc::truncatedHash = macaron::Base64::Encode(trunc_str);
    return misc::truncatedHash;
}

std::string getHostname() {
    char hostname[MAX_COMPUTERNAME_LENGTH + 1];
    DWORD hostnameSize = sizeof(hostname);
    if (GetComputerNameExA(ComputerNameDnsFullyQualified, hostname, &hostnameSize) != 0) {
        return std::string(hostname);
    }
    return std::string();
}

std::string getUsername() {
    HANDLE token;
    if (!OpenProcessToken(GetCurrentProcess(), TOKEN_QUERY, &token)) {
        std::cerr << "Failed to open process token. Error code: " << GetLastError() << std::endl;
        return std::string();
    }

    DWORD bufferSize = 0;
    GetTokenInformation(token, TokenUser, NULL, 0, &bufferSize);
    if (bufferSize == 0) {
        std::cerr << "Failed to get token information size. Error code: " << GetLastError() << std::endl;
        CloseHandle(token);
        return std::string();
    }

    std::vector<char> buffer(bufferSize);
    if (!GetTokenInformation(token, TokenUser, buffer.data(), bufferSize, &bufferSize)) {
        std::cerr << "Failed to get token information. Error code: " << GetLastError() << std::endl;
        CloseHandle(token);
        return std::string();
    }

    CloseHandle(token);

    TOKEN_USER* tokenUser = reinterpret_cast<TOKEN_USER*>(buffer.data());

    DWORD userNameSize = 0;
    DWORD domainNameSize = 0;
    SID_NAME_USE sidNameUse;
    LookupAccountSid(NULL, tokenUser->User.Sid, NULL, &userNameSize, NULL, &domainNameSize, &sidNameUse);
    if (userNameSize == 0) {
        std::cerr << "Failed to get account name size. Error code: " << GetLastError() << std::endl;
        return std::string();
    }

    std::vector<char> userNameBuffer(userNameSize);
    std::vector<char> domainNameBuffer(domainNameSize);

    if (!LookupAccountSid(NULL, tokenUser->User.Sid, userNameBuffer.data(), &userNameSize, domainNameBuffer.data(), &domainNameSize, &sidNameUse)) {
        std::cerr << "Failed to lookup account name. Error code: " << GetLastError() << std::endl;
        return std::string();
    }

    std::wstring usernameWStr(userNameBuffer.begin(), userNameBuffer.end());
    std::wstring_convert<std::codecvt_utf8<wchar_t>> converter;
    return converter.to_bytes(usernameWStr);
}

std::vector<std::string> getIP() {
    ULONG outBufLen = 0;
    DWORD dwRetVal = 0;
    std::vector<std::string> addresses;

    // Define pointers to variables needed to hold adapter information
    PIP_ADAPTER_ADDRESSES pAddresses = NULL;
    PIP_ADAPTER_ADDRESSES pCurrAddresses = NULL;

    // Allocate memory to retrieve adapter information
    outBufLen = sizeof(IP_ADAPTER_ADDRESSES);
    pAddresses = (IP_ADAPTER_ADDRESSES *) malloc(outBufLen);
    if (pAddresses == NULL) {
        std::cerr << "Memory allocation failed." << std::endl;
        return addresses;
    }

    // Call GetAdaptersAddresses to retrieve adapter information
    if (GetAdaptersAddresses(AF_UNSPEC, GAA_FLAG_INCLUDE_PREFIX, NULL, pAddresses, &outBufLen) == ERROR_BUFFER_OVERFLOW) {
        free(pAddresses);
        pAddresses = (IP_ADAPTER_ADDRESSES *) malloc(outBufLen);
        if (pAddresses == NULL) {
            std::cerr << "Memory allocation failed." << std::endl;
            return addresses;
        }
    }

    // Call GetAdaptersAddresses again with the allocated memory
    dwRetVal = GetAdaptersAddresses(AF_UNSPEC, GAA_FLAG_INCLUDE_PREFIX, NULL, pAddresses, &outBufLen);
    if (dwRetVal == NO_ERROR) {
        pCurrAddresses = pAddresses;
        while (pCurrAddresses) {
            // Print IP addresses for each adapter
            IP_ADAPTER_UNICAST_ADDRESS *pUnicast = pCurrAddresses->FirstUnicastAddress;
            while (pUnicast) {
                sockaddr *sa = pUnicast->Address.lpSockaddr;
                if (sa->sa_family == AF_INET) {
                    const sockaddr_in* sin = reinterpret_cast<const sockaddr_in*>(sa);
                    char buffer[INET_ADDRSTRLEN];
                    strcpy_s(buffer, INET_ADDRSTRLEN, inet_ntoa(sin->sin_addr));
                    addresses.push_back(buffer);
                }
                pUnicast = pUnicast->Next;
            }
            pCurrAddresses = pCurrAddresses->Next;
        }
    } else {
        std::cerr << "GetAdaptersAddresses failed with error: " << dwRetVal << std::endl;
    }

    // Free allocated memory
    if (pAddresses) {
        free(pAddresses);
    }

    return addresses;
}

std::string generateRandomString(int length) {
    static const char charset[] =
        "0123456789"
        "abcdefghijklmnopqrstuvwxyz"
        "ABCDEFGHIJKLMNOPQRSTUVWXYZ";
    std::mt19937 rng(std::random_device{}());
    std::uniform_int_distribution<int> dist(0, sizeof(charset) - 2);

    std::string result;
    result.reserve(length);
    for (int i = 0; i < length; ++i) {
        result += charset[dist(rng)];
    }
    return result;
}

std::string buildUID() {
    OSINFO osInfo;
    ULONG ulFlags = GAA_FLAG_INCLUDE_PREFIX;
    ULONG ulFamily = AF_UNSPEC;
    unsigned char* pszAddress = nullptr;

    DWORD verdw;

    verdw = GetWindowsVersion();

    std::cout << "Version " << verdw << std::endl; // shows its 0

    std::string hostname = getHostname();
    std::string execUsername = getUsername();
    std::string randomString = generateRandomString(4);
    std::string concatString = randomString + "-" + hostname + "-" + execUsername + "-" +std::to_string(verdw) + "-";
    std::vector<std::string> ipAddresses = getIP();
        for (const auto& address : ipAddresses) {
            concatString += "," + address;
        }

    std::string base64Concat = macaron::Base64::Encode(concatString);
    misc::uid = base64Concat;
    misc::truncatedHash = truncatedSHA256(uid);
    return misc::uid;
}

}
