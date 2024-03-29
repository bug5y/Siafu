
package main

import (
    "fmt"
    "net"
    "net/http"
    "strings"
    "bytes"
    "encoding/base64"
    "encoding/json"
    "io/ioutil"
    "time"
    "github.com/fstanis/screenresolution"
    "github.com/gotk3/gotk3/gtk"
    "github.com/gotk3/gotk3/gdk"
    "github.com/gotk3/gotk3/glib"
    "github.com/gotk3/gotk3/pango"
    "log"
    "os/exec"
)

const (
    operatorID = "123abc"

    lightBlue = "#90afc5" /* selected items & other interactive components */
    brightBlue = "#1e90ff" /* use for notifications */
    brightRed = "#e14e19" /* use for error notifications */
    allWhite = "#ffffff" /* text */
)

// Windows-10 theme
// font scale .85

type Command struct {
    Group    string `json:"Group"`
    String   string `json:"String"`
    Response string `json:"Response"`
}
var serverURL string
var serverIP string
var implant_cmd bool
var server_cmd bool
var ID_Set bool
var response string
var entryDialog *gtk.Dialog
var currentBuffer *gtk.TextBuffer
var currentEntry *gtk.Entry 
var currentPage int
var ID string
var currentID string
var logBuffer *gtk.TextBuffer
var labelMap map[*gtk.Label]int
var nilButton *gtk.Button = nil
var protos = []string{"HTTP", "HTTPS"}
var mRefProvider *gtk.CssProvider
var port string
var defaultPort = "8443"
var screen *gdk.Screen

func main() {
    gtk.Init(nil)
    
    win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
    if err != nil {
        fmt.Println("Unable to create window:", err)
    }
    win.SetTitle("Siafu C2 Operator")
    win.Connect("destroy", func() {
        gtk.MainQuit()
    })

    // Load CSS stylesheet
    mRefProvider, _  = gtk.CssProviderNew()
    mRefProvider.LoadFromPath("./UI/Material-DeepOcean/gtk-dark.css")
 
    // Apply to whole app
    screen, _ = gdk.ScreenGetDefault()
    gtk.AddProviderForScreen(screen, mRefProvider, gtk.STYLE_PROVIDER_PRIORITY_USER)

    ScreenWidth := 0
    ScreenHeight := 0
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
    vbox, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 10)
    entry, _ := gtk.EntryNew()
    notebook, _ := gtk.NotebookNew()
    paned, _ := gtk.PanedNew(gtk.ORIENTATION_VERTICAL)

    // Create a new scrolled window
    labelMap = make(map[*gtk.Label]int)
    cmdPlaceHolder := "Siafu>>"
    tabName := ""
    fileMenu := createFileMenu(win, notebook, vbox, cmdPlaceHolder, tabName, paned)
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
    treeView, err := gtk.TreeViewNew()
    if err != nil {
        log.Fatal("Unable to create tree view:", err)
    }
    topPaned.Pack1(treeView, true, true)
    TableWidth := int(float64(ScreenWidth) * 0.5)
    topPaned.SetPosition(TableWidth)

    // Create a list store
    store, err := gtk.ListStoreNew(glib.TYPE_STRING, glib.TYPE_STRING, glib.TYPE_STRING, glib.TYPE_STRING, glib.TYPE_STRING, glib.TYPE_STRING, glib.TYPE_STRING, glib.TYPE_STRING)
    if err != nil {
        log.Fatal("Unable to create list store:", err)
    }

    createTable(treeView, store)

    treeView.Connect("button-press-event", func(tv *gtk.TreeView, ev *gdk.Event) {
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
                showContextMenu(notebook, vbox, win, cmdPlaceHolder, parent, tv, ev, store, paned)
            }
        }
    })

    // Debugging
    data := []string{"Linux", "Desktop", "123", "192.168.1.100", "192.168.0.1, 192.168.0.2", "hostname", "user1", "2024-03-22"}
    createRow(store, data)

    data = []string{"Linux", "Desktop", "234", "192.168.1.100", "192.168.0.1, 192.168.0.2", "hostname", "user1", "2024-03-22"}
    createRow(store, data)

    scrolledText, _ := gtk.ScrolledWindowNew(nil, nil)
    topPaned.Pack2(scrolledText, true, true)
    servertxt, _ := gtk.TextViewNew()
    servertxt.SetEditable(false)
    scrolledText.Add(servertxt)

    if serverURL == "" {
        defaultIP, err := getIP()
        if err != nil {
            Text := "Failed to set an address for the server"
            insertLogText(Text)
        }
        port = defaultPort
        serverURL = "http://" + defaultIP + ":" + port
    }

    serverMenu := createServerMenu()
    menubar.Append(serverMenu)
    // Verify server is up -- serverURL = "http://" + defaultIP + ":" + port
    logBuffer, _ = servertxt.GetBuffer()

    if verifyServerUp(serverURL) {
        Text := "<span>Connected to server: <span foreground=\"" + allWhite + "\">" + serverURL + "</span></span>"
            insertLogMarkup(Text)

    } else {
        Text := "<span foreground=\"" + brightRed + "\">Unable to connect to server: <span foreground=\"" + allWhite + "\">" + serverURL + "</span></span>"
        insertLogMarkup(Text)
    }

    go func() {
        for {
            cmdGroup := "log"
            cmdString := ""
            responseChan := make(chan string)
            go func() {
                servercommand(cmdGroup, cmdString, responseChan)
            }()
            // Wait for the response or timeout
            select {
            case logText := <-responseChan:
                if logText != "" {
                    insertLogText(logText)
                }
            
            case <-time.After(5 * time.Second):
            }
    
            // Sleep for 5 seconds before fetching logs again
            time.Sleep(5 * time.Second)
        }
    }()

    // Connect signal to handle tab switch
    notebook.Connect("switch-page", func(notebook *gtk.Notebook, page *gtk.Widget, pageNum int) {
        tabInfo := tabInfoMap[pageNum]

        if tabInfo != nil {
            currentID = tabInfo.ID
            currentBuffer = tabInfo.Buffer
            currentEntry = tabInfo.Entry
            currentPage = tabInfo.PageIndex
        } 
        curpg := notebook.GetCurrentPage()
        for in, tab := range tabInfoMap {
            if in == curpg {
                tab.Entry.SetVisible(true)
            } else {
                tab.Entry.SetVisible(false)
            }
        }


    })
    
    initPlaceHolder(entry, cmdPlaceHolder)

    // Set the default size and show all widgets
    win.SetDefaultSize(ScreenWidth, ScreenHeight)
    win.ShowAll()

    // Start the GTK main event loop
    gtk.Main()
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
        // UI
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var bufferIndex int

