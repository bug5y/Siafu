package _UI

import (
	_Client "Operator/Client"
	_Common "Operator/Common"
	"fmt"
	"log"
	"os/exec"
	"sync"

	"github.com/fstanis/screenresolution"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/pango"
)

var nilButton *gtk.Button = nil
var entryCount = 0
var ScreenWidth int
var ScreenHeight int
var mRefProvider *gtk.CssProvider
var entryDialog *gtk.Dialog
var win *gtk.Window
var screen *gdk.Screen
var infoTable *gtk.TreeView
var notebook *gtk.Notebook
var paned *gtk.Paned
var vbox *gtk.Box

var ID string
var PlaceHolder = "Siafu>>"

func InitUI() {
	gtk.Init(nil)

	settings, _ := gtk.SettingsGetDefault()
	settings.SetProperty("gtk-theme-name", "Adwaita-dark")

	win, _ = gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	win.SetTitle("Siafu Operator")
	win.Connect("destroy", func() {
		gtk.MainQuit()
	})

	// Load CSS stylesheet
	//cssPath := "." + SiafuBase + "/UI/Material-DeepOcean/gtk-dark.css"
	mRefProvider, _ = gtk.CssProviderNew()
	mRefProvider.LoadFromPath("./UI/Material-DeepOcean/gtk-dark.css")

	// Apply to whole app
	screen, _ = gdk.ScreenGetDefault()
	gtk.AddProviderForScreen(screen, mRefProvider, gtk.STYLE_PROVIDER_PRIORITY_USER)

	ScreenWidth = 0
	ScreenHeight = 0
	resolution := screenresolution.GetPrimary()
	if resolution.Width == 0 && resolution.Height == 0 {
		ScreenWidth = 600
		ScreenHeight = 400
	} else {
		ScreenWidth = int(float64(resolution.Width) * 0.5)
		ScreenHeight = int(float64(resolution.Height) * 0.5)
	}

	menubar, err := gtk.MenuBarNew()
	if err != nil {
		fmt.Println("Unable to create menubar:", err)
	}
	vbox, _ = gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 10)
	entry, _ := gtk.EntryNew()
	notebook, _ = gtk.NotebookNew()
	paned, _ = gtk.PanedNew(gtk.ORIENTATION_VERTICAL)

	_Common.LabelMap = make(map[*gtk.Label]int)

	fileMenu := createFileMenu()
	buildMenu := createBuildMenu()

	menubar.Append(fileMenu)

	menubar.Append(buildMenu)

	win.Add(vbox)

	vbox.PackStart(menubar, false, false, 0)

	vbox.PackStart(paned, true, true, 0)

	// Pane for top half
	topPaned, _ := gtk.PanedNew(gtk.ORIENTATION_HORIZONTAL)
	paned.Pack1(topPaned, true, true)

	// Table
	scrolledTable, _ := gtk.ScrolledWindowNew(nil, nil)
	scrolledTable.SetPolicy(gtk.POLICY_AUTOMATIC, gtk.POLICY_AUTOMATIC)
	infoTable, _ = gtk.TreeViewNew()
	scrolledTable.Add(infoTable)
	topPaned.Pack1(scrolledTable, true, true)
	TableWidth := int(float64(ScreenWidth) * 0.55)
	topPaned.SetPosition(TableWidth)

	// Create a list store
	_Common.Shared.Store, _ = gtk.ListStoreNew(glib.TYPE_STRING, glib.TYPE_STRING, glib.TYPE_STRING, glib.TYPE_STRING, glib.TYPE_STRING, glib.TYPE_STRING, glib.TYPE_STRING, glib.TYPE_STRING)
	createTable(infoTable, _Common.Shared.Store, TableWidth)

	scrolledText, _ := gtk.ScrolledWindowNew(nil, nil)
	topPaned.Pack2(scrolledText, true, true)
	servertxt, _ := gtk.TextViewNew()
	servertxt.SetEditable(false)
	scrolledText.Add(servertxt)

	serverMenu := createServerMenu()
	menubar.Append(serverMenu)

	buffer, err := servertxt.GetBuffer()
	if err != nil {
		fmt.Println("Error getting log buffer")
	}

	al := &_Common.ActivityLog{
		LogBuffer: buffer,
		LbMutex:   sync.Mutex{},
	}
	_Common.SetActivityLog(al)
	initPlaceHolder(entry)
}

