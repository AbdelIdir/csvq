package cmd

import (
	"math/rand"
	"sync"
	"time"
)

var (
	random  *rand.Rand
	getRand sync.Once

	location    *time.Location
	getLocation sync.Once
)

func GetRand() *rand.Rand {
	getRand.Do(func() {
		random = rand.New(rand.NewSource(time.Now().UnixNano()))
	})
	return random
}

func GetLocation() *time.Location {
	getLocation.Do(func() {
		location, _ = time.LoadLocation(GetFlags().Location)
	})
	return location
}

func Now() time.Time {
	if 0 < len(GetFlags().Now) {
		t, _ := time.ParseInLocation("2006-01-02 15:04:05.999999999", GetFlags().Now, GetLocation())
		return t
	}
	GetLocation()
	return time.Now()
}
