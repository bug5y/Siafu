package UI

import (
	"github.com/fstanis/screenresolution"
	"github.com/gotk3/gotk3/gtk"
    "github.com/gotk3/gotk3/gdk"
	"time"
    "sync"
    "log"
)

var logBuffer *gtk.TextBuffer
var mRefProvider *gtk.CssProvider
var screen *gdk.Screen
var ScreenWidth int 
var ScreenHeight int 
var win *gtk.Window 
var logMutex sync.Mutex
var bufferMu sync.Mutex 

const (
    lightBlue = "#90afc5"
    brightBlue = "#1e90ff" /* use for non-error notifications */
    brightRed = "#e14e19" /* use for error notifications */
    allWhite = "#ffffff" /* text */
	Reset = "#000000"
)

var indicator = "<span>[" + "<span foreground=\"" + lightBlue + "\">+</span>" + "]</span>"


func InitUI() {
    gtk.Init(nil)
    
    settings, _ := gtk.SettingsGetDefault()
    settings.SetProperty("gtk-theme-name", "Adwaita-dark")
    
    win, _ = gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
    win.SetTitle("Siafu Team Server")
    win.Connect("destroy", func() {
        gtk.MainQuit()
    })

    mRefProvider, _  = gtk.CssProviderNew()
    //cssPath := SiafuBase + "/UI/Material-DeepOcean/gtk-dark.css"
    //fmt.Println(cssPath)
    mRefProvider.LoadFromPath("./UI/Material-DeepOcean/gtk-dark.css")
 
    // Apply to whole app
    screen, _ = gdk.ScreenGetDefault()
    gtk.AddProviderForScreen(screen, mRefProvider, gtk.STYLE_PROVIDER_PRIORITY_USER)
	box, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
    menubar, _ := gtk.MenuBarNew()
	fileMenu := createFileMenu()
    menubar.Append(fileMenu)

	box.PackStart(menubar, false, false, 0)

	notebook, err := gtk.NotebookNew()
	if err != nil {
		log.Fatal("Unable to create notebook:", err)
	}

	textBox, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
	textBox.SetHExpand(true)
	textBox.SetVExpand(true)

	scrolledWin, _ := gtk.ScrolledWindowNew(nil, nil)
	scrolledWin.SetPolicy(gtk.POLICY_AUTOMATIC, gtk.POLICY_AUTOMATIC)

	servertxt, _ := gtk.TextViewNew()
	servertxt.SetEditable(false)
	scrolledWin.Add(servertxt)

	textBox.PackStart(scrolledWin, true, true, 0)

	page, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
	if err != nil {
		log.Fatal("Unable to create notebook page:", err)
	}
	page.SetHExpand(true)
	page.SetVExpand(true)
	page.Add(textBox)

	label, err := gtk.LabelNew("Actvity Log")
	if err != nil {
		log.Fatal("Unable to create label:", err)
	}
	notebook.AppendPage(page, label)

    box.PackEnd(notebook, true, true, 0)

	win.Add(box)

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
    logBuffer, _ = servertxt.GetBuffer()
}

func BuildUI() {
	win.SetDefaultSize(ScreenWidth, ScreenHeight)
    win.ShowAll()
    gtk.Main()
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


func InsertLogMarkup(Text string) { // Inserts to log
    logMutex.Lock()
    bufferMu.Lock()
    
    iter := logBuffer.GetEndIter()
    currentTime := time.Now()
    formattedTime := "<span foreground=\"" + lightBlue + "\">" + "  " + currentTime.Format("2006-01-02 15:04:05") + "</span>"
    markup := formattedTime + " " + indicator + " " + Text + "\n"
    logBuffer.InsertMarkup(iter, markup)

    bufferMu.Unlock()
    logMutex.Unlock()

}
