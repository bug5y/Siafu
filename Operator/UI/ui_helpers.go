package UI

import (
	"fmt"
    "github.com/gotk3/gotk3/gtk"
    //"Operator/Common"
)

func initPlaceHolder(entry *gtk.Entry) {
    fmt.Println("initplaceholder")
    entry.SetPlaceholderText(PlaceHolder)
}

func showErrorDialog(parent *gtk.Dialog, message string) {
    fmt.Println("showerror")
    dialog := gtk.MessageDialogNew(parent, gtk.DIALOG_MODAL, gtk.MESSAGE_ERROR, gtk.BUTTONS_OK, message)
    dialog.SetTitle("Error")
    dialog.Run()
    dialog.Destroy()
}

func createEntry() *gtk.Entry {
    fmt.Println("createentry")
    entry, _ := gtk.EntryNew()
    entryCount = entryCount + 1
    return entry
}





