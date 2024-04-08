package UI

import (
	"fmt"
	"log"
    "strings"
    "os/exec"
    "sync"

    "github.com/fstanis/screenresolution"
	"github.com/gotk3/gotk3/gtk"
    "github.com/gotk3/gotk3/gdk"
    "github.com/gotk3/gotk3/glib"
    "github.com/gotk3/gotk3/pango"

	"Operator/Common"
	"Operator/Client"
)

var bufferIndex int

var nilButton *gtk.Button = nil
var entryCount = 0

var ScreenWidth int 
var ScreenHeight int

var mRefProvider *gtk.CssProvider

var entry *gtk.Entry
var entryDialog *gtk.Dialog
var win *gtk.Window
var screen *gdk.Screen
var infoTable *gtk.TreeView
var notebook *gtk.Notebook
var paned *gtk.Paned
var vbox *gtk.Box

var connectionMutex sync.Mutex
var cmdMutex sync.Mutex
var ID string
var PlaceHolder = "Siafu>>"


func InitUI() {
    gtk.Init(nil)
    
    win, _ = gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
    win.SetTitle("Siafu Operator")
    win.Connect("destroy", func() {
        gtk.MainQuit()
    })

    // Load CSS stylesheet
    mRefProvider, _  = gtk.CssProviderNew()
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

    Common.LabelMap = make(map[*gtk.Label]int)
    tabName := ""

    fileMenu := createFileMenu(win, notebook, vbox, PlaceHolder, tabName, paned)
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
    Common.Store, _ = gtk.ListStoreNew(glib.TYPE_STRING, glib.TYPE_STRING, glib.TYPE_STRING, glib.TYPE_STRING, glib.TYPE_STRING, glib.TYPE_STRING, glib.TYPE_STRING, glib.TYPE_STRING)
    createTable(infoTable, Common.Store)

    scrolledText, _ := gtk.ScrolledWindowNew(nil, nil)
    topPaned.Pack2(scrolledText, true, true)
    servertxt, _ := gtk.TextViewNew()
    servertxt.SetEditable(false)
    scrolledText.Add(servertxt)

    serverMenu := createServerMenu()
    menubar.Append(serverMenu)

    Common.AL.LogBufferMu.Lock()
    defer Common.AL.LogBufferMu.Unlock()
    Common.AL.LogBuffer, _ = servertxt.GetBuffer()

	initPlaceHolder(entry)
}

func BuildUI() {

    win.SetDefaultSize(ScreenWidth, ScreenHeight)
    win.ShowAll()

    gtk.Main()
}

func Connections() {
    
    infoTable.Connect("button-press-event", func(tv *gtk.TreeView, ev *gdk.Event) {
        connectionMutex.Lock()
        defer connectionMutex.Unlock()
        event := &gdk.EventButton{ev}
        if event.Button() == 3 { // Right mouse button
            // Get the path at the clicked coordinates
            path, _, _, _, _ := tv.GetPathAtPos(int(event.X()), int(event.Y()))
            if path != nil {
                // Select the row at the clicked path
                selection, err := tv.GetSelection()
                if err != nil {
                    log.Fatal("Unable to get selection:", err)
                }
                selection.SelectPath(path)
                // Display the context menu
                var parent *gtk.Window

                Common.CT.CurrentMutex.Lock()
                defer Common.CT.CurrentMutex.Unlock()
                showContextMenu(notebook, vbox, win, PlaceHolder, parent, tv, ev, Common.Store, paned)
            }
        }
    })
    
    notebook.Connect("switch-page", func(notebook *gtk.Notebook, page *gtk.Widget, pageNum int) {
        tab := Common.TabMap[pageNum]

        if tab != nil {
            Common.CT.CurrentMutex.Lock()
            defer Common.CT.CurrentMutex.Unlock()
        
            Common.CT.CurrentID = tab.ID
            Common.CT.CurrentBuffer = tab.Buffer
            Common.CT.CurrentEntry = tab.Entry
            Common.CT.CurrentPage = tab.PageIndex

        } 
        curpg := notebook.GetCurrentPage()
        for in, tab := range Common.TabMap {
            if in == curpg {
                tab.Entry.SetVisible(true)
            } else {
                tab.Entry.SetVisible(false)
            }
        }
    })
    
}

