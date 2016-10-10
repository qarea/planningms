package cfg

import (
	"strings"
	"time"

	"github.com/powerman/narada-go/narada"
)

var log = narada.NewLog("")

var (
	// Debug enable debug logs
	Debug bool

	// LockTimeout for narada.SharedLock
	LockTimeout time.Duration

	// RSAPublicKey for JWT token verification
	RSAPublicKey []byte

	// MySQL configuration
	MySQL struct {
		Host     string
		Port     int
		DB       string
		Login    string
		Password string
	}

	// HTTP configuration for application http server
	HTTP struct {
		Listen       string
		BasePath     string
		RealIPHeader string
	}

	// TimeSpent configuration for backup of in-memory data
	TimeSpent struct {
		Folder    string
		Frequency time.Duration
	}

	// Plannings configuration for plannings types
	Plannings struct {
		MaxAge           time.Duration
		OldestLastUpdate time.Duration
	}
)

func init() {
	if err := load(); err != nil {
		log.Fatal(err)
	}
}

func load() error {
	Debug = narada.GetConfigLine("log/level") == "DEBUG"

	HTTP.Listen = narada.GetConfigLine("http/listen")
	if strings.Index(HTTP.Listen, ":") == -1 {
		log.Fatal("please setup config/http/listen")
	}

	HTTP.BasePath = narada.GetConfigLine("http/basepath")
	if HTTP.BasePath != "" && (HTTP.BasePath[0] != '/' || HTTP.BasePath[len(HTTP.BasePath)-1] == '/') {
		log.Fatal("config/http/basepath should begin with / and should not end with /")
	}

	HTTP.RealIPHeader = narada.GetConfigLine("http/real_ip_header")

	MySQL.Host = narada.GetConfigLine("mysql/host")
	MySQL.Port = narada.GetConfigInt("mysql/port")
	MySQL.DB = narada.GetConfigLine("mysql/db")
	MySQL.Login = narada.GetConfigLine("mysql/login")
	MySQL.Password = narada.GetConfigLine("mysql/pass")

	var err error
	RSAPublicKey, err = narada.GetConfig("rsa_public_key")
	if err != nil {
		return err
	}

	TimeSpent.Folder = narada.GetConfigLine("timespent/backup/folder")
	if TimeSpent.Folder == "" {
		log.Fatal("Please setup backup folder timespent/backup/folder")
	}
	TimeSpent.Frequency = narada.GetConfigDuration("timespent/backup/frequency")

	Plannings.MaxAge = narada.GetConfigDuration("plannings/max_age")
	Plannings.OldestLastUpdate = narada.GetConfigDuration("plannings/oldest_last_update")

	LockTimeout = narada.GetConfigDuration("lock_timeout")
	return nil
}
