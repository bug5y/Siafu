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
#include <codecvt>
#include <locale>
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

DWORD ver;
DWORD version;

DWORD _getVersion(void)
{
  OSVERSIONINFOEXW osvi;
  osvi.dwOSVersionInfoSize = sizeof(OSVERSIONINFOEXW);

    if (GetVersionExW((OSVERSIONINFOW *)&osvi) != FALSE && osvi.dwPlatformId == VER_PLATFORM_WIN32_NT)
    {
        if (osvi.wProductType == VER_NT_WORKSTATION)
        {
            // Windows 2000 - 5.0
            if (osvi.dwMajorVersion == 5 && osvi.dwMinorVersion == 0) ver = 0x0500;
            // Windows XP -  5.1
            else if (osvi.dwMajorVersion == 5 && osvi.dwMinorVersion == 1) ver = 0x0501;
            // Windows XP Professional x64 Edition - 5.2
            else if (osvi.dwMajorVersion == 5 && osvi.dwMinorVersion == 2) ver = 0x0502;
            // Windows Vista - 6.0
            else if (osvi.dwMajorVersion == 6 && osvi.dwMinorVersion == 0) ver = 0x0600;
            // Windows 7 - 6.1
            else if (osvi.dwMajorVersion == 6 && osvi.dwMinorVersion == 1) ver = 0x0601;
            // Windows 8 - 6.2
            else if (osvi.dwMajorVersion == 6 && osvi.dwMinorVersion == 2) ver = 0x0602;
            // Windows 8.1 - 6.3
            else if (osvi.dwMajorVersion == 6 && osvi.dwMinorVersion == 3) ver = 0x0603;
            // Windows 10 - 10.0
            else if (osvi.dwMajorVersion == 10 && osvi.dwMinorVersion == 0) ver = 0x0A00;
            // Windows 11 - 11.0
            else if (osvi.dwMajorVersion == 10 && osvi.dwMinorVersion == 0 && osvi.dwBuildNumber >= 22000) ver = 0x0B00;
        }
        else if (osvi.wProductType == VER_NT_DOMAIN_CONTROLLER || osvi.wProductType == VER_NT_SERVER)
        {
            // Windows Server 2003 - 5.2, Windows Server 2003 R2 - 5.2, Windows Home Server - 5.2
            if (osvi.dwMajorVersion == 5 && osvi.dwMinorVersion == 2) ver = 0x0502;
            // Windows Server 2008 - 6.0
            else if (osvi.dwMajorVersion == 6 && osvi.dwMinorVersion == 0) ver = 0x0600;
            // Windows Server 2008 R2 - 6.1
            else if (osvi.dwMajorVersion == 6 && osvi.dwMinorVersion == 1) ver = 0x0601;
            // Windows Server 2012 - 6.2
            else if (osvi.dwMajorVersion == 6 && osvi.dwMinorVersion == 2) ver = 0x0602;
            // Windows Server 2012 R2 - 6.3
            else if (osvi.dwMajorVersion == 6 && osvi.dwMinorVersion == 3) ver = 0x0603;
            // Windows Server 2016 - 10.0
            else if (osvi.dwMajorVersion == 10 && osvi.dwMinorVersion == 0) ver = 0x0A00;
            // Windows Server 2019 - 10.0
            else if (osvi.dwMajorVersion == 10 && osvi.dwMinorVersion == 0 && osvi.dwBuildNumber >= 17763) ver = 0x0A00;
        }
    }
    return ver;
}

void _getVersionEx(OSINFO *oi)
{
    memset(oi, 0, sizeof(OSINFO));

    OSVERSIONINFOEXW osvi;
    osvi.dwOSVersionInfoSize = sizeof(OSVERSIONINFOEXW);

    if (GetVersionExW((OSVERSIONINFOW *)&osvi) != FALSE)
    {
        SYSTEM_INFO si;
        GetNativeSystemInfo(&si);

        oi->version = _getVersion();
        oi->sp = (osvi.wServicePackMajor > 0xFF || osvi.wServicePackMinor != 0) ? 0 : LOBYTE(osvi.wServicePackMajor);
        oi->build = osvi.dwBuildNumber > 0xFFFF ? 0 : LOWORD(osvi.dwBuildNumber);
        oi->architecture = si.wProcessorArchitecture;
    }
}

std::vector<std::string> getIPs() {
    std::vector<std::string> machineIPs;
    // Implementation to get all ips
    return machineIPs;
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

    std::wstring usernameWStr(domainNameBuffer.begin(), domainNameBuffer.end());
    usernameWStr += L"\\"; // Adding backslash between domain name and username
    usernameWStr.insert(usernameWStr.end(), userNameBuffer.begin(), userNameBuffer.end());

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
                    sockaddr_in *sin = reinterpret_cast<sockaddr_in*>(sa);
                    char buffer[INET_ADDRSTRLEN];
                    inet_ntop(AF_INET, &(sin->sin_addr), buffer, INET_ADDRSTRLEN);
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

std::string buildUID() {
    OSINFO osInfo;
    ULONG ulFlags = GAA_FLAG_INCLUDE_PREFIX;
    ULONG ulFamily = AF_UNSPEC;
    unsigned char* pszAddress = nullptr;

    _getVersionEx(&osInfo);
    _getVersion();

    std::string hostname = getHostname();
    std::string execUsername = getUsername();
    std::string concatString = hostname + "-" + execUsername + "-" +std::to_string(ver) + "-";
    std::vector<std::string> ipAddresses = getIP();
        for (const auto& address : ipAddresses) {
            concatString += "," + address;
        }
    std::cout << "Hostname: " << hostname << std::endl;
    std::cout << "Executable Username: " << execUsername << std::endl;
    std::cout << "Concat: " << concatString << std::endl;
    // Convert the string to a byte array
    std::vector<Bytef> data(concatString.begin(), concatString.end());

    misc::uid = concatString;
    return misc::uid;
}

}