func createTable(infoTable *gtk.TreeView, store *gtk.ListStore) {
    columns := []string{"OS", "Agent Type", "UID", "Host Name", "User", "Internal IP(s)", "External IP(s)", "Last Seen"}

    // Set the list store as the model for the tree view
    infoTable.SetModel(store)

    // Create columns
    for i, columnName := range columns {
        rows, err := gtk.CellRendererTextNew()
        if err != nil {
            log.Fatal("Unable to create cell renderer:", err)
        }
        rows.SetAlignment(0.5, 0.5)
        column, err := gtk.TreeViewColumnNewWithAttribute(columnName, rows, "text", i)
        if err != nil {
            log.Fatal("Unable to create tree view column:", err)
        }
        //column.SetAlignment(0.5)
        // Add column to tree view
        infoTable.AppendColumn(column)
    }

}

func showContextMenu(notebook *gtk.Notebook, vbox *gtk.Box, win *gtk.Window, PlaceHolder string, parent *gtk.Window, tv *gtk.TreeView, ev *gdk.Event, store *gtk.ListStore, paned *gtk.Paned) {
	menu, _ := gtk.MenuNew()
    
	// "Open Console" 
	menuItemOpenConsole, _ := gtk.MenuItemNewWithLabel("Open Console")
	menuItemOpenConsole.Connect("activate", func() {
        newConsole(notebook, vbox, win, PlaceHolder, tv, store, paned)
	})
	menu.Append(menuItemOpenConsole)

	// "Kill Implant" 
	menuItemKillImplant, _ := gtk.MenuItemNewWithLabel("Kill Implant")
	menuItemKillImplant.Connect("activate", func() {
		dialog := gtk.MessageDialogNew(parent, gtk.DIALOG_MODAL, gtk.MESSAGE_QUESTION, gtk.BUTTONS_YES_NO, "Are you sure?\n This will remove the implant from its host")
		response := dialog.Run()
		if response == gtk.RESPONSE_YES {
            removeImplant(tv, store, notebook, paned, win, vbox)
		}
		dialog.Destroy()
	})
	menu.Append(menuItemKillImplant)
	menu.ShowAll()
	menu.PopupAtPointer(nil)
}

