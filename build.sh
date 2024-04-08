#!/bin/bash

sudo apt install golang-go -y
sudo apt install python3 -y
sudo apt install g++-mingw-w64-x86-64-win32 -y
# requirements for gtk3
sudo apt-get install libgirepository1.0-dev libgtk-3-dev libcairo-gobject2 gir1.2-freedesktop -y
sudo apt-get install libglib2.0-dev
sudo apt install gtk2-engines-murrine

get go version
'go version go1.18.1 linux/amd64'

change the mod file to match

cd into Team-Server
go get
go build 

then create symlink to path

cd into Operator 
go get
go build

then create symlink to path