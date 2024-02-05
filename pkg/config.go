package pkg

import (
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"sync"
)

func GetSpreadsheetID() string {
	return getInstance().SpreadsheetID
}

func GetCalendarID() string {
	return getInstance().CalendarID
}

// TODO can this be used without a global variable?
var lock = &sync.Mutex{}

type Configuration struct {
	CalendarID    string `envconfig:"CALENDAR_ID"    yaml:"calendarId"`
	SpreadsheetID string `envconfig:"SPREADSHEET_ID" yaml:"spreadsheetId"`
}

func newConfiguration() *Configuration {
	var cfg Configuration

	readConfig(&cfg)
	readEnv(&cfg)

	return &cfg
}

func readConfig(cfg *Configuration) {
	filename := "config.yaml"
	fileHandle, err := os.Open(filename)

	if err != nil {
		log.Printf("Could not load %s file: %v", filename, err)
		// Return so we don't report on a decoding error for a file we couldn't load.
		return
	}
	defer fileHandle.Close()

	decoder := yaml.NewDecoder(fileHandle)
	err = decoder.Decode(cfg)

	if err != nil {
		log.Printf("Error decoding %s: %v", filename, err)
	}
}

func readEnv(cfg *Configuration) {
	err := envconfig.Process("", cfg)
	if err != nil {
		log.Printf("Error reading environment variables: %v", err)
	}
}

var singleInstance *Configuration

func getInstance() *Configuration {
	if singleInstance == nil {
		lock.Lock()
		defer lock.Unlock()

		if singleInstance == nil {
			singleInstance = newConfiguration()
		}
	}

	return singleInstance
}