func BuildUI() {
	win.SetDefaultSize(ScreenWidth, ScreenHeight)
	win.ShowAll()
	gtk.Main()
}

func StartUIConnections() {
	_Common.GoRoutine(Connections)
}

func Connections() {

	infoTable.Connect("button-press-event", func(tv *gtk.TreeView, ev *gdk.Event) {
		event := &gdk.EventButton{
			Event: ev,
		}
		if event.Button() == 3 { // Right click
			path, _, _, _, err := tv.GetPathAtPos(int(event.X()), int(event.Y()))
			if !err {
				return
			}
			if path != nil {
				selection, err := tv.GetSelection()
				if err != nil {
					log.Fatal("Unable to get selection:", err)
				}
				selection.SelectPath(path)
				var parent *gtk.Window

				showContextMenu(notebook, vbox, win, PlaceHolder, parent, tv, _Common.Shared.Store, paned)

			}
		}
	})

	notebook.Connect("switch-page", func(notebook *gtk.Notebook, page *gtk.Widget, pageNum int) {
		tab := _Common.TabMap[pageNum]

		if tab != nil {
			_Common.SetCurrentTab(&_Common.CurrentTab{
				CurrentID:     tab.ID,
				CurrentBuffer: tab.Buffer,
				CurrentEntry:  tab.Entry,
				CurrentPage:   tab.PageIndex,
				CtMutex:       sync.Mutex{},
			})
		}
		curpg := notebook.GetCurrentPage()
		for in, tab := range _Common.TabMap {
			if in == curpg {
				tab.Entry.SetVisible(true)
			} else {
				tab.Entry.SetVisible(false)
			}
		}
	})

}

func createTable(infoTable *gtk.TreeView, store *gtk.ListStore, TableWidth int) {
	columns := []string{"OS", "Agent Type", "UID", "Host Name", "User", "Last Seen", "Internal IP(s)", "External IP(s)"}

	infoTable.SetModel(store)
	columnwidth := TableWidth / len(columns)
	for i, columnName := range columns {
		contentCellRenderer, err := gtk.CellRendererTextNew()
		if err != nil {
			log.Fatal("Unable to create cell renderer:", err)
		}
		contentCellRenderer.SetAlignment(0.5, 0.5) // Center the content

		column, err := gtk.TreeViewColumnNewWithAttribute(columnName, contentCellRenderer, "text", i)
		if err != nil {
			log.Fatal("Unable to create tree view column:", err)
		}

		column.SetMinWidth(columnwidth)
		column.SetResizable(true)
		column.SetExpand(false)
		infoTable.AppendColumn(column)
	}
}

func showContextMenu(notebook *gtk.Notebook, vbox *gtk.Box, win *gtk.Window, PlaceHolder string, parent *gtk.Window, tv *gtk.TreeView, store *gtk.ListStore, paned *gtk.Paned) {
	menu, _ := gtk.MenuNew()

	// "Open Console"
	menuItemOpenConsole, _ := gtk.MenuItemNewWithLabel("Open Console")
	menuItemOpenConsole.Connect("activate", func() {
		newConsole(notebook, vbox, win, PlaceHolder, tv, paned)
	})
	menu.Append(menuItemOpenConsole)

	// "Kill Implant"
	menuItemKillImplant, _ := gtk.MenuItemNewWithLabel("Kill Implant")
	menuItemKillImplant.Connect("activate", func() {
		dialog := gtk.MessageDialogNew(parent, gtk.DIALOG_MODAL, gtk.MESSAGE_QUESTION, gtk.BUTTONS_YES_NO, "Are you sure?\n This will remove the implant from its host")
		response := dialog.Run()
		if response == gtk.RESPONSE_YES {
			removeImplant(tv, store, notebook, paned, vbox)
		}
		dialog.Destroy()
	})
	menu.Append(menuItemKillImplant)
	menu.ShowAll()
	menu.PopupAtPointer(nil)
}

