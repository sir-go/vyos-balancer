package main

import (
	"bufio"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

type (
	// Threshold contains info about Uplink limits
	Threshold struct {
		Limit         int  // bandwidth limit
		OverflowCount uint // amount of overflowing cases
		Overflowed    bool // is limit overflowed now?
	}

	// Uplink contains info about Uplink interface
	Uplink struct {
		Alias   string     `toml:"alias"` // network interface name
		Lz      int        `toml:"lz"`    // zero-level (if rates lower, interface is down)
		L0      *Threshold `toml:"l0"`    // lower limit for traffic consumption rate
		L1      *Threshold `toml:"l1"`    // upper limit for traffic consumption rate
		Nat     string     `toml:"nat"`   // NAT subnet
		Current int        // current traffic consumption rate
		bytes0  uint64     // previous rate value
		bytes1  uint64     // current rate value
	}
)

// CheckLimit decodes if limit was overflowed by val
func (t *Threshold) CheckLimit(val int, ovfCount uint) {
	if val > t.Limit {
		if t.OverflowCount < ovfCount {
			t.OverflowCount++
		}
	} else {
		t.OverflowCount = 0
	}
	t.Overflowed = t.OverflowCount > ovfCount
}

// ReadLines reads entire file as array of strings
func ReadLines(filename string) []string {
	f, err := os.Open(path.Clean(filename))
	if err != nil {
		LOG.Panicln("read", filename, err)
	}

	defer func() {
		if err := f.Close(); err != nil {
			LOG.Println("close", filename, err)
		}
	}()
	var ret []string
	r := bufio.NewReader(f)
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			break
		}
		ret = append(ret, strings.Trim(line, "\n"))
	}
	return ret
}

// GetByAlias find Uplink in the array of Uplinks by it's alias
func GetByAlias(alias string, ups ...*Uplink) *Uplink {
	for _, u := range ups {
		if u.Alias == alias {
			return u
		}
	}
	return nil
}

// GetInterfacesInfo gets network interfaces rates and updates theirs info in the array of Uplink
func GetInterfacesInfo(ups ...*Uplink) {
	lines := ReadLines("/proc/net/dev")

	for _, line := range lines {
		fields := strings.Split(line, ":")
		if len(fields) < 2 {
			continue
		}
		key := strings.TrimSpace(fields[0])
		value := strings.Fields(strings.TrimSpace(fields[1]))

		u := GetByAlias(key, ups...)
		if u == nil {
			continue
		}

		r, err := strconv.ParseUint(value[0], 10, 64)
		if err != nil {
			LOG.Println(key, value[0], err)
			break
		}

		u.bytes0 = u.bytes1
		u.bytes1 = r
	}
}

// GetDevBW calculates bandwidth for each Uplink in arguments and updates it's field
func GetDevBW(ups ...*Uplink) {
	ZeroBytes(ups...)
	GetInterfacesInfo(ups...)
	time.Sleep(time.Second)
	GetInterfacesInfo(ups...)

	for _, u := range ups {
		u.Current = int((u.bytes1 - u.bytes0) * 8 / 1e6)
	}
}

// IsDown shows is the Uplinks' current rate lower than zero-level
func (u *Uplink) IsDown() bool {
	return u.Current < u.Lz
}

// CalcLimits updates lower and upper limits for each Uplink and updates their overflow counters
func CalcLimits(overflowCount uint, ups ...*Uplink) {
	for _, u := range ups {
		u.L0.CheckLimit(u.Current, overflowCount)
		u.L1.CheckLimit(u.Current, overflowCount)
	}
}

// ZeroBytes resets byte counters for each Uplink
func ZeroBytes(ups ...*Uplink) {
	for _, u := range ups {
		u.bytes0 = 0
		u.bytes1 = 0
	}
}

// ZeroCounts resets calculated limits for each Uplink
func ZeroCounts(ups ...*Uplink) {
	for _, u := range ups {
		u.L0.Overflowed = false
		u.L0.OverflowCount = 0
		u.L1.Overflowed = false
		u.L1.OverflowCount = 0
	}
}
