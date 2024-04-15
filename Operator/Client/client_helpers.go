package _Client

import (
	_Common "Operator/Common"
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func PassCMD(cmd string) {
	resultChan := make(chan *_Common.NonSharedStruct)
	errChan := make(chan error)

	_Common.GoRoutine(
		func() {
			defer close(resultChan)
			defer close(errChan)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			cs := &_Common.CommandStruct{
				Group:  strings.Split(cmd, " ")[0],
				String: strings.TrimSpace(strings.Join(strings.Split(cmd, " ")[1:], " ")),
			}

			ns := &_Common.NonSharedStruct{CS: cs}
			RouteCMD(ctx, ns, resultChan, errChan)
		})
	select {
	case result := <-resultChan:
		// Print
		PrintErr := _Common.PrintResults(result, nil)
		if PrintErr != nil {
			return
		}
	case err := <-errChan:
		// Handle the error
		fmt.Println("Error:", err)
	}
}

func verifyServerUp(serverURL string) bool {
	resp, err := http.Get(serverURL)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return true
}

func UpdateLog() {
	for {
		result, _ := updateLogLoop()
		if result != nil {
			PrintErr := _Common.PrintResults(result, nil)
			if PrintErr != nil {
				return
			}
		}
		time.Sleep(2 * time.Second)
	}
}

func updateLogLoop() (*_Common.NonSharedStruct, error) {
	resultChan := make(chan *_Common.NonSharedStruct)
	errChan := make(chan error)
	cmd := "log"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cs := &_Common.CommandStruct{
		Group:  strings.Split(cmd, " ")[0],
		String: strings.TrimSpace(strings.Join(strings.Split(cmd, " ")[1:], " ")),
	}

	ns := &_Common.NonSharedStruct{CS: cs}

	go func() {
		RouteCMD(ctx, ns, resultChan, errChan)
	}()

	select {
	case result := <-resultChan:
		defer cancel()
		return result, nil
	case err := <-errChan:
		fmt.Println("Error:", err)
		defer cancel()
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}

}

func NewListener(cmd string) {
	PassCMD(cmd)
}