func setupTab(notebook *gtk.Notebook, vbox *gtk.Box, win *gtk.Window, PlaceHolder string, tabName string, paned *gtk.Paned) {
	i := 0
	cmdlog, _ := gtk.ScrolledWindowNew(nil, nil)
	cmdtxt, _ := gtk.TextViewNew()
	cmdtxt.SetEditable(false)
	cmdlog.Add(cmdtxt)

	// Create a vertical box to contain cmdlog and cmdtxt
	cmdBox, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
	cmdBox.PackStart(cmdlog, true, true, 0)

	var tabLabel *gtk.Label

	if tabName == "" {
		tabLabel, _ = gtk.LabelNew(fmt.Sprintf("Tab %d", notebook.GetNPages()+1))
	} else {
		tabLabel, _ = gtk.LabelNew(tabName)
	}

	tabWidget, button := createTabWidget(tabLabel, notebook, paned, vbox, tabName)

	notebook.AppendPage(cmdBox, tabWidget)
	notebook.SetScrollable(true)

	pageIndex := notebook.GetNPages() - 1

	cmdentry := createEntry()

	initPlaceHolder(cmdentry)

	i++

	cmdentry.SetName("entrytab")

	// Create a unique buffer for each tab
	cmdbuffer, _ := cmdtxt.GetBuffer()
	//bufferIndex := 0
	//bufferIndex = bufferIndex + 1
	vbox.PackEnd(cmdentry, false, false, 0)

	_Common.TabMap[pageIndex] = &_Common.Tabs{
		ID:        ID,
		PageIndex: pageIndex,
		TabLabel:  tabLabel,
		Buffer:    cmdbuffer,
		Entry:     cmdentry,
		Button:    button,
	}

	cmdentry.Connect("activate", func() {
		cmd, _ := cmdentry.GetText()
		ct := _Common.GetCurrentTab()
		buffer := ct.CurrentBuffer
		iter := buffer.GetEndIter()
		buffer.Insert(iter, fmt.Sprintf("%s %s\n", PlaceHolder, cmd))

		go func() {
			_Client.PassCMD(cmd)
		}()
		cmdentry.SetText("")
	})

	notebook.ShowAll()
	newPage := notebook.GetCurrentPage()
	for in, tab := range _Common.TabMap {
		if in == newPage {
			tab.Entry.SetVisible(true)
		} else {
			tab.Entry.SetVisible(false)
		}
	}
	if notebook.GetNPages() == 1 {
		paned.Pack2(notebook, true, true)
		win.ShowAll()
	}

}

func removeImplant(tv *gtk.TreeView, store *gtk.ListStore, notebook *gtk.Notebook, paned *gtk.Paned, vbox *gtk.Box) {
	selection, _ := tv.GetSelection()
	_, iter, _ := selection.GetSelected()

	col2Value, _ := store.GetValue(iter, 2) // ID
	col2, _ := col2Value.GetString()
	removeID := col2

	if iter != nil {
		store.Remove(iter)
		removeTab(notebook, paned, vbox, nilButton, removeID)
	}
}

func newConsole(notebook *gtk.Notebook, vbox *gtk.Box, win *gtk.Window, PlaceHolder string, tv *gtk.TreeView, paned *gtk.Paned) {
	selection, err := tv.GetSelection()
	if err != nil {
		fmt.Println("Error getting selection:", err)
		return
	}
	_, iter, status := selection.GetSelected()
	if !status {
		fmt.Println("Error getting selected item")
		return
	}

	model, err := tv.GetModel()
	if err != nil {
		fmt.Println("Error getting model:", err)
		return
	}

	listStore, ok := model.(*gtk.ListStore)
	if !ok {
		fmt.Println("Error casting model to *gtk.ListStore")
		return
	}

	if iter != nil {
		// Retrieve data from each column and convert to string
		col2Value, err := listStore.GetValue(iter, 2) // ID
		if err != nil {
			fmt.Println("Error getting value from column 2:", err)
			return
		}
		col2, _ := col2Value.GetString()

		col4Value, err := listStore.GetValue(iter, 4) // User
		if err != nil {
			fmt.Println("Error getting value from column 4:", err)
			return
		}
		col4, _ := col4Value.GetString()

		col3Value, err := listStore.GetValue(iter, 3) // Host
		if err != nil {
			fmt.Println("Error getting value from column 3:", err)
			return
		}
		col3, _ := col3Value.GetString()

		tabName := col2 + ": " + col4 + "@" + col3

		ID = col2

		setupTab(notebook, vbox, win, PlaceHolder, tabName, paned)

	} else {
		fmt.Println("No item selected")
	}
}

