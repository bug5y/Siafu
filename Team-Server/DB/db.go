package DB

import (
	"crypto/sha256"
	"encoding/hex"
	"net"
	"strings"
	"sync"
	"time"
)

var ConnectionMutex sync.Mutex
var DbDir string
var DbPath string

type Connections map[string]ConnectionDetails

type ConnectionDetails struct {
	// randstr
	HostVersion string `json:"HostVersion"`
	HostName    string `json:"HostName"`
	User        string `json:"User"`
	AgentType   string `json:"AgentType"`
	ImplantID   string `json:"ImplantID"`

	InternalIP string `json:"InternalIP"`
	ExternalIP string `json:"ExternalIP"`

	LastSeen string `json:"LastSeen"`
	FullHash string `json:"FullHash"`
}

var ConnectionLog = Connections{}

func UpdateLastSeen(key string) {
	ConnectionMutex.Lock()
	defer ConnectionMutex.Unlock()

	// search by key then update LastSeen
	if implant, ok := ConnectionLog[key]; ok {
		currentTime := time.Now()
		timeString := currentTime.Format("2006-01-02 15:04:05")
		implant.LastSeen = timeString

		ConnectionLog[key] = implant
	}
}

func ParseUID(parts []string, protocol string, uid string, base64UID string) (ConnectionDetails, string, error) {
	ConnectionMutex.Lock()
	defer ConnectionMutex.Unlock()

	currentTime := time.Now()
	timeString := currentTime.Format("2006-01-02 15:04:05")

	hash := sha256.New()
	hash.Write([]byte(string(base64UID)))
	hashSum := hash.Sum(nil)
	hexString := hex.EncodeToString(hashSum)
	truncatedHash := hexString[:8]

	VerStr := parts[3]

	ExternalIPs, InternalIPs := sortIPs(parts[4])

	uidParts := ConnectionDetails{
		HostVersion: VerStr,
		AgentType:   protocol,
		ImplantID:   truncatedHash,
		InternalIP:  InternalIPs,
		ExternalIP:  ExternalIPs,
		User:        parts[2],
		HostName:    parts[1],
		LastSeen:    timeString,
		FullHash:    hexString,
	}

	if ConnectionLog == nil {
		ConnectionLog = make(Connections)
	}

	ConnectionLog[truncatedHash] = uidParts

	return uidParts, hexString, nil
}

func sortIPs(ipsStr string) (string, string) {
	IPs := strings.Split(ipsStr, ",")

	InternalRanges := []string{"0.0.0.0/8", "169.254.0.0/16", "127.0.0.0/8", "10.0.0.0/8", "100.64.0.0/10", "172.16.0.0/12", "192.0.0.0/24", "198.18.0.0/15", "192.168.0.0/16"}
	var InternalIPs, ExternalIPs []string

	for _, ip := range IPs {
		ip = strings.TrimSpace(ip)
		// Check if the IP address is within any of the internal ranges
		isInternal := false
		for _, internalRange := range InternalRanges {
			if isIPInRange(ip, internalRange) {
				InternalIPs = append(InternalIPs, ip)
				isInternal = true
				break
			}
		}
		if !isInternal {
			ExternalIPs = append(ExternalIPs, ip)
		}
	}

	return strings.Join(ExternalIPs, ", "), strings.Join(InternalIPs, ", ")
}

func isIPInRange(ip, cidr string) bool {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return false
	}

	ipAddr := net.ParseIP(ip)
	return ipNet.Contains(ipAddr)
}
