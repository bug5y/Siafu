
package main

import (
    "Operator/UI"
    "Operator/Client"
    "fmt"
)

func main() {

    fmt.Println("main")
    UI.InitUI()

    Client.ServerConnection()

    go Client.UpdateLog()

    go UI.Connections()

    UI.BuildUI()
}