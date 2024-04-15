package _Common

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gotk3/gotk3/gtk"
)

const (
	LightBlue  = "#90afc5" /* selected items & other interactive components */
	BrightBlue = "#1e90ff" /* use for notifications */
	BrightRed  = "#e14e19" /* use for error notifications */
	AllWhite   = "#ffffff" /* text */
)

// Shared variables
var Shared = struct {
	Tabs      map[int]*Tabs
	LM        map[*gtk.Label]int
	CT        *CurrentTab
	AL        *ActivityLog
	LS        *ListenerStruct
	SS        *Status
	CL        Connections //map
	Protos    []string
	Store     *gtk.ListStore
	ServerURL string
	ServerIP  string
	Indicator string
}{}

type Tabs struct {
	PageIndex int
	ID        string
	TabLabel  *gtk.Label
	Buffer    *gtk.TextBuffer
	Entry     *gtk.Entry
	Button    *gtk.Button
}

var TabMap = make(map[int]*Tabs)
var LabelMap map[*gtk.Label]int

type CurrentTab struct {
	CurrentBuffer  *gtk.TextBuffer
	CurrentEntry   *gtk.Entry
	CurrentPage    int
	CmdPlaceHolder string
	CtMutex        sync.Mutex
	CurrentID      string
}

type ActivityLog struct {
	LogBuffer *gtk.TextBuffer
	LbMutex   sync.Mutex
}

type ListenerStruct struct {
	Ip    string
	Port  string
	Proto string
}

type Status struct {
	ServerStatus bool
	StatusMutex  sync.Mutex
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

type NonSharedStruct struct {
	Result     string
	Error      error
	Base64Json string
	Sort       *Sorting
	CS         *CommandStruct
}

type CommandStruct struct {
	Group    string `json:"Group"`
	String   string `json:"String"`
	Response string `json:"Response"`

	HttpResp *http.Response // Not marshalled
	CmdMutex sync.Mutex     // Not marshalled
}

type Sorting struct {
	ImplantCmd bool
	ServerCMD  bool
	SortMutex  sync.Mutex
}

func init() {
	Shared.CT = &CurrentTab{
		CtMutex: sync.Mutex{},
	}
	Shared.AL = &ActivityLog{
		LbMutex: sync.Mutex{},
	}
	Shared.LS = &ListenerStruct{}
	Shared.SS = &Status{
		StatusMutex: sync.Mutex{},
	}
	Shared.Tabs = make(map[int]*Tabs)
	Shared.LM = make(map[*gtk.Label]int)
	Shared.CL = make(Connections)
	Shared.Protos = []string{"http", "https"}
	Shared.Indicator = "<span>[" + "<span foreground=\"" + LightBlue + "\">+</span>" + "]</span>"

}

func GetIP() (string, error) {
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
			Shared.ServerURL = ipnet.IP.String()
			return ipnet.IP.String(), nil
		}
	}
	return "", fmt.Errorf("unable to determine IP address")
}

func GetServerStatus() bool {
	Shared.SS.StatusMutex.Lock()
	defer Shared.SS.StatusMutex.Unlock()
	return Shared.SS.ServerStatus
}

func SetServerStatus(status bool) {
	Shared.SS.StatusMutex.Lock()
	defer Shared.SS.StatusMutex.Unlock()
	Shared.SS.ServerStatus = status
}

func GetCurrentTab() *CurrentTab {
	Shared.CT.CtMutex.Lock()
	defer Shared.CT.CtMutex.Unlock()
	return Shared.CT
}

func SetCurrentTab(ct *CurrentTab) {
	Shared.CT.CtMutex.Lock()
	defer Shared.CT.CtMutex.Unlock()
	Shared.CT = ct
}

func GetActivityLog() *ActivityLog {
	Shared.AL.LbMutex.Lock()
	defer Shared.AL.LbMutex.Unlock()
	return Shared.AL
}

func SetActivityLog(al *ActivityLog) {
	Shared.AL.LbMutex.Lock()
	defer Shared.AL.LbMutex.Unlock()
	Shared.AL = al
}

func CommandOutput(response string) { // Inserts to CMD console
	output := response + "\n"

	ct := GetCurrentTab()
	Shared.CT.CtMutex.Lock()
	if ct.CurrentBuffer == nil {
		return
	}
	abuffer := ct.CurrentBuffer
	iter := abuffer.GetEndIter()
	abuffer.Insert(iter, output)
	Shared.CT.CtMutex.Unlock()
}