func createTabWidget(label *gtk.Label, notebook *gtk.Notebook, paned *gtk.Paned, vbox *gtk.Box, tabName string) (*gtk.Box, *gtk.Button) {
	tabBox, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)

	closeButton, _ := gtk.ButtonNew()
	closeButton.SetLabel("x")
	closeButton.SetRelief(gtk.RELIEF_NONE)

	minLabel := 1
	minWidth := 50
	maxWidth := 200

	labelLen := len(tabName)

	if labelLen < minLabel {
		return nil, nil
	}

	labelWidth := labelLen // Value fo labelLen is between min and max

	if labelLen < minWidth {
		labelWidth = minWidth
		label.SetEllipsize(pango.ELLIPSIZE_NONE)
	} else if labelLen > maxWidth {
		labelWidth = maxWidth
		//tabName = tabName[:maxWidth] + "..."
		label.SetEllipsize(pango.ELLIPSIZE_END)
	}

	label.SetSizeRequest(labelWidth, -1)

	tabBox.PackStart(label, true, true, 0)
	emptyBox, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
	tabBox.PackStart(emptyBox, false, false, 0)
	tabBox.PackStart(closeButton, false, false, 0)

	closeButton.Connect("clicked", func() {
		var removeID = ""
		removeTab(notebook, paned, vbox, closeButton, removeID)
	})

	tabBox.ShowAll()
	return tabBox, closeButton
}

func removeTab(notebook *gtk.Notebook, paned *gtk.Paned, vbox *gtk.Box, button *gtk.Button, removeID string) {
	var page int
	var e *gtk.Entry
	if button != nil && removeID == "" {
		// Get the page index using the button value
		for pageIndex, tab := range _Common.TabMap {
			if tab.Button == button {
				page = pageIndex
				fmt.Println("From page", pageIndex)
				e = _Common.TabMap[pageIndex].Entry
				break
			}
		}
	} else if button == nil && removeID != "" {
		for pageIndex, tab := range _Common.TabMap {
			if tab.ID == removeID {
				page = pageIndex
				e = _Common.TabMap[pageIndex].Entry
			}
		}
	} else {
		fmt.Println("Something went wrong")
	}

	tab, exists := _Common.TabMap[page]

	if !exists {
		fmt.Println(page, "does not exist")
		return
	}

	notebook.RemovePage(page)

	if tab != nil {
		delete(_Common.TabMap, page)
	} else {
		fmt.Println("tab does not exist for", page)
		return
	}

	for _, info := range _Common.TabMap {
		if info.PageIndex > page {
			// Decrement pageIndex for tabs after the removed page
			info.PageIndex--
		}
	}

	// Update the keys
	updateMap := make(map[int]*_Common.Tabs)
	for _, info := range _Common.TabMap {
		updateMap[info.PageIndex] = info
	}
	_Common.TabMap = updateMap

	vbox.Remove(e)

	newPage := notebook.GetCurrentPage()
	for in, tab := range _Common.TabMap {
		if in == newPage {
			tab.Entry.SetVisible(true)
		} else {
			tab.Entry.SetVisible(false)
		}
	}

	if len(_Common.TabMap) == 0 {
		paned.Remove(notebook)
	}
}

func createFileMenu() *gtk.MenuItem {
	menu, _ := gtk.MenuItemNewWithLabel("File")
	submenu, _ := gtk.MenuNew()

	// "Quit"
	quitItem, _ := gtk.MenuItemNewWithLabel("Quit")
	quitItem.Connect("activate", func() {
		gtk.MainQuit()
	})

	submenu.Append(quitItem)
	menu.SetSubmenu(submenu)

	return menu
}