type TabInfo struct {
    PageIndex int
    ID string
    TabLabel   *gtk.Label       // Label for the tab
    Buffer *gtk.TextBuffer           // Name of the buffer associated with the tab
    Entry      *gtk.Entry       // Entry widget associated with the tab
    Button *gtk.Button
}

var tabInfoMap = make(map[int]*TabInfo)


func createTable(treeView *gtk.TreeView, store *gtk.ListStore) {
    columns := []string{"OS", "Agent Type", "Implant ID", "External IP", "Internal IP(s)", "Host Name", "User", "Last Seen"}

    // Set the list store as the model for the tree view
    treeView.SetModel(store)

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
        treeView.AppendColumn(column)
    }

}

// Function to add a row to the paned widget
func createRow(store *gtk.ListStore, rowData []string) {
    iter := store.Append()

    for i, value := range rowData {
        store.SetValue(iter, i, value)
    }
}

func showContextMenu(notebook *gtk.Notebook, vbox *gtk.Box, win *gtk.Window, cmdPlaceHolder string, parent *gtk.Window, tv *gtk.TreeView, ev *gdk.Event, store *gtk.ListStore, paned *gtk.Paned) {
	menu, _ := gtk.MenuNew()
    
	// "Open Console" 
	menuItemOpenConsole, _ := gtk.MenuItemNewWithLabel("Open Console")
	menuItemOpenConsole.Connect("activate", func() {
        newConsole(notebook, vbox, win, cmdPlaceHolder, tv, store, paned)
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

var entryCount = 0
var entry *gtk.Entry

func createEntry() *gtk.Entry {
    entry, _ := gtk.EntryNew()
    entryCount = entryCount + 1
    return entry
}

func setupTab(notebook *gtk.Notebook, vbox *gtk.Box, win *gtk.Window, cmdPlaceHolder string, tabName string, paned *gtk.Paned) {
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
    entry := createEntry()

    initPlaceHolder(entry, cmdPlaceHolder)
    
    entry.SetName("entrytab_" + string(i))

    // Create a unique buffer for each tab
    buffer, _ := cmdtxt.GetBuffer()
    bufferIndex++
    vbox.PackEnd(entry, false, false, 0)

    // Store tab information in the map with tabLabel text as the key
    tabInfoMap[pageIndex] = &TabInfo{
        ID: ID,
        PageIndex: pageIndex,
        TabLabel:   tabLabel,
        Buffer: buffer,
        Entry:      entry,
        Button: button,
    }

    entry.Connect("activate", func() {
        cmd, _ := entry.GetText()
        handleCmd(cmd, currentBuffer, currentEntry, cmdPlaceHolder)
        entry.SetText("")
    })
    notebook.ShowAll()
    newPage := notebook.GetCurrentPage()
        for in, tab := range tabInfoMap {
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

func newConsole(notebook *gtk.Notebook, vbox *gtk.Box, win *gtk.Window, cmdPlaceHolder string, tv *gtk.TreeView, store *gtk.ListStore, paned *gtk.Paned) {
    selection, _ := tv.GetSelection()
    _, iter, _ := selection.GetSelected()

    model, err := tv.GetModel()
    if err != nil {
    }

    listStore, ok := model.(*gtk.ListStore)
    if !ok {
    }

    if iter != nil {
        // Retrieve data from each column and convert to string
        col2Value, _ := listStore.GetValue(iter, 2) // ID
        col2, _ := col2Value.GetString()

        col5Value, _ := listStore.GetValue(iter, 5) // User
        col5, _ := col5Value.GetString()

        col6Value, _ := listStore.GetValue(iter, 6) // Host
        col6, _ := col6Value.GetString()

        tabName := col2 + ": " + col6 + "@" + col5

        ID = col2

        for a := 0; a < 50; a++ {
            // Calculate tab name
            tabName := strings.Repeat("A", a+1)
            setupTab(notebook, vbox, win, cmdPlaceHolder, tabName, paned)
        }

        setupTab(notebook, vbox, win, cmdPlaceHolder, tabName, paned)

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

func printTabInfoMap() {
    fmt.Println("Printing tabInfoMap contents:")
    for key, value := range tabInfoMap {
        fmt.Printf("Key: %v, Value: {ID: %v, PageIndex: %v, Button: %v}\n", key, value.ID, value.PageIndex, value.Button)
    }
}

    //if button = nil {
    //pageIndex := notebook.PageNum(id)
    //fmt.Println("From ID", pageIndex)
//}

func removeTab(notebook *gtk.Notebook, paned *gtk.Paned, vbox *gtk.Box, win *gtk.Window, button *gtk.Button, removeID string) {
    var page int
    var e *gtk.Entry
    if button != nil && removeID == "" {
        // Get the page index using the button value
        for pageIndex, tabInfo := range tabInfoMap {
            if tabInfo.Button == button {
                page = pageIndex
                fmt.Println("From page", pageIndex)
                e = tabInfoMap[pageIndex].Entry
                break
            }
        } 
    } else if button == nil && removeID != ""{
        for pageIndex, tabInfo := range tabInfoMap {
            if tabInfo.ID == removeID {
                page = pageIndex
                e = tabInfoMap[pageIndex].Entry
            }
        } 
    } else {
        fmt.Println("Something went wrong")
    }

    tabInfo, exists := tabInfoMap[page]

    if !exists {
        fmt.Println(page, "does not exist")
        return 
    }

    notebook.RemovePage(page)

    if tabInfo != nil {
        delete(tabInfoMap, page)
    } else {
        fmt.Println("tabInfo does not exist for", page)
        return
    }

    for _, info := range tabInfoMap {
        if info.PageIndex > page {
            // Decrement pageIndex for tabs after the removed page
            info.PageIndex--
        }
    }

    // Update the keys 
    updateMap := make(map[int]*TabInfo)
    for _, info := range tabInfoMap {
        updateMap[info.PageIndex] = info
    }
    tabInfoMap = updateMap

    vbox.Remove(e)

    newPage := notebook.GetCurrentPage()
        for in, tab := range tabInfoMap {
            if in == newPage {
                tab.Entry.SetVisible(true)
            } else {
                tab.Entry.SetVisible(false)
            }
        }

    if len(tabInfoMap) == 0 {
        paned.Remove(notebook)
    }
}


func createFileMenu(win *gtk.Window, notebook *gtk.Notebook, vbox *gtk.Box, cmdPlaceHolder string, tabName string, paned *gtk.Paned) *gtk.MenuItem {
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
    if serverURL != "" {
        entry.SetText(serverURL)
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
        fmt.Println("Server IP:", serverIP)
        serverURL = serverIP
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
    
        fmt.Println("Listener IP:", ip)
        fmt.Println("Listener Port:", port)
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

func showErrorDialog(parent *gtk.Dialog, message string) {
    dialog := gtk.MessageDialogNew(parent, gtk.DIALOG_MODAL, gtk.MESSAGE_ERROR, gtk.BUTTONS_OK, message)
    dialog.SetTitle("Error")
    dialog.Run()
    dialog.Destroy()
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
            go startListener(ip, port, proto)
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

func initPlaceHolder(entry *gtk.Entry, cmdPlaceHolder string) {
    entry.SetPlaceholderText(cmdPlaceHolder)
}

func handleCmd(cmd string, buffer *gtk.TextBuffer, entry *gtk.Entry, cmdPlaceHolder string) {
    iter := buffer.GetEndIter()
    buffer.Insert(iter, fmt.Sprintf("%s %s\n", cmdPlaceHolder, cmd))
    responseChan := make(chan string)

    // Execute the command asynchronously
    go func() {
        routeCMD(cmd, responseChan, buffer, entry)

        close(responseChan)
    }()
    
    output, ok := <-responseChan
    if implant_cmd && !ok {
        fmt.Println("Response channel was closed unexpectedly")
        return
    }
    handleOutput(output, buffer)
}


func verifyServerUp(serverURL string) bool {
    resp, err := http.Get(serverURL)
    if err != nil {
        return false
    }
    defer resp.Body.Close()
    return true
}

func startListener(ip string, port string, proto string) {
    fmt.Print(proto, " listener at ", ip + ":" + port)

    cmdGroup := "listener"

    cmdString := proto + "," + ip + "," +  port
    responseChan := make(chan string)
    go servercommand(cmdGroup, cmdString, responseChan)
    
    output := <-responseChan
    fmt.Println(output)

    if strings.Contains(output, "listener started") {
        output := proto + " listener started at " + serverIP + ":" + port
        insertLogText(output)
    } else {
        output := "Unable to start listener"
        insertLogMarkup(output)
    }

}

func builder(ip string, port string, proto string) {
    // Command to run the Python script
    cmd := exec.Command("python3", "./Build/builder.py", ip, port, proto)

    // Execute the command
    err := cmd.Run()
    if err != nil {
        fmt.Println("Error running build script:", err)
    }
    
    go startListener(ip, port, proto)
    
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
        // Route and Send CMDs
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////


func routeCMD(input string, responseChan chan<- string, buffer *gtk.TextBuffer, entry *gtk.Entry) {
        // Trim spaces and newline characters from the input
        input = strings.TrimSpace(input)
        
        // Parse the input command
        parts := strings.Fields(input)
        if len(parts) < 2 {
            output := "Invalid command. Please provide a command group and a command.\n"
            warn := true
            handleAlerts(output, buffer, warn)
            return
        }

        cmdGroup := parts[0]
        switch cmdGroup {
        case "shell":
            implant_cmd = true
        case "implant":
            ID_Set = false
            for _, part := range parts {
                if strings.HasPrefix(part, "-s") {
                    fmt.Println("Remove this")
                }
            }
            if !ID_Set {
                server_cmd = true
            }
            
        default:
            output := "Invalid command group\n"
            warn := true
            handleAlerts(output, buffer, warn)
            return
        }
        cmdString := strings.Join(parts[1:], " ")

        if implant_cmd {
            err := sendCommand(cmdGroup, cmdString, responseChan, currentBuffer, currentID)
            if err != nil {
                output := fmt.Sprintf("Error sending command: %v\n", err)
                warn := true
                handleAlerts(output, buffer, warn)
                return
            }
        }
        
        if server_cmd {
            
            err := servercommand(cmdGroup, cmdString, responseChan)
            if err != nil {
                output := fmt.Sprintf("Error sending command: %v\n", err)
                warn := true
                handleAlerts(output, buffer, warn)
                return
            }
        }
    }


func sendCommand(cmdGroup, cmdString string, responseChan chan<- string, buffer *gtk.TextBuffer, implantID string) error {
    endpoint := "/operator"
    server := serverURL + endpoint
    client := &http.Client{}

    commandURL := server
    
    cmdData := Command{
        Group:   cmdGroup,
        String:  cmdString,
        Response: "", 
    }

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
        // Send request to server
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
        
    // Serialize command data to JSON
    jsonData, err := json.Marshal(cmdData)
    if err != nil {
        return fmt.Errorf("error marshaling JSON data: %w", err)
    }

    // Base64 encode the JSON data
    base64Data := base64.StdEncoding.EncodeToString(jsonData)

    // Create HTTP request
    req, err := http.NewRequest("POST", commandURL, bytes.NewBuffer([]byte(base64Data)))
    if err != nil {
        return fmt.Errorf("error creating HTTP request: %w", err)
    }
    req.Header.Set("Content-Type", "text/plain") // Set content type to text/plain
    req.Header.Set("Operator-ID", operatorID)
    req.Header.Set("Implant-ID", implantID)

    // Send request
    resp, err := client.Do(req)
    if err != nil {
        return err
    } else {
        output := "Task sent to implant\n"
        warn := false
        handleAlerts(output, buffer, warn)
    }
    
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
        // Receive response from server
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

    var responseBytes []byte
    var responseBody string
    var decodedData []byte
    for {
        // Read the response
        responseBytes, err = ioutil.ReadAll(resp.Body)
        if err != nil {
            return fmt.Errorf("error reading response body: %w", err)
        }
        responseBody = string(responseBytes) // Convert []byte to string
    // Extract and decode data value
        dataIndex := strings.Index(responseBody, "Data:")
        if dataIndex != -1 {
            data := strings.TrimSpace(responseBody[dataIndex+len("Data:"):])
            decodedData, err = base64.StdEncoding.DecodeString(data)
            if err != nil {
                return fmt.Errorf("error decoding base64 data: %w", err)
            }
        } else {
            output := "No data found in the response\n"
            warn := true
            handleAlerts(output, buffer, warn)
        }
        if len(responseBody) > 0 {
            output := "Received ouput\n"
            warn := false
            handleAlerts(output, buffer, warn)
            break
        }
        time.Sleep(100 * time.Millisecond) // Wait for a short duration before trying again
    }
    

    // Unmarshal JSON
   var respstruct Command

   // Trim leading and trailing whitespace
   decodedData = bytes.TrimSpace(decodedData)
   
   err = json.Unmarshal(decodedData, &respstruct)
   if err != nil {
       return fmt.Errorf("Failed to unmarshal JSON: %s", err)
   }

    responseChan <- respstruct.Response
    defer resp.Body.Close()
    return nil
}


func servercommand(cmdGroup, cmdString string, responseChan chan<- string) error {
    endpoint := "/info"
    server := serverURL + endpoint
    client := &http.Client{}

    commandURL := server
    
    cmdData := Command{
        Group:   cmdGroup,
        String:  cmdString,
        Response: "", 
    }

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
        // Send request to server
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
        
    // Serialize command data to JSON
    jsonData, err := json.Marshal(cmdData)
    if err != nil {
        return fmt.Errorf("error marshaling JSON data: %w", err)
    }

    // Base64 encode the JSON data
    base64Data := base64.StdEncoding.EncodeToString(jsonData)

    // Create HTTP request
    req, err := http.NewRequest("GET", commandURL, bytes.NewBuffer([]byte(base64Data)))
    if err != nil {
        return fmt.Errorf("error creating HTTP request: %w", err)
    }
    req.Header.Set("Content-Type", "text/plain") // Set content type to text/plain
    req.Header.Set("Operator-ID", operatorID)

    // Send request
    resp, err := client.Do(req)
    if err != nil {
        return err
    }

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
        // Receive response from server
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
        
    var responseBytes []byte
    var responseBody string
    var decodedData []byte
    for {
        // Read the response
        responseBytes, err = ioutil.ReadAll(resp.Body)
        if err != nil {
            return fmt.Errorf("error reading response body: %w", err)
        }
        responseBody = string(responseBytes) // Convert []byte to string
    // Extract and decode data value
        dataIndex := strings.Index(responseBody, "Data:")
        if dataIndex != -1 {
            data := strings.TrimSpace(responseBody[dataIndex+len("Data:"):])
            decodedData, err = base64.StdEncoding.DecodeString(data)
            if err != nil {
                return fmt.Errorf("error decoding base64 data: %w", err)
            }
        } else {
            fmt.Println("Data not found in response")
        }
        if len(responseBody) > 0 {
            break
        }
        time.Sleep(100 * time.Millisecond) // Wait for a short duration before trying again
    }
    
    // Unmarshal JSON
    var respstruct Command

   // Trim leading and trailing whitespace
    decodedData = bytes.TrimSpace(decodedData)
   
    err = json.Unmarshal(decodedData, &respstruct)
    if err != nil {
        return fmt.Errorf("Failed to unmarshal JSON: %s", err)
    }
   
    responseChan <- respstruct.Response
    defer resp.Body.Close()
    return nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
        // Helpers
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var indicator = "<span>[" + "<span foreground=\"" + lightBlue + "\">+</span>" + "]</span>"

func handleOutput(output string, buffer *gtk.TextBuffer){ // Inserts to CMD console
    // Insert the output into the text buffer
    iter := buffer.GetEndIter()
    buffer.Insert(iter, output)
}

func handleAlerts(output string, buffer *gtk.TextBuffer, warn bool){ // Inserts to CMD console
    var alertcolor string

    if warn {
        alertcolor = brightRed    
    } else {
        alertcolor = brightBlue 
    }

    alertindicator := "<span foreground=\"" + alertcolor + "\">[</span>" + "<span foreground=\"" + lightBlue + "\">+</span>" + "<span foreground=\"" + alertcolor + "\">]</span> "
    output = alertindicator + output

    iter := buffer.GetEndIter()
    buffer.InsertMarkup(iter, output)
}

func getIP() (string, error) {
    addrs, err := net.InterfaceAddrs()
    if err != nil {
        return "", fmt.Errorf("failed to retrieve interface addresses: %v", err)
    }

    for _, addr := range addrs {
        ipnet, ok := addr.(*net.IPNet)
        if !ok {
            continue
        }
        if ipnet.IP.IsGlobalUnicast() && ipnet.IP.To4() != nil {
			serverURL = ipnet.IP.String()
            return ipnet.IP.String(), nil
        }
    }
    return "", fmt.Errorf("unable to determine IP address")
}


func insertLogMarkup(Text string) { // Inserts to log
    
    iter := logBuffer.GetEndIter()
    
    currentTime := time.Now()
    formattedTime := "<span foreground=\"" + lightBlue + "\">" + currentTime.Format("2006-01-02 15:04:05") + "</span>"

    markup := formattedTime + " " + indicator + " " + Text

    logBuffer.InsertMarkup(iter, markup)

}

func insertLogText(Text string) { // Inserts to log 
        iter := logBuffer.GetEndIter()
        logBuffer.Insert(iter, Text)

}

// Helper functions to convert uint32 to bytes and vice versa
func uint32ToBytes(n uint32) []byte {
    b := make([]byte, 4)
    b[0] = byte(n >> 24)
    b[1] = byte(n >> 16)
    b[2] = byte(n >> 8)
    b[3] = byte(n)
    return b
}

func bytesToUint32(b []byte) uint32 {
    return uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
}
