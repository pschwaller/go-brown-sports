package pkg

import (
	"fmt"
	"reflect"
	"time"
)

type SportingEvent struct {
	Datetime time.Time
	Sport    string
	Emails   []string
	Roles    []string // Text representation
}

func (event SportingEvent) Format() string {
	text := ""
	text += fmt.Sprintf("When: %s\nSport: %s\n", event.Datetime, event.Sport)

	for _, role := range event.Roles {
		text = text + role + "\n"
	}

	return text
}

func (event SportingEvent) GetKey() string {
	return fmt.Sprintf("%s @ %s", event.Datetime, event.Sport)
}

func (event SportingEvent) IsMostlyEqual(event2 SportingEvent) bool {
	// TODO using reflect() is _probably_ unnecessary here.  Need to test using == instead.
	return event.Datetime.Equal(event2.Datetime) &&
		event.Sport == event2.Sport &&
		reflect.DeepEqual(event.Roles, event2.Roles)
}

func GetKeysFromSportingEventMap(myMap map[string]SportingEvent) []interface{} {
	keys := make([]interface{}, 0, len(myMap))
	for k := range myMap {
		keys = append(keys, k)
	}

	return keys
}
