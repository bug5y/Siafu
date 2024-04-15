package Server

import (
	"Team-Server/UI"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
)

var reset = "\033[0m"
var red = "\033[31m"
var operatorPort = "8443"

type CommandQueueItem struct {
	Commands [][]string
	IDMask   string
}

var commandQueue []CommandQueueItem

type Command struct {
	Group    string `json:"Group"`
	String   string `json:"String"`
	Response string `json:"Response"`
}

func OperatorServer() {
	http.HandleFunc("/operator", handleOperator)
	http.HandleFunc("/info", handleServerCMDs)

	serverIP, err := getServerIP()
	if err != nil {
		log.Fatal("Error:", err)
	}

	if serverIP == "" {
		log.Fatal("Failed to determine server IP address")
	}

	operatorPort = ":" + operatorPort

	output := "Server started at http://" + serverIP + operatorPort

	UI.InsertLogMarkup(output)

	server := http.Server{
		Addr: operatorPort,
	}

	// Serve
	err = server.ListenAndServe()
	if err != nil {
		return
	}
	defer server.Close()
}

func createListener(cmdString string, respChan chan<- string) {
	parts := strings.Split(cmdString, ",")
	if len(parts) != 3 {
		return
	}

	listenerProto := parts[0]
	listenerIP := parts[1]
	listenerPort := parts[2]

	listenerAddr := fmt.Sprintf("%s:%s", listenerIP, listenerPort)
	listener, err := net.Listen("tcp", listenerAddr)
	if err != nil {
		resp := fmt.Sprintf("Error: %s\n", err)
		UI.InsertLogMarkup(resp)
		respChan <- resp
		return
	}
	listener.Close()

	switch strings.ToUpper(listenerProto) {
	case "HTTP":
		resp := httpListener(listenerIP, listenerPort)
		respChan <- resp

	case "HTTPS":
		resp := httpsListener(listenerIP, listenerPort)
		respChan <- resp
	default:
		resp := fmt.Sprintf("Invalid protocol: %s", listenerProto)
		respChan <- resp
	}
}

func httpListener(listenerIP string, listenerPort string) string {
	endPoint := "/" + listenerPort
	http.HandleFunc(endPoint, handleImplant)

	server := &http.Server{
		Addr: listenerIP + ":" + listenerPort,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			output := fmt.Sprintf("Error starting HTTP listener: %s\n", err)
			UI.InsertLogMarkup(output)
		}
	}()

	output := "Listener started at http://" + listenerIP + ":" + listenerPort
	UI.InsertLogMarkup(output)

	return output
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
