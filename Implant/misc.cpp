#include <vector>
#include <winsock2.h>
#include <windows.h>
#include <iostream>
#include "misc.h"
#include <string>
#include <iphlpapi.h>
#include <ws2tcpip.h>
#include <random>
#include "zlib\zlib.h"

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

/*
std::string versionString;

std::string getVersionString(DWORD version) {
    if (version == 0x0500) versionString = "Windows-2000";
    else if (version == 0x0501) versionString = "Windows-XP";
    else if (version == 0x0502) versionString = "Windows-XP-Professional";
    else if (version == 0x0600) versionString = "Windows-Vista";
    else if (version == 0x0601) versionString = "Windows-7";
    else if (version == 0x0602) versionString = "Windows-8";
    else if (version == 0x0603) versionString = "Windows-8.1";
    else if (version == 0x0A00) versionString = "Windows-10";
    else if (version == 0x0B00) versionString = "Windows-11";
    else if (version == 0x0503) versionString = "Windows-Server-2003-or-Windows-Home-Server";
    else if (version == 0x0604) versionString = "Windows-Server-2008";
    else if (version == 0x0605) versionString = "Windows-Server-2008-R2";
    else if (version == 0x0606) versionString = "Windows-Server-2012";
    else if (version == 0x0607) versionString = "Windows-Server-2012-R2";
    else if (version == 0x0A08) versionString = "Windows-Server-2016";
    else if (version == 0x0A09) versionString = "Windows-Server-2019";
    else versionString = "Unknown";

    std::cout << "Version: " << versionString << std::endl;
    return versionString;
}
*/

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


int getMAC(ULONG ulFlags, ULONG ulFamily, unsigned char** pszAddress)
{
        PIP_ADAPTER_ADDRESSES pCurrAddresses = NULL;
        PIP_ADAPTER_ADDRESSES pAddresses = NULL;

        int nAddressCount = 0;
        DWORD dwRetVal = 0;
        ULONG ulBufLen = sizeof(IP_ADAPTER_ADDRESSES);
        HANDLE hHeap = NULL;

        hHeap = GetProcessHeap();
        pAddresses = (PIP_ADAPTER_ADDRESSES)HeapAlloc(hHeap, 0x00, ulBufLen);
        if (pAddresses == NULL) {
               return 0;
        }      

        dwRetVal = GetAdaptersAddresses(ulFamily, ulFlags, NULL, pAddresses, &ulBufLen);
        if (dwRetVal == ERROR_BUFFER_OVERFLOW)
        {
               HeapFree(hHeap, 0x00, pAddresses);
               pAddresses = (PIP_ADAPTER_ADDRESSES)HeapAlloc(hHeap, 0x00, ulBufLen);
        }

        if (pAddresses == NULL) {
               return 0;
        }      

        dwRetVal = GetAdaptersAddresses(ulFamily, ulFlags, NULL, pAddresses, &ulBufLen);
        if (dwRetVal == NO_ERROR)
        {
               pCurrAddresses = pAddresses;
               while (pCurrAddresses)
               {
                       pCurrAddresses = pCurrAddresses->Next;
                       ++nAddressCount;
               }

               *pszAddress = (unsigned char*)HeapAlloc(hHeap, 0x00, MAX_ADAPTER_ADDRESS_LENGTH * nAddressCount);
               pCurrAddresses = pAddresses;
               nAddressCount = 0;
               while (pCurrAddresses)
               {
                       RtlCopyMemory(*pszAddress + (MAX_ADAPTER_ADDRESS_LENGTH * nAddressCount++),
                          pCurrAddresses->PhysicalAddress,
                                               MAX_ADAPTER_ADDRESS_LENGTH);
                       pCurrAddresses = pCurrAddresses->Next;
               }
        }

         if (pAddresses) {
               HeapFree(hHeap, 0x00, pAddresses);
               pAddresses = NULL;
        }
        return nAddressCount;
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

    _getVersionEx(&osInfo);
    _getVersion();

    std::string randomString = generateRandomString(8);
    std::string concatString = randomString + "-" + std::to_string(ver);
    std::cout << "Concat: " << concatString << std::endl;
    // Convert the string to a byte array
    std::vector<Bytef> data(concatString.begin(), concatString.end());

    misc::uid = concatString;
    std::cout << "UID: " << misc::uid << std::endl;
    return misc::uid;
}

}
