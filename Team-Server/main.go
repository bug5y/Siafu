package main
import (
    "Team-Server/UI"
    "Team-Server/DB"
    "Team-Server/Server"
)

/*
# Team-Server/UI
UI/ui.go:4:5: undefined: gtk
UI/ui.go:6:17: undefined: gtk
UI/ui.go:8:9: undefined: fmt
UI/ui.go:12:9: undefined: gtk
UI/ui.go:16:5: undefined: mRefProvider
UI/ui.go:16:24: undefined: gtk
UI/ui.go:17:5: undefined: mRefProvider
UI/ui.go:20:5: undefined: screen
UI/ui.go:20:17: undefined: gdk
UI/ui.go:69:24: undefined: gtk
UI/ui.go:20:5: too many errors
# Team-Server/DB
DB/db.go:9:9: undefined: log
DB/db.go:14:18: undefined: sql
DB/db.go:16:16: undefined: os
DB/db.go:18:9: undefined: log
DB/db.go:22:14: undefined: filepath
DB/db.go:23:18: undefined: os
DB/db.go:24:16: undefined: os
DB/db.go:26:13: undefined: log
DB/db.go:31:14: undefined: filepath
DB/db.go:32:16: undefined: sql
DB/db.go:32:16: too many errors


*/

var err error

var ip string
var cmdGroup string 
var cmdString string













func main() {

    UI.InitUI()

    DB.InitDB()

    // Start the server in a goroutine
    go OperatorServer()

    UI.BuildUI()
}


