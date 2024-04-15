package _UI

import (
	"github.com/gotk3/gotk3/gtk"
)

func initPlaceHolder(entry *gtk.Entry) {
	entry.SetPlaceholderText(PlaceHolder)
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
