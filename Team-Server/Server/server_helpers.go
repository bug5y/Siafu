package Server

import (
	"Team-Server/DB"
	"bytes"
	"encoding/json"
	"fmt"
	"net"
)

func updateConnections(respChan chan<- string) {
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
	resp := buf.String()
	respChan <- string(resp)
}

func addToCommandQueue(commands [][]string, idMask string) {
	fmt.Println("addcmdtoqueue")
	item := CommandQueueItem{
		Commands: commands,
		IDMask:   idMask,
	}
	// mutex
	commandQueue = append(commandQueue, item)
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
			return ipnet.IP.String(), nil
		}
	}
	return "", fmt.Errorf("unable to determine server IP address")
}
