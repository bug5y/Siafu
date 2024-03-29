
package main

import (
    "Operator/UI"
    "Operator/Client"
)

func main() {

    UI.InitUI()

    Client.InitConnection()

    go Client.UpdateLog()

    go UI.Connections()

    UI.BuildUI()
}