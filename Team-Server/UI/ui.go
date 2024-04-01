package UI

import (
	"github.com/fstanis/screenresolution"
	"github.com/gotk3/gotk3/gtk"
    "github.com/gotk3/gotk3/gdk"
	"time"
)


var logBuffer *gtk.TextBuffer
var mRefProvider *gtk.CssProvider
var screen *gdk.Screen
var ScreenWidth int 
var ScreenHeight int 
var win *gtk.Window 

const (
    lightBlue = "#90afc5" /* selected items & other interactive components */
    brightBlue = "#1e90ff" /* use for notifications */
    brightRed = "#e14e19" /* use for error notifications */
    allWhite = "#ffffff" /* text */
	Reset = "#000000"
)

var indicator = "<span>[" + "<span foreground=\"" + lightBlue + "\">+</span>" + "]</span>"


func InitUI() {
    gtk.Init(nil)
    
    win, _ = gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
    win.SetTitle("Siafu Team Server")
    win.Connect("destroy", func() {
        gtk.MainQuit()
    })

    // Load CSS stylesheet
    mRefProvider, _  = gtk.CssProviderNew()
    mRefProvider.LoadFromPath("./UI/Material-DeepOcean/gtk-dark.css")
 
    // Apply to whole app
    screen, _ = gdk.ScreenGetDefault()
    gtk.AddProviderForScreen(screen, mRefProvider, gtk.STYLE_PROVIDER_PRIORITY_USER)
    
    // Create a Box for organizing widgets
	box, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)

	// Create menu
	menuBar := createFileMenu()
	box.PackStart(menuBar, false, false, 0)

	// Create a box for holding the text view
	textBox, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
	box.PackStart(textBox, true, true, 0)

	// Create a scrolled window
	scrolledWin, _ := gtk.ScrolledWindowNew(nil, nil)
	scrolledWin.SetPolicy(gtk.POLICY_AUTOMATIC, gtk.POLICY_AUTOMATIC)
	textBox.PackStart(scrolledWin, true, true, 0)

	// Create a text view
	servertxt, _ := gtk.TextViewNew()
	servertxt.SetEditable(false)
	scrolledWin.Add(servertxt)

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

    // Start the GTK main event loop
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

    //submenu.Append(newTabItem)
    submenu.Append(quitItem)
    menu.SetSubmenu(submenu)

    return menu
}


func InsertLogMarkup(Text string) { // Inserts to log

    iter := logBuffer.GetEndIter()
    
    currentTime := time.Now()
    formattedTime := "<span foreground=\"" + lightBlue + "\">" + currentTime.Format("2006-01-02 15:04:05") + "</span>"

    markup := formattedTime + " " + indicator + " " + Text + "\n"

    logBuffer.InsertMarkup(iter, markup)

}

func InsertLogText(Text string) { // Inserts to log 
	    iter := logBuffer.GetEndIter()
        logBuffer.Insert(iter, Text)

}