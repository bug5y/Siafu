#!/bin/bash

echo "Installing dependencies" 

Siafu_Base=$(pwd)
echo "Base: $Siafu_Base"
if ! command -v snap &> /dev/null; then
    echo "Please install the latest version of Snap"
    exit 1
fi

go_version=$(go version | awk '{print $3}' | sed 's/go//')
major_version=$(echo "$go_version" | cut -d'.' -f1)
minor_version=$(echo "$go_version" | cut -d'.' -f2)

if [ "$major_version" -eq 1 ] && [ "$minor_version" -lt 21 ]; then
	sudo snap install go --classic
fi
# rm /path/to/symlink
# Install Python3
sudo apt install python3 -y

# Install MinGW-W64 compiler
sudo apt install g++-mingw-w64-x86-64-win32 -y

# Install GTK3 dependencies
sudo apt-get install libgirepository1.0-dev libgtk-3-dev libcairo-gobject2 gir1.2-freedesktop -y
sudo apt-get install libglib2.0-dev -y
sudo apt install gtk2-engines-murrine -y

echo "Building Team-Server"
echo "This may take a minute"
# Build and run the Team-Server application
cd "$Siafu_Base/Team-Server"
#sed -i 's#var SiafuBase = "/path/to/siafu"#var SiafuBase = "'"$Siafu_Base"'"#' UI/ui.go

go get
go build -o Team-Server
# Allow to bind to privileged ports
sudo setcap 'cap_net_bind_service=+ep' "$Siafu_Base/Team-Server/Team-Server"
#ln -s "$Siafu_Base/Team-Server/Team-Server" /usr/local/bin/Team-Server

echo "Building Operator"
echo "This may take a minute"
cd "$Siafu_Base/Operator"

#sed -i 's#var SiafuBase = "/path/to/siafu"#var SiafuBase = "'"$Siafu_Base"'"#' UI/ui.go

go get
go build -o Operator
#ln -s "$Siafu_Base/Operator/Operator" /usr/local/bin/Operator

#echo "Starting Team-Server..."
#/usr/local/bin/Team-Server &

#echo "Starting Operator..."
#/usr/local/bin/Operator &