func setupTab(notebook *gtk.Notebook, vbox *gtk.Box, win *gtk.Window, PlaceHolder string, tabName string, paned *gtk.Paned) {
    fmt.Println("setuptab")
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

    tabWidget, button  := createTabWidget(tabLabel, notebook, paned, win, vbox, tabName)

    // Append the custom tab
    notebook.AppendPage(cmdBox, tabWidget)
    notebook.SetScrollable(true)
    // Determine the index
    pageIndex := notebook.GetNPages() - 1

    // Create a unique entry field for each tab
    cmdentry := createEntry()

    initPlaceHolder(cmdentry)
    
    i++

    cmdentry.SetName("entrytab_" + string(i))

    // Create a unique buffer for each tab
    cmdbuffer, _ := cmdtxt.GetBuffer()
    bufferIndex++
    vbox.PackEnd(cmdentry, false, false, 0)

    Common.TabMap[pageIndex] = &Common.Tabs{
        ID: ID,
        PageIndex: pageIndex,
        TabLabel:   tabLabel,
        Buffer: cmdbuffer,
        Entry:      cmdentry,
        Button: button,
    }

    cmdentry.Connect("activate", func() {
        cmd, _ := cmdentry.GetText()

        handleCmd(cmd)
        cmdentry.SetText("")
    })

    notebook.ShowAll()
    newPage := notebook.GetCurrentPage()
        for in, tab := range Common.TabMap {
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

func removeImplant(tv *gtk.TreeView, store *gtk.ListStore, notebook *gtk.Notebook, paned *gtk.Paned, win *gtk.Window, vbox *gtk.Box) {
    selection, _ := tv.GetSelection()
	_ , iter, _ := selection.GetSelected()

        col2Value, _ := store.GetValue(iter, 2) // ID
        col2, _ := col2Value.GetString()
        removeID := col2

    fmt.Println("Remove:", removeID)
	if iter != nil {
		store.Remove(iter)
        removeTab(notebook, paned, vbox, win, nilButton, removeID)
	}
}

func newConsole(notebook *gtk.Notebook, vbox *gtk.Box, win *gtk.Window, PlaceHolder string, tv *gtk.TreeView, store *gtk.ListStore, paned *gtk.Paned) {
    fmt.Println("newconsole")
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

func createTabWidget(label *gtk.Label, notebook *gtk.Notebook, paned *gtk.Paned, win *gtk.Window, vbox *gtk.Box, tabName string) (*gtk.Box, *gtk.Button) {
    tabBox, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)

    closeButton, _ := gtk.ButtonNew()
    closeButton.SetLabel("x")
    closeButton.SetRelief(gtk.RELIEF_NONE) // Remove button border

    minLabel := 1
    minWidth := 50 // Minimum width for the tab
    maxWidth := 200 // Maximum width for the tab

    labelLen := len(tabName)

    if labelLen < minLabel {
        return nil, nil
    }

    labelWidth := labelLen // Value fo labelLen is between min and max

    // Ensure the label width is within the bounds of minWidth and maxWidth
    if labelLen < minWidth {
        labelWidth = minWidth
        label.SetEllipsize(pango.ELLIPSIZE_NONE)
    } else if labelLen > maxWidth {
        labelWidth = maxWidth
        tabName = tabName[:maxWidth] + "..."
        label.SetEllipsize(pango.ELLIPSIZE_END)
    }

    // Set the size request for the label
    label.SetSizeRequest(labelWidth, -1)

    tabBox.PackStart(label, true, true, 0)
    // Add some padding between the label and the close button
    emptyBox, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
    tabBox.PackStart(emptyBox, false, false, 0)
    // Pack close button without expansion
    tabBox.PackStart(closeButton, false, false, 0)

    closeButton.Connect("clicked", func() {
        var removeID = ""
        removeTab(notebook, paned, vbox, win, closeButton, removeID)
    })

    tabBox.ShowAll()
    return tabBox, closeButton
}

func removeTab(notebook *gtk.Notebook, paned *gtk.Paned, vbox *gtk.Box, win *gtk.Window, button *gtk.Button, removeID string) {
    var page int
    var e *gtk.Entry
    if button != nil && removeID == "" {
        // Get the page index using the button value
        for pageIndex, tab := range Common.TabMap {
            if tab.Button == button {
                page = pageIndex
                fmt.Println("From page", pageIndex)
                e = Common.TabMap[pageIndex].Entry
                break
            }
        } 
    } else if button == nil && removeID != ""{
        for pageIndex, tab := range Common.TabMap {
            if tab.ID == removeID {
                page = pageIndex
                e = Common.TabMap[pageIndex].Entry
            }
        } 
    } else {
        fmt.Println("Something went wrong")
    }

    tab, exists := Common.TabMap[page]

    if !exists {
        fmt.Println(page, "does not exist")
        return 
    }

    notebook.RemovePage(page)

    if tab != nil {
        delete(Common.TabMap, page)
    } else {
        fmt.Println("tab does not exist for", page)
        return
    }

    for _, info := range Common.TabMap {
        if info.PageIndex > page {
            // Decrement pageIndex for tabs after the removed page
            info.PageIndex--
        }
    }

    // Update the keys 
    updateMap := make(map[int]*Common.Tabs)
    for _, info := range Common.TabMap {
        updateMap[info.PageIndex] = info
    }
    Common.TabMap = updateMap

    vbox.Remove(e)

    newPage := notebook.GetCurrentPage()
        for in, tab := range Common.TabMap {
            if in == newPage {
                tab.Entry.SetVisible(true)
            } else {
                tab.Entry.SetVisible(false)
            }
        }

    if len(Common.TabMap) == 0 {
        paned.Remove(notebook)
    }
}


func createFileMenu(win *gtk.Window, notebook *gtk.Notebook, vbox *gtk.Box, PlaceHolder string, tabName string, paned *gtk.Paned) *gtk.MenuItem {
    menu, _ := gtk.MenuItemNewWithLabel("File")
    submenu, _ := gtk.MenuNew()
    
    // "Quit"
    quitItem, _ := gtk.MenuItemNewWithLabel("Quit")
    quitItem.Connect("activate", func() {
        gtk.MainQuit()
    })

    //submenu.Append(newTabItem)
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
    entry.SetHExpand(true) // Allow entry to expand horizontally

    // If serverURL is set, prefill the entry field with its value
    if Common.ServerURL != "" {
        entry.SetText(Common.ServerURL)
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
		Common.ServerIP = serverIP
        fmt.Println("Server IP:", Common.ServerIP)
        Common.ServerURL = Common.ServerIP
        entryDialog.Hide()
    })

    cancelButton.Connect("clicked", func() {
        entryDialog.Hide()
    })

    entryItem.Connect("activate", func() {
        entryDialog.ShowAll()
    })

    // Menu item for starting listener
    startListenerItem, _ := gtk.MenuItemNewWithLabel("Start Listener")

    startListenerDialog, _ := gtk.DialogNew()
    startListenerDialog.SetTransientFor(nil)
    startListenerDialog.SetTitle("Enter IP and Port for Listener")

    // Create a box to hold the labels, entries, and protocol selection horizontally
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

    // Protocol selection (assuming you have protos []string defined somewhere)
    protoBox, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 6)
    protoLabel, _ := gtk.LabelNew("Protocol:")
    protoLabel.SetHAlign(gtk.ALIGN_START)
    protoLabel.SetHExpand(true)

    protoCombo, _ := gtk.ComboBoxTextNew()
	protos := Common.Protos
    for _, p := range protos {
        protoCombo.AppendText(p)
    }
    protoCombo.SetActive(0)

    protoBox.PackStart(protoLabel, false, false, 0)
    protoBox.PackStart(protoCombo, true, true, 0)

    // Adding IP, Port, and Protocol boxes to the main box
    box.PackStart(ipBox, false, false, 0)
    box.PackStart(portBox, false, false, 0)
    box.PackStart(protoBox, false, false, 0)

    // Adding the main box to the content area of the dialog
    contentAreaStartListener, _ := startListenerDialog.GetContentArea()
    contentAreaStartListener.Add(box)

    // OK and Cancel buttons
    okButtonListener, _ := startListenerDialog.AddButton("OK", gtk.RESPONSE_OK)
    cancelButtonListener, _ := startListenerDialog.AddButton("Cancel", gtk.RESPONSE_CANCEL)

    // Handler for OK button click
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
    
        // Add your listener logic here
        go startListener(ip, port, proto)
        startListenerDialog.Hide()
    })

    cancelButtonListener.Connect("clicked", func() {
        startListenerDialog.Hide()
    })

    startListenerItem.Connect("activate", func() {
        startListenerDialog.ShowAll()
    })

    // Appending items to submenu
    submenu.Append(entryItem)
    submenu.Append(startListenerItem)

    menu.SetSubmenu(submenu)

    return menu
}

