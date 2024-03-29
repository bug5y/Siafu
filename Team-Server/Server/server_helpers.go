package Server

import (
	"database/sql"
	"strconv"
	"fmt"
	"net"
	"Team-Server/DB"
)
var serverIP string

func updateConnections() []string {
    // New connections are added to connectionlog in addToDB()
    serverresp = nil
    // If changes were made get the changes
    if len(DB.NewConnections) != prevLength {
        // If changes were made, format the log
        var formattedLog string
        for _, conn := range DB.NewConnections {
            formattedLog += fmt.Sprintf("Time: %s, Host Version: %s, ID: %d\n", conn.Time, conn.HostVersion, conn.ID)
        }
        // Append the formatted log to serverresp
        serverresp = append(serverresp, formattedLog)
        prevLength = len(DB.NewConnections)
        return serverresp
    }
    return serverresp // No changes
}

func getVersionNamesAndUIDs() ([]string, []string, []string, error) {
    var idMasks []string
    var versionNames []string
    var uids []string

    // Collect all ID masks from the ImplantMap
    for _, idMask := range ImplantMap {
        idMasks = append(idMasks, strconv.Itoa(idMask))
    }

    // Open a connection to the database
    db, err := sql.Open("sqlite3", DB.DbPath)
    if err != nil {
        return nil, nil, nil, fmt.Errorf("failed to open database connection: %s", err)
    }
    defer db.Close()

    // Query the database for versionName and uid associated with each ID mask
    for _, idMask := range idMasks {
        var versionName string
        var uid string

        // Execute the query
        row := db.QueryRow("SELECT versionName, uid FROM implants WHERE IDMask = ?", idMask)

        // Scan the query result into variables
        if err := row.Scan(&versionName, &uid); err != nil {
            if err == sql.ErrNoRows {
                // If no rows are found for the given ID mask, continue to the next one
                continue
            }
            return nil, nil, nil, fmt.Errorf("failed to scan row: %s", err)
        }

        // Append versionName and uid to the slices
        versionNames = append(versionNames, versionName)
        uids = append(uids, uid)
    }

    return idMasks, versionNames, uids, nil
}

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

func containsIDMask(idMask int) bool {
    for _, item := range commandQueue {
        if item.IDMask == idMask {
            return true
        }
    }
    return false
}

func addToCommandQueue(commands [][]string, idMask int) {
    item := CommandQueueItem{
        Commands: commands,
        IDMask:   idMask,
    }
    commandQueue = append(commandQueue, item)
}

func handleUID(uid string) {
	randUID, versionstr := DB.ParseUID(uid)
	if randUID == "" || versionstr == "" {
		fmt.Println("Invalid UID format")
		return
	}
	
    value, err := strconv.ParseUint(versionstr, 10, 32)
    if err != nil {
        fmt.Println("Error:", err)
        return
    }

    uintValue := uint32(value)
    hexString := fmt.Sprintf("0x%04X", uintValue)

    versionName := getVersionString(hexString)

    IDMask, err := DB.IsUIDInDB(uid, versionName)
	if err != nil {
		fmt.Println("Error checking UID:", err)
		return
	}

    // Update the map with the uid and implant-id mapping
    ImplantMap[uid] = IDMask
}

func getVersionString(versionStr string) string {
    version, err := strconv.ParseUint(versionStr[2:], 16, 32) // Remove "0x" prefix
    if err != nil {
        return "Invalid version string"
    }

    switch uint32(version) {
    case 0x0500:
        return "Windows 2000"
    case 0x0501:
        return "Windows XP"
    case 0x0502:
        return "Windows XP Professional"
    case 0x0600:
        return "Windows Vista"
    case 0x0601:
        return "Windows 7"
    case 0x0602:
        return "Windows 8"
    case 0x0603:
        return "Windows 8.1"
    case 0x0A00:
        return "Windows 10"
    case 0x0B00:
        return "Windows 11"
    case 0x0503:
        return "Windows Server 2003 or Windows Home Server"
    case 0x0604:
        return "Windows Server 2008"
    case 0x0605:
        return "Windows Server 2008 R2"
    case 0x0606:
        return "Windows Server 2012"
    case 0x0607:
        return "Windows Server 2012 R2"
    case 0x0A08:
        return "Windows Server 2016"
    case 0x0A09:
        return "Windows Server 2019"
    default:
        return "Unknown"
    }
}

func getServerIP() (string, error) {
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
			serverIP = ipnet.IP.String()
            return ipnet.IP.String(), nil
        }
    }
    return "", fmt.Errorf("unable to determine server IP address")
}
