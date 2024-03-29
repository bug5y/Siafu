#!/usr/bin/env python3

import sys
import os

def main():
    if len(sys.argv) != 4:
        print("Usage: python3 builder.py <IP> <port> <protocol>")
        sys.exit(1)

    ip = sys.argv[1]
    port = sys.argv[2]
    proto = sys.argv[3]

    folder_name = "Builder-Output"
    if not os.path.exists(folder_name):
        try:
            os.makedirs(folder_name)
        except OSError as e:
            print(f"Error creating folder '{folder_name}': {e}")
            sys.exit(1)
    elif not os.path.isdir(folder_name):
        print(f"'{folder_name}' exists but is not a directory.")
        sys.exit(2)

    implant_name = f"{proto}_{port}"

    file_path = os.path.join(folder_name, implant_name)
    

    # Write the data to a file
    with open(file_path, "w") as f:
        f.write(f"IP: {ip}\n")
        f.write(f"Port: {port}\n")
        f.write(f"Protocol: {proto}\n")

if __name__ == "__main__":
    main()
