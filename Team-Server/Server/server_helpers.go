package Server

import (
	"fmt"
	"net"
    "encoding/base64"
	"Team-Server/DB"
    "encoding/json"
    "bytes"
)
var serverIP string

func updateConnections(respChan chan<- string) {
    fmt.Println("updateconnections")
    DB.ConnectionMutex.Lock()
    defer DB.ConnectionMutex.Unlock()
        jsonData, err := json.Marshal(DB.ConnectionLog)
        if err != nil {
            fmt.Println("Error:", err)
            return
        }
        var buf bytes.Buffer
        if _, err := buf.Write(jsonData); err != nil {
            fmt.Println("Error:", err)
            return
        }
        resp := base64.StdEncoding.EncodeToString(buf.Bytes())
        prevLength = len(DB.ConnectionLog)
        respChan <- resp
}

func uint32ToBytes(n uint32) []byte {
    b := make([]byte, 4)
    b[0] = byte(n >> 24)
    b[1] = byte(n >> 16)
    b[2] = byte(n >> 8)
    b[3] = byte(n)
    return b
}

func bytesToUint32(b []byte) uint32 {
    return uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
}

func addToCommandQueue(commands [][]string, idMask string) {
    fmt.Println("addcmdtoqueue")
    item := CommandQueueItem{
        Commands: commands,
        IDMask:   idMask,
    }
    // mutex
    commandQueue = append(commandQueue, item)

    return
}

func getServerIP() (string, error) {
    addrs, err := net.InterfaceAddrs()
    if err != nil {
        return "", fmt.Errorf("failed to retrieve interface addresses: %v", err)
    }

    for _, addr := range addrs {
        ipnet, ok := addr.(*net.IPNet)
        if !ok {
            continue
        }
        if ipnet.IP.IsGlobalUnicast() && ipnet.IP.To4() != nil {
			serverIP = ipnet.IP.String()
            return ipnet.IP.String(), nil
        }
    }
    return "", fmt.Errorf("unable to determine server IP address")
}