func createBuildMenu() *gtk.MenuItem {
    menu, _ := gtk.MenuItemNewWithLabel("Build")
    submenu, _ := gtk.MenuNew()

    // Menu item for building implant
    buildImplantItem, _ := gtk.MenuItemNewWithLabel("Build Implant")

    // Connect the "activate" signal to a function to handle the build action
    buildImplantItem.Connect("activate", func() {
        // Create the dialog for building the implant
        dialog, _ := gtk.DialogNew()
        dialog.SetTransientFor(nil)
        dialog.SetTitle("Build Implant")

        // Create a box to hold the labels, entries, and protocol selection horizontally
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
		protos := Common.Protos
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

            fmt.Println("Listener IP:", ip)
            fmt.Println("Listener Port:", port)
            protoIndex := protoCombo.GetActive()
            proto := protos[protoIndex]
            // Add your listener logic here
            builder(ip, port, proto)
            dialog.Hide()
        })

        cancelButton.Connect("clicked", func() {
            dialog.Hide()
        })

        dialog.ShowAll()
    })

    // Append the build implant item to the submenu
    submenu.Append(buildImplantItem)

    // Set the submenu for the "Build" menu item
    menu.SetSubmenu(submenu)

    return menu
}

func handleCmd(cmd string) {
    fmt.Println("handlecmd")
    Common.CT.CurrentMutex.Lock()
    defer Common.CT.CurrentMutex.Unlock()

    buffer := Common.CT.CurrentBuffer
    
    var sort Common.Sorting

    cmdMutex.Lock()
    defer cmdMutex.Unlock()

    iter := buffer.GetEndIter()
    buffer.Insert(iter, fmt.Sprintf("%s %s\n", PlaceHolder, cmd))
    go func() { 
        Client.RouteCMD(cmd, Common.CommandChan)

        close(Common.CommandChan)
   
        sort.SortingMutex.Lock()
        defer sort.SortingMutex.Unlock()
        output, ok := <-Common.CommandChan // from http implantcmd
        fmt.Println("Output", output)
        if sort.ImplantCmd && !ok {
            fmt.Println("Response channel was closed unexpectedly")
            return
        }
        Common.CT.CurrentMutex.Lock()
        defer Common.CT.CurrentMutex.Unlock()
        Common.CommandOutput(output, Common.CT.CurrentBuffer)
    }()
}

func startListener(ip string, port string, proto string) {
    cmdGroup := "listener"

    cmdString := proto + "," + ip + "," +  port
    responseChan := make(chan string)
    go Client.ServerCommand(cmdGroup, cmdString, responseChan)
    
    output := <-responseChan

    if strings.Contains(output, "listener started") {
        output := "Listener started at " + proto + "://" + ip + ":" + port + "\n"
        Common.InsertLogMarkup(output)
    } else {
        //output := "Unable to start listener"
        Common.InsertLogMarkup(output)
    }

}

func builder(ip string, port string, proto string) {
    // Command to run the Python script
    cmd := exec.Command("python3", "./Builder/builder.py", ip, port, proto)

    fmt.Println(cmd)
    // Execute the command
    err := cmd.Run()
    if err != nil {
        fmt.Println("Error running build script:", err)
    }
    
    go startListener(ip, port, proto)
    
}
