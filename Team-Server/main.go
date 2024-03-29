package main
import (
    "Team-Server/UI"
    "Team-Server/DB"
    "Team-Server/Server"
)

var err error
var ip string
var cmdGroup string 
var cmdString string

func main() {

    UI.InitUI()

    DB.InitDB()

    // Start the server in a goroutine
    go Server.OperatorServer()

    UI.BuildUI()
}


