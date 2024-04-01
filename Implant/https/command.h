#ifndef cmd
#define cmd

#include <string>
#include <Windows.h>

namespace wincmd {
bool execute_cmd(xhttp::CommandQueue& queue, std::string& current_dir, DWORD time_out); // const std::string& current_dir,
std::string current_dir = "."; // Initialize current_dir with a valid directory path
DWORD time_out = 5000; // Initialize time_out with a timeout value in milliseconds

}

#endif