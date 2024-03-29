package DB 

import (
	"fmt"
	"log"
	"strings"
	"math/rand"
	"time"
	"database/sql"
    "os"
    "path/filepath"
	"Team-Server/UI"
	//"Team-Server/Server"
	_ "github.com/mattn/go-sqlite3"
)

var DbDir string
var DbPath string
var NewConnections []ConnectionLog
type ConnectionLog struct {
    Time string
    HostVersion string
    ID          int
}

var green = "#7CFC00"

func InitDB() {
	db, err := buildDB()
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
}

func buildDB() (*sql.DB, error) {
    // Get the current working directory
    wd, err := os.Getwd()
    if err != nil {
        log.Fatal(err)
    }

    // Create the database directory if it doesn't exist
    dbDir := filepath.Join(wd, "DB")
    if _, err := os.Stat(dbDir); os.IsNotExist(err) {
        err := os.MkdirAll(dbDir, 0755)
        if err != nil {
            log.Fatal(err)
        }
    }

    // Open SQLite database file
    DbPath = filepath.Join(DbDir, "siafu.db")
    db, err := sql.Open("sqlite3", DbPath)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Create table if it does not exist
    _, err = db.Exec(`CREATE TABLE IF NOT EXISTS implants (
        uid TEXT PRIMARY KEY,
        versionName TEXT,
        IDMask INTEGER
    )`)
    if err != nil {
    log.Fatal(err)
    }

    return db, nil
}


func ParseUID(uid string) (string, string) {
    // Split the UID into its parts
    parts := strings.Split(uid, "-")
    if len(parts) != 2 {
        return "", ""
    }
    return parts[0], parts[1]
}

func IsUIDInDB(uid, versionName string) (int, error) {
    // Open database connection
    db, err := sql.Open("sqlite3", DbPath)
    if err != nil {
        return 0, err
    }
    defer db.Close()

    var IDMask int
    // Query the database to check if UID exists
    err = db.QueryRow("SELECT IDMask FROM implants WHERE uid = ? AND versionName = ?", uid, versionName).Scan(&IDMask)
    switch {
    case err == sql.ErrNoRows:
        // UID not found, add it to the database
        IDMask, err := addToDB(uid, versionName)
        if err != nil {
            return 0, err
        }
        return IDMask, nil
    case err != nil:
        return 0, err // error occurred
    default:
        return IDMask, nil // UID found
    }
}

func generateIDMask() (int, error) {
    // Generate a random ID mask
    return rand.Intn(1000), nil
}

func addToDB(uid, versionName string) (int, error) {
    // Generate IDMask
    IDMask, err := generateIDMask()
    if err != nil {
        return 0, err
    }

    // Open database connection
    db, err := sql.Open("sqlite3", DbPath)
    if err != nil {
        return 0, err
    }
    defer db.Close()

    // Insert values into database
    _, err = db.Exec("INSERT INTO implants (uid, versionName, IDMask) VALUES (?, ?, ?)", uid, versionName, IDMask)
    if err != nil {
        return 0, err
    }

    fmt.Print(green)
    fmt.Println("New Connection")
    fmt.Println("Host Version:", versionName)
    fmt.Println("ID:", IDMask)

    currentTime := time.Now()
    timeString := currentTime.Format("2006-01-02 15:04:05")

    NewConnections = append(NewConnections, ConnectionLog{
        Time: timeString,
        HostVersion: versionName,
        ID:          IDMask,
    })

    fmt.Print(UI.Reset)
    return IDMask, nil
}

