package Common

import (
    "time"
    "github.com/gotk3/gotk3/gtk"
)
const (
    LightBlue = "#90afc5" /* selected items & other interactive components */
    BrightBlue = "#1e90ff" /* use for notifications */
    BrightRed = "#e14e19" /* use for error notifications */
    AllWhite = "#ffffff" /* text */
)
var Protos = []string{"http", "https"}

var ServerURL string // One of these needs to go
var ServerIP string // One of these needs to go
var indicator = "<span>[" + "<span foreground=\"" + LightBlue + "\">+</span>" + "]</span>"
var LogBuffer *gtk.TextBuffer
var CurrentBuffer *gtk.TextBuffer
var CurrentID string
var ImplantCmd bool
var ServerCmd bool 

func InsertLogMarkup(Text string) { // Inserts to log
    
    iter := LogBuffer.GetEndIter()
    
    currentTime := time.Now()
    formattedTime := "<span foreground=\"" + LightBlue + "\">" + currentTime.Format("2006-01-02 15:04:05") + "</span>"

    markup := formattedTime + " " + indicator + " " + Text

    LogBuffer.InsertMarkup(iter, markup)
}

func InsertLogText(Text string) { // Inserts to log 
        iter := LogBuffer.GetEndIter()
        LogBuffer.Insert(iter, Text)
}

func HandleAlerts(output string, buffer *gtk.TextBuffer, warn bool){ // Inserts to CMD console
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
