package UI

import (
	"fmt"
    "github.com/gotk3/gotk3/gtk"
)



func initPlaceHolder(entry *gtk.Entry, cmdPlaceHolder string) {
    entry.SetPlaceholderText(cmdPlaceHolder)
}

func showErrorDialog(parent *gtk.Dialog, message string) {
    dialog := gtk.MessageDialogNew(parent, gtk.DIALOG_MODAL, gtk.MESSAGE_ERROR, gtk.BUTTONS_OK, message)
    dialog.SetTitle("Error")
    dialog.Run()
    dialog.Destroy()
}

func createEntry() *gtk.Entry {
    entry, _ := gtk.EntryNew()
    entryCount = entryCount + 1
    return entry
}

func printTabInfoMap() { // For debugging
    fmt.Println("Printing tabInfoMap contents:")
    for key, value := range tabInfoMap {
        fmt.Printf("Key: %v, Value: {ID: %v, PageIndex: %v, Button: %v}\n", key, value.ID, value.PageIndex, value.Button)
    }
}

func handleOutput(output string, buffer *gtk.TextBuffer){ // Inserts to CMD console
    // Insert the output into the text buffer
    iter := buffer.GetEndIter()
    buffer.Insert(iter, output)
}


