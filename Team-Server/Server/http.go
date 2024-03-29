package Server 

import (
	"strings"
    "sync"
	"fmt"
	"log"
	"net/http"
	"Team-Server/UI"
)

var opMutex sync.Mutex

var httpMutex sync.Mutex

var prevLength = 0

var ImplantMap = make(map[string]int)
var green = "\033[32m"
var reset = "\033[0m"
var red = "\033[31m"
var operatorPort = "8443"

type CommandQueueItem struct {
    Commands [][]string
    IDMask   int
}

var commandQueue []CommandQueueItem

type Command struct {
	Group    string `json:"Group"`
	String   string `json:"String"`
	Response string `json:"Response"`
}

func OperatorServer() {
    opMutex.Lock()
    defer opMutex.Unlock()

    http.HandleFunc("/operator", handleOperator)
    http.HandleFunc("/info", handleServerCMDs)
    serverIP, err := getServerIP()
    if err != nil {
        log.Fatal("Error:", err)
    }
    // Get the server's IP address
    if serverIP == "" {
        log.Fatal("Failed to determine server IP address")
    }
    
    operatorPort = ":" + operatorPort
	
    output := "Server started at http://" + serverIP + operatorPort

    UI.InsertLogMarkup(output)

    // Server configuration
	server := http.Server{
		Addr:      operatorPort,
	}

	// Serve
    err = server.ListenAndServe()
    defer server.Close()
}

func createListener(cmdString string, respChan chan<- string) {
    // proto + "," + ip + "," +  port
    parts := strings.Split(cmdString, ",")
    if len(parts) != 3 {
        return
    }

    listenerProto := parts[0]
    listenerIP := parts[1]
    listenerPort := parts[2]

    // Handling different protocols
    switch strings.ToUpper(listenerProto) {
    case "HTTP":
        resp := httpListener(listenerIP, listenerPort)
        respChan <- resp
    case "HTTPS":
        resp := httpsListener(listenerIP, listenerPort)
        respChan <- resp
    default:
        // Invalid protocol
        resp := fmt.Sprintf("Invalid protocol: %s", listenerProto)
        respChan <- resp
    }    
}


func httpListener(listenerIP string, listenerPort string) string {
    httpMutex.Lock()
    defer httpMutex.Unlock()

    http.HandleFunc("/implant", handleImplant)

	// Server configuration
    server := &http.Server{
        Addr:    listenerIP + ":" + listenerPort,
    }

    go func() {
        if err := server.ListenAndServe(); err != nil {
            output := fmt.Sprintf("Error starting HTTP listener: %s\n", err)
            UI.InsertLogMarkup(output)
        }
    }()

    output := "Listener started at http://" + listenerIP + ":" + listenerPort
    UI.InsertLogMarkup(output)

    status := "listener started"
    return status
}

func httpsListener(listenerIP string, listenerPort string) string {
    /*
    https.HandleFunc("/implant", handleImplant)

	// Server configuration
    server := &https.Server{
        Addr:    listenerIP + ":" + listenerPort,
    }

    // Attempt to start the server
    err := server.ListenAndServe()
    if err != nil {
        status := "Error starting HTTPS listener:", err
        return status
    }

    fmt.Printf("Listener started at https://%s:%s\n", listenerIP, listenerPort)
    status := "listener started"
    return status
    */
    status := "Not implemented"
    fmt.Println(status)
    return status
}

