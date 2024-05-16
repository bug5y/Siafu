**Not ready for use**<br>
**Still a work in progress**<br>

# Siafu

Siafu is a C2 written in Go and C++ consisting of three main components:

1. **Operator** (Go): 
    - A graphical user interface to send and receive data from the client/implant.
    - Payload generation
    - Listener creation
    <br>

![Screenshot of the Operator interface](/assets/images/Operator.png)<br>

2. **Team-Server** (Go): 
    - A command and control server thats designed to support multiple operators and implants.<br>

![Screenshot of the Team-Server interface](/assets/images/team-server.png)<br>

3. **Client/Implant** (C++): 
    - Client application.
    - Supports Windows OS.<br>

## Install
A build script is provided for installing Siafu and its dependencies:<br>
<sub> Build script has only been tested on Ubuntu & Kali. Elevated privileges are required to perform "apt" updates and installs. </sub><br>

```
git clone https://github.com/BUG5Y/Siafu.git
cd Siafu
sudo build.sh
```

## Usage
Operator
```
cd /Siafu/Operator
./Operator
```

Team-Server<br>
<sub> Elevated privileges are required to use privileged ports</sub>
```
cd /Siafu/Team-Server
sudo ./Team-Server
```