func createServerMenu() *gtk.MenuItem {
	menu, _ := gtk.MenuItemNewWithLabel("Server")
	submenu, _ := gtk.MenuNew()

	entryItem, _ := gtk.MenuItemNewWithLabel("Set Server IP")

	entryDialog, _ = gtk.DialogNew()
	entryDialog.SetTransientFor(nil)
	entryDialog.SetTitle("Enter Server IP")

	// Create an entry field
	entry, _ := gtk.EntryNew()
	entry.SetHExpand(true)

	// If serverURL is set, prefill the entry field with its value
	if _Common.Shared.ServerURL != "" {
		entry.SetText(_Common.Shared.ServerURL)
	}

	contentArea, err := entryDialog.GetContentArea()
	if err != nil {
		fmt.Println("Error getting content area:", err)
		return menu
	}
	contentArea.Add(entry)

	okButton, _ := entryDialog.AddButton("OK", gtk.RESPONSE_OK)
	cancelButton, _ := entryDialog.AddButton("Cancel", gtk.RESPONSE_CANCEL)

	okButton.Connect("clicked", func() {
		serverIP, _ := entry.GetText()
		_Common.Shared.ServerIP = serverIP
		_Common.Shared.ServerURL = _Common.Shared.ServerIP
		entryDialog.Hide()
	})

	cancelButton.Connect("clicked", func() {
		entryDialog.Hide()
	})

	entryItem.Connect("activate", func() {
		entryDialog.ShowAll()
	})

	startListenerItem, _ := gtk.MenuItemNewWithLabel("Start Listener")

	startListenerDialog, _ := gtk.DialogNew()
	startListenerDialog.SetTransientFor(nil)
	startListenerDialog.SetTitle("Enter IP and Port for Listener")

	box, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 6) // 6 is spacing between children

	// IP entry field
	ipBox, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 6)
	ipLabel, _ := gtk.LabelNew("IP:")
	ipLabel.SetHAlign(gtk.ALIGN_START)
	ipLabel.SetHExpand(true)

	ipEntry, _ := gtk.EntryNew()
	ipEntry.SetHExpand(true)

	ipBox.PackStart(ipLabel, false, false, 0)
	ipBox.PackStart(ipEntry, true, true, 0)

	// Port entry field
	portBox, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 6)
	portLabel, _ := gtk.LabelNew("Port:")
	portLabel.SetHAlign(gtk.ALIGN_START)
	portLabel.SetHExpand(true)

	portEntry, _ := gtk.EntryNew()
	portEntry.SetHExpand(true)

	portBox.PackStart(portLabel, false, false, 0)
	portBox.PackStart(portEntry, true, true, 0)

	// Protocol selection
	protoBox, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 6)
	protoLabel, _ := gtk.LabelNew("Protocol:")
	protoLabel.SetHAlign(gtk.ALIGN_START)
	protoLabel.SetHExpand(true)

	protoCombo, _ := gtk.ComboBoxTextNew()
	protos := _Common.Shared.Protos
	for _, p := range protos {
		protoCombo.AppendText(p)
	}
	protoCombo.SetActive(0)

	protoBox.PackStart(protoLabel, false, false, 0)
	protoBox.PackStart(protoCombo, true, true, 0)

	box.PackStart(ipBox, false, false, 0)
	box.PackStart(portBox, false, false, 0)
	box.PackStart(protoBox, false, false, 0)

	contentAreaStartListener, _ := startListenerDialog.GetContentArea()
	contentAreaStartListener.Add(box)

	okButtonListener, _ := startListenerDialog.AddButton("OK", gtk.RESPONSE_OK)
	cancelButtonListener, _ := startListenerDialog.AddButton("Cancel", gtk.RESPONSE_CANCEL)

	// Handler for OK button
	okButtonListener.Connect("clicked", func() {
		ip, _ := ipEntry.GetText()
		if ip == "" {
			showErrorDialog(startListenerDialog, "Please provide a valid IP or URL\nIP must be formatted as xxx.xxx.xxx.xxx\nURL must be formatted as example.com")
			return
		}

		port, _ := portEntry.GetText()
		if port == "" {
			showErrorDialog(startListenerDialog, "Port must be provided")
			return
		}

		protoIndex := protoCombo.GetActive()
		if protoIndex < 0 || protoIndex >= len(protos) {
			showErrorDialog(startListenerDialog, "Invalid protocol selection")
			return
		}
		proto := protos[protoIndex]

		startListener(ip, port, proto)
		startListenerDialog.Hide()
	})

	cancelButtonListener.Connect("clicked", func() {
		startListenerDialog.Hide()
	})

	startListenerItem.Connect("activate", func() {
		startListenerDialog.ShowAll()
	})

	submenu.Append(entryItem)
	submenu.Append(startListenerItem)

	menu.SetSubmenu(submenu)

	return menu
}

