package _Client

import (
	_Common "Operator/Common"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	operatorID = "123abc" // Will later be used with authentication
)

var port string
var defaultPort = "8443"

var RouteMutex sync.Mutex

func ServerConnection() {
	_Common.Shared.ServerURL = SetServer()
	verifyAndUpdateServerStatus()
}

func SetServer() string {
	if _Common.Shared.ServerURL == "" {
		defaultIP, err := _Common.GetIP()
		if err != nil {
			Text := "Failed to set an address for the server"
			_Common.InsertLogMarkup(Text)
		}
		port = defaultPort
		url := "http://" + defaultIP + ":" + port
		return url
	} else {
		return _Common.Shared.ServerURL
	}
}

func verifyAndUpdateServerStatus() {
	if verifyServerUp(_Common.Shared.ServerURL) {
		Text := "<span>Connected to server: <span foreground=\"" + _Common.AllWhite + "\">" + _Common.Shared.ServerURL + "</span></span>"
		_Common.InsertLogMarkup(Text)
		_Common.SetServerStatus(true)
	} else {
		Text := "<span foreground=\"" + _Common.BrightRed + "\">Unable to connect to server: <span foreground=\"" + _Common.AllWhite + "\">" + _Common.Shared.ServerURL + "</span></span>"
		_Common.InsertLogMarkup(Text)
		_Common.SetServerStatus(false)
	}
	_Common.GoRoutine(monitorServerStatus)
}

func monitorServerStatus() {
	for {
		if !_Common.GetServerStatus() {
			if verifyServerUp(_Common.Shared.ServerURL) {
				Text := "<span>Connected to server: <span foreground=\"" + _Common.AllWhite + "\">" + _Common.Shared.ServerURL + "</span></span>"
				_Common.InsertLogMarkup(Text)
				_Common.SetServerStatus(true)
			}
		}
		if _Common.GetServerStatus() {
			if !verifyServerUp(_Common.Shared.ServerURL) {
				Text := "<span foreground=\"" + _Common.BrightRed + "\">Unable to connect to server: <span foreground=\"" + _Common.AllWhite + "\">" + _Common.Shared.ServerURL + "</span></span>"
				_Common.InsertLogMarkup(Text)
				_Common.SetServerStatus(false)
			}
		}
		time.Sleep(1 * time.Second)
	}
}

func RouteCMD(ctx context.Context, ns *_Common.NonSharedStruct, resultChan chan<- *_Common.NonSharedStruct, errChan chan<- error) { // Returns channels
	if !_Common.GetServerStatus() {
		//errChan <- fmt.Errorf("server is not running")
		return
	}

	ct := _Common.GetCurrentTab()
	id := ct.CurrentID
	// Reset
	ns.Sort = &_Common.Sorting{
		ImplantCmd: false,
		ServerCMD:  false,
	}

	switch ns.CS.Group {
	case "shell":
		ns.Sort = &_Common.Sorting{
			ImplantCmd: true,
			ServerCMD:  false,
		}
		result, err := Request(ns, id, resultChan, errChan)
		if err != nil {
			fmt.Println("err creating cmd", err)
			output := fmt.Sprintf("Error sending command: %v\n", err)
			warn := true
			_Common.ConsoleAlerts(output, warn)
			errChan <- err
			return
		}
		resultChan <- result

	case "listener":
		ns.Sort = &_Common.Sorting{
			ImplantCmd: false,
			ServerCMD:  true,
		}
		result, err := Request(ns, id, resultChan, errChan)
		if err != nil {
			if ct.CurrentBuffer != nil && ct.CurrentBuffer.GetCharCount() > 0 {
				output := fmt.Sprintf("Error sending command: %v\n", err)
				warn := true
				_Common.ConsoleAlerts(output, warn)
			}
			errChan <- err
			return
		}
		resultChan <- result

	case "log":
		ns.Sort = &_Common.Sorting{
			ImplantCmd: false,
			ServerCMD:  true,
		}
		result, err := Request(ns, id, resultChan, errChan)
		if err != nil {
			output := fmt.Sprintf("Error sending command: %v\n", err)
			warn := true
			_Common.ConsoleAlerts(output, warn)
			errChan <- err
			return
		}
		resultChan <- result

	default:
		output := "invalid command group\n"
		warn := true
		_Common.ConsoleAlerts(output, warn)
		errChan <- fmt.Errorf(output)
	}
}

func Request(ns *_Common.NonSharedStruct, id string, resultChan chan<- *_Common.NonSharedStruct, errChan chan<- error) (*_Common.NonSharedStruct, error) {
	// Build
	jsonData, err := json.Marshal(ns.CS)
	if err != nil {
		return nil, err
	}
	ns.Base64Json = base64.StdEncoding.EncodeToString(jsonData)

	// Send
	ns.CS.HttpResp, err = SendRequest(ns, id)
	if err != nil {
		output := "error sending request"
		warn := true
		_Common.ConsoleAlerts(output, warn)
		return nil, err
	}

	// Process
	ns.CS.Response, err = ProcessResponse(ns)
	if err != nil {
		return nil, err
	}
	return ns, nil
}

func SendRequest(ns *_Common.NonSharedStruct, id string) (*http.Response, error) {
	client := &http.Client{}

	if !_Common.GetServerStatus() {
		return nil, nil
	}

	if ns.Sort.ImplantCmd {
		endpoint := "/operator"
		server := _Common.Shared.ServerURL + endpoint
		req, err := http.NewRequest("POST", server, bytes.NewBuffer([]byte(ns.Base64Json)))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "text/plain")
		req.Header.Set("Operator-ID", operatorID)
		req.Header.Set("Implant-ID", id)
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		return resp, nil
	}

	if ns.Sort.ServerCMD {
		endpoint := "/info"
		server := _Common.Shared.ServerURL + endpoint
		req, err := http.NewRequest("GET", server, bytes.NewBuffer([]byte(ns.Base64Json)))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "text/plain")
		req.Header.Set("Operator-ID", operatorID)
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		return resp, nil
	} else {
		return nil, nil
	}

}

func ProcessResponse(ns *_Common.NonSharedStruct) (string, error) {
	var decodedData []byte

	if ns.CS.HttpResp == nil {
		return "", fmt.Errorf("response is nil")
	}
	responseBytes, err := io.ReadAll(ns.CS.HttpResp.Body)
	if err != nil {
		return "", err
	}
	defer ns.CS.HttpResp.Body.Close()

	responseBody := string(responseBytes)

	dataIndex := strings.Index(responseBody, "Data:")
	if dataIndex != -1 {
		data := strings.TrimSpace(responseBody[dataIndex+len("Data:"):])
		decodedData, err = base64.StdEncoding.DecodeString(data)
		if err != nil {
			return "", err
		}
	} else {
		output := "No data found in the response\n"
		return output, err
	}
	response := bytes.TrimSpace(decodedData)

	if err = json.Unmarshal(response, &ns.CS); err != nil {
		return "JSON error", err
	}

	return ns.CS.Response, nil
}
