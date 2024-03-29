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
var Protos = []string{"HTTP", "HTTPS"}

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


/*
UI/ui.go:15:2: "Operator/Client" imported and not used
UI/ui.go:776:9: undefined: routeCMD -
UI/ui.go:782:8: undefined: implant_cmd
UI/ui.go:796:8: undefined: servercommand
UI/ui.go:802:53: undefined: serverIP
UI/ui.go:803:9: undefined: insertLogText
UI/ui.go:806:9: undefined: insertLogMarkup
UI/ui.go:813:12: undefined: exec
UI/ui_helpers.go:6:2: "log" imported and not used
UI/ui_helpers.go:7:5: "github.com/fstanis/screenresolution" imported and not used
UI/ui_helpers.go:7:5: too many errors

*/