func createBuildMenu() *gtk.MenuItem {
	menu, _ := gtk.MenuItemNewWithLabel("Build")
	submenu, _ := gtk.MenuNew()

	buildImplantItem, _ := gtk.MenuItemNewWithLabel("Build Implant")

	buildImplantItem.Connect("activate", func() {
		dialog, _ := gtk.DialogNew()
		dialog.SetTransientFor(nil)
		dialog.SetTitle("Build Implant")

		box, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 6)

		// Entry for IP
		ipBox, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 6)
		ipLabel, _ := gtk.LabelNew("IP:")
		ipLabel.SetHAlign(gtk.ALIGN_START)
		ipLabel.SetHExpand(true)

		ipEntry, _ := gtk.EntryNew()
		ipEntry.SetHExpand(true)

		ipBox.PackStart(ipLabel, false, false, 0)
		ipBox.PackStart(ipEntry, true, true, 0)

		// Entry for Port
		portBox, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 6)
		portLabel, _ := gtk.LabelNew("Port:")
		portLabel.SetHAlign(gtk.ALIGN_START)
		portLabel.SetHExpand(true)

		portEntry, _ := gtk.EntryNew()
		portEntry.SetHExpand(true)

		portBox.PackStart(portLabel, false, false, 0)
		portBox.PackStart(portEntry, true, true, 0)

		// Protocol selection
		protoBox, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 6)
		protoLabel, _ := gtk.LabelNew("Protocol:")
		protoLabel.SetHAlign(gtk.ALIGN_START)
		protoLabel.SetHExpand(true)

		protoCombo, _ := gtk.ComboBoxTextNew()
		protos := _Common.Shared.Protos
		for _, p := range protos {
			protoCombo.AppendText(p)
		}
		protoCombo.SetActive(0)

		protoBox.PackStart(protoLabel, false, false, 0)
		protoBox.PackStart(protoCombo, true, true, 0)

		box.PackStart(ipBox, false, false, 0)
		box.PackStart(portBox, false, false, 0)
		box.PackStart(protoBox, false, false, 0)

		contentArea, _ := dialog.GetContentArea()
		contentArea.Add(box)

		okButton, _ := dialog.AddButton("OK", gtk.RESPONSE_OK)
		cancelButton, _ := dialog.AddButton("Cancel", gtk.RESPONSE_CANCEL)

		okButton.Connect("clicked", func() {
			ip, _ := ipEntry.GetText()
			if ip == "" {
				showErrorDialog(dialog, "Please provide a valid IP or URL\nIP must be formatted as xxx.xxx.xxx.xxx\nURL must be formatted as example.com")
				return
			}
			port, _ := portEntry.GetText()
			if port == "" {
				showErrorDialog(dialog, "Port must be provided")
				return
			}
			protoIndex := protoCombo.GetActive()
			proto := protos[protoIndex]

			_Common.GoRoutine(builder, ip, port, proto)
			dialog.Hide()
		})

		cancelButton.Connect("clicked", func() {
			dialog.Hide()
		})

		dialog.ShowAll()
	})

	submenu.Append(buildImplantItem)

	menu.SetSubmenu(submenu)

	return menu
}

func startListener(ip string, port string, proto string) {
	cmdGroup := "listener"
	cmd := cmdGroup + " " + proto + "," + ip + "," + port
	_Common.GoRoutine(_Client.NewListener, cmd)
}

func builder(ip string, port string, proto string) {
	builderPath := "./Builder/builder.py"
	cmd := exec.Command("python3", builderPath, ip, port, proto)
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error running build script:", err)
	} else {
		Text := "Build completed for" + ip + ":" + port
		_Common.InsertLogMarkup(Text)
	}

	startListener(ip, port, proto)
}
