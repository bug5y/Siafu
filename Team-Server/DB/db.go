package DB 

import (
    //"fmt"
	"log"
	"strings"
	"time"
	"database/sql"
    "os"
    "path/filepath"
    "encoding/base64"
    "crypto/sha256"
    "encoding/hex"
	//"Team-Server/UI"
	//"Team-Server/Server"
	_ "github.com/mattn/go-sqlite3"
)
//var mutex sync.RWMutex
var DbDir string
var DbPath string

type Connections map[string]ConnectionDetails

type ConnectionDetails struct {
	HostVersion string `json:"HostVersion"`
	AgentType   string `json:"AgentType"`
	ImplantID   string `json:"ImplantID"`
	InternalIP  string `json:"InternalIP"`
	ExternalIP  string `json:"ExternalIP"`
	User        string `json:"User"`
	HostName    string `json:"HostName"`
	LastSeen    string `json:"LastSeen"`
	FullHash    string `json:"FullHash"`
	OrgUID      string `json:"OrgUID"`
}

var ConnectionLog = Connections{}
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
        IDMask TEXT
    )`)
    if err != nil {
    log.Fatal(err)
    }

    return db, nil
}
    //IPs := strings.Split(string(parts[4]), ",")
func ParseUID(encodedUID string, protocol string) (ConnectionDetails, string, error) {
    uid, _ := base64.StdEncoding.DecodeString(encodedUID)
    parts := strings.Split(string(uid), "-")

    currentTime := time.Now()
    timeString := currentTime.Format("2006-01-02 15:04:05")

    hash := sha256.New()
    hash.Write([]byte(string(uid)))
    hashSum := hash.Sum(nil)
    hexString := hex.EncodeToString(hashSum)
    truncatedHash := hexString[:8]

    uidParts := ConnectionDetails{
        HostVersion: parts[3],
        AgentType:   protocol,
        ImplantID:   truncatedHash,
        InternalIP:  parts[4], // Assuming first IP is Internal
        ExternalIP:  parts[4], // Assuming first IP is External
        User:        parts[2],
        HostName:    parts[1],
        LastSeen:    timeString,
        FullHash:    hexString,
        OrgUID:      encodedUID,
    }

	if ConnectionLog == nil {
		ConnectionLog = make(Connections)
	}
    //mutex.Lock()
	//defer mutex.Unlock()
    ConnectionLog[hexString] = uidParts

    return uidParts, hexString, nil
}

func IsUIDInDB(FullHash, versionName string) (error) {
    var truncatedHash string    // Open database connection
    db, err := sql.Open("sqlite3", DbPath)
    if err != nil {
        return err
    }
    defer db.Close()

    // Query the database to check if UID exists
    err = db.QueryRow("SELECT IDMask FROM implants WHERE uid = ? AND versionName = ?", FullHash, versionName).Scan(&truncatedHash)
    switch {
    case err == sql.ErrNoRows:
        // UID not found, add it to the database
        err := addToDB(FullHash, versionName, truncatedHash)
        if err != nil {
            return err
        }
        return nil
    case err != nil:
        return err // error occurred
    default:
        return nil // UID found
    }
}

func addToDB(FullHash, versionName, truncatedHash string) (error) {
    // Open database connection
    db, err := sql.Open("sqlite3", DbPath)
    if err != nil {
        return err
    }
    defer db.Close()

    // Insert values into database
    _, err = db.Exec("INSERT INTO implants (uid, versionName, IDMask) VALUES (?, ?, ?)", FullHash, versionName, truncatedHash)
    if err != nil {
        return err
    }
    return err
}