func InsertLogMarkup(Text string) { // Inserts to activity log
	al := GetActivityLog()

	if al.LogBuffer == nil {
		fmt.Println("Nil log")
		return
	}
	for i := 0; i < 3; { // Attempt up to 3 times
		Shared.AL.LbMutex.Lock()

		iter := al.LogBuffer.GetEndIter()

		// If logbuffer is not intialized then init

		currentTime := time.Now()
		formattedTime := "<span foreground=\"" + LightBlue + "\">" + "  " + currentTime.Format("2006-01-02 15:04:05") + "</span>"
		markup := formattedTime + " " + Shared.Indicator + " " + Text + "\n"

		func() {
			defer func() {
				if r := recover(); r != nil {
					i++
					fmt.Printf("Error inserting log markup: %v\n", r)
				}
				Shared.AL.LbMutex.Unlock()
			}()
			al.LogBuffer.InsertMarkup(iter, markup)
		}()
		return
	}

	/*
			if strings.Contains(err.Error(), "Invalid text buffer iterator") {
				Shared.AL.LbMutex.Unlock()
				time.Sleep(100 * time.Millisecond) // Wait a short time before retrying
				continue
			}

			// If it's a different error, log it and unlock
			fmt.Printf("Error inserting log markup: %v\n", err)
			Shared.AL.LbMutex.Unlock()
			return
		}
	*/
}

func ConsoleAlerts(output string, warn bool) { // Inserts to CMD console
	fmt.Println("consolealert")
	ct := GetCurrentTab()
	Shared.CT.CtMutex.Lock()
	if ct.CurrentBuffer == nil {
		return
	}
	abuffer := ct.CurrentBuffer

	var alertcolor string
	if warn {
		alertcolor = BrightRed
	} else {
		alertcolor = BrightBlue
	}
	alertindicator := "<span foreground=\"" + alertcolor + "\">[</span>" + "<span foreground=\"" + LightBlue + "\">+</span>" + "<span foreground=\"" + alertcolor + "\">]</span> "
	output = alertindicator + output
	iter := abuffer.GetEndIter()
	abuffer.InsertMarkup(iter, output)
	Shared.CT.CtMutex.Unlock()
}

func CreateRow(rowData []string) {
	fmt.Println("createrow")
	iter := Shared.Store.Append()

	for i, value := range rowData {
		Shared.Store.SetValue(iter, i, value)
	}
}

func UpdateRow(key string, lastseen string) {
	iter, _ := Shared.Store.ToTreeModel().GetIterFirst()
	for {
		value, _ := Shared.Store.GetValue(iter, 5) //8th column

		_, err := value.GetString()
		if err != nil {
			return
		}

		Shared.Store.SetValue(iter, 5, lastseen)

		if !Shared.Store.ToTreeModel().IterNext(iter) {
			break
		}
	}
}

func GoRoutine(f interface{}, args ...interface{}) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Panic occurred in the goroutine: %v\n", r)
			}
		}()

		switch f := f.(type) {
		case func():
			f()
		case func(interface{}):
			f(args[0])
		case func(string):
			f(args[0].(string))
		case func(string, string, string):
			f(args[0].(string), args[1].(string), args[2].(string))
		default:
			panic(fmt.Sprintf("RunInBackground: unsupported function type %T", f))
		}
	}()
}

func PrintResults(ns *NonSharedStruct, err error) error {
	Text := ns.CS.Response
	if Text == "" || Text == "{}" {
		return nil
	}

	switch ns.CS.Group {
	case "log":
		updateConnectionLog(Text)
		return nil
	case "listener":
		InsertLogMarkup(Text)
		return nil
	default:
		if err != nil {
			warn := true
			output := "error processing response"
			ConsoleAlerts(output, warn)
			return nil
		} else {
			CommandOutput(Text)
			return nil
		}
	}
}

func updateConnectionLog(Text string) {
	var updatedLog map[string]ConnectionDetails

	err := json.Unmarshal([]byte(Text), &updatedLog)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	for key, connectionDetails := range updatedLog {
		if existingConnection, exists := ConnectionLog[key]; exists {
			existingConnection.LastSeen = connectionDetails.LastSeen
			ConnectionLog[key] = existingConnection
			UpdateRow(key, connectionDetails.LastSeen)
		} else {
			text := "New connection" + " " + key
			InsertLogMarkup(text)
			ConnectionLog[key] = connectionDetails
			data := []string{connectionDetails.HostVersion, connectionDetails.AgentType, connectionDetails.ImplantID, connectionDetails.HostName, connectionDetails.User, connectionDetails.LastSeen, connectionDetails.InternalIP, connectionDetails.ExternalIP}
			CreateRow(data)
		}
	}
	ConnectionLog = updatedLog
}
