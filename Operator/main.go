package main

import (
	_Client "Operator/Client"
	_Common "Operator/Common"
	_UI "Operator/UI"
)

func main() {
	_UI.InitUI()
	_Client.ServerConnection()
	_Common.GoRoutine(_UI.StartUIConnections)
	_Common.GoRoutine(_Client.UpdateLog)
	_UI.BuildUI()
}
