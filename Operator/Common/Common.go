package Common

import (
    "time"
    "github.com/gotk3/gotk3/gtk"
    "sync"
    "fmt"
    //"github.com/gotk3/gotk3/glib"
)



const (
    LightBlue = "#90afc5" /* selected items & other interactive components */
    BrightBlue = "#1e90ff" /* use for notifications */
    BrightRed = "#e14e19" /* use for error notifications */
    AllWhite = "#ffffff" /* text */
)

var Protos = []string{"http", "https"}
var Store *gtk.ListStore
var ServerURL string // One of these needs to go
var ServerIP string // One of these needs to go
var indicator = "<span>[" + "<span foreground=\"" + LightBlue + "\">+</span>" + "]</span>"

type Tabs struct {
    PageIndex int
    ID string
    TabLabel   *gtk.Label       // Label for the tab
    Buffer *gtk.TextBuffer           // Name of the buffer associated with the tab
    Entry      *gtk.Entry       // Entry widget associated with the tab
    Button *gtk.Button
}

var TabMap = make(map[int]*Tabs)
var LabelMap map[*gtk.Label]int

type CurrentTab struct {
    CurrentBuffer *gtk.TextBuffer
    CurrentEntry *gtk.Entry 
    CurrentPage int
    CmdPlaceHolder string
    CurrentMutex sync.Mutex
    CurrentID string
}
var CT *CurrentTab

type ActivityLog struct { 
    LogBuffer *gtk.TextBuffer
    LogBufferMu sync.Mutex
}
var AL *ActivityLog

var CmdPlaceHolder = "Siafu>>"

type Sorting struct { // (ss *Sorting)
    ImplantCmd bool
    ServerCmd bool 
    SortingMutex sync.Mutex
}

var (
    mu sync.Mutex
    
)

func init() {
    AL = &ActivityLog{
        LogBuffer:   nil, // or initialize the buffer here
        LogBufferMu: sync.Mutex{},
    }
}

// Update the CurrentTab struct
func (ct *CurrentTab) Update(buffer *gtk.TextBuffer, entry *gtk.Entry, page int, placeholder string) {
    ct.CurrentMutex.Lock()
    defer ct.CurrentMutex.Unlock()

    ct.CurrentBuffer = buffer
    ct.CurrentEntry = entry
    ct.CurrentPage = page
    ct.CmdPlaceHolder = placeholder
}

func InsertLogMarkup(Text string) { // Inserts to activity log
    fmt.Println("insertlogmarkup")
    AL.LogBufferMu.Lock()
    defer AL.LogBufferMu.Unlock()

    iter := AL.LogBuffer.GetEndIter()
    currentTime := time.Now()
    formattedTime := "<span foreground=\"" + LightBlue + "\">" + "  " + currentTime.Format("2006-01-02 15:04:05") + "</span>"
    markup := formattedTime + " " + indicator + " " + Text
    AL.LogBuffer.InsertMarkup(iter, markup)
}

func ConsoleAlerts(output string, buffer *gtk.TextBuffer, warn bool){ // Inserts to CMD console
    fmt.Println("consolealert")
    var alertcolor string

    if warn {
        alertcolor = BrightRed    
    } else {
        alertcolor = BrightBlue 
    }

    alertindicator := "<span foreground=\"" + alertcolor + "\">[</span>" + "<span foreground=\"" + LightBlue + "\">+</span>" + "<span foreground=\"" + alertcolor + "\">]</span> "
    output = alertindicator + output

    iter := buffer.GetEndIter()
    buffer.InsertMarkup(iter, output)
}

func CreateRow(rowData []string) {
    fmt.Println("createrow")
    iter := Store.Append()

    for i, value := range rowData {
        Store.SetValue(iter, i, value)
    }
}


func UpdateRow(key string, lastseen string) {
    fmt.Println("updaterow")

    // Find the row index where the key is in the third column
    var rowIndex int

    iter, _ := Store.ToTreeModel().GetIterFirst()
    for {
        value, _ := Store.GetValue(iter, 2)
        keyValue, _ := value.GetString()
        if keyValue == key {
            treePath, _ := Store.ToTreeModel().GetPath(iter)
            rowIndex = treePath.GetIndices()[0]
            break
        }
        if !Store.ToTreeModel().IterNext(iter) {
            break
        }
    }

    fmt.Println("row index", rowIndex)


}


var CommandChan = make(chan string)
func CommandOutput(output string, buffer *gtk.TextBuffer){ // Inserts to CMD console
    CT.CurrentMutex.Lock()
    defer CT.CurrentMutex.Unlock()

    fmt.Println("commandoutput")
    iter := buffer.GetEndIter()
    buffer.Insert(iter, output)
}
