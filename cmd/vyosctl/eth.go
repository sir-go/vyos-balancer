package main

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"time"
)

type (
	Threshold struct {
		Limit         int
		OverflowCount uint
		Overflowed    bool
	}

	Uplink struct {
		Alias   string     `toml:"alias"`
		Lz      int        `toml:"lz"`
		L0      *Threshold `toml:"l0"`
		L1      *Threshold `toml:"l1"`
		Nat     string     `toml:"nat"`
		Current int
		bytes0  uint64
		bytes1  uint64
	}
)

func (t *Threshold) CheckLimit(val int) {
	if val > t.Limit {
		if t.OverflowCount < CFG.OverflowCount {
			t.OverflowCount++
		}
	} else {
		t.OverflowCount = 0
	}
	t.Overflowed = t.OverflowCount >= CFG.OverflowCount
}

func ReadLines(filename string) []string {
	f, err := os.Open(filename)
	eh(err)
	defer func() { eh(f.Close()) }()
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

func GetByAlias(ups []*Uplink, alias string) *Uplink {
	for _, u := range ups {
		if u.Alias == alias {
			return u
		}
	}
	return nil
}

func GetIntefacesInfo(ups []*Uplink) {
	lines := ReadLines("/proc/net/dev")

	for _, line := range lines {
		fields := strings.Split(line, ":")
		if len(fields) < 2 {
			continue
		}
		key := strings.TrimSpace(fields[0])
		value := strings.Fields(strings.TrimSpace(fields[1]))

		u := GetByAlias(ups, key)
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

func GetDevBW(ups ...*Uplink) {
	ZeroBytes(ups...)
	GetIntefacesInfo(ups)
	time.Sleep(time.Second)
	GetIntefacesInfo(ups)

	for _, u := range ups {
		u.Current = int((u.bytes1 - u.bytes0) * 8 / 1e6)
	}
}

func (u *Uplink) IsDown() bool {
	return u.Current < u.Lz
}

func CalcLimits(ups ...*Uplink) {
	for _, u := range ups {
		u.L0.CheckLimit(u.Current)
		u.L1.CheckLimit(u.Current)
	}
}

func ZeroBytes(ups ...*Uplink) {
	for _, u := range ups {
		u.bytes0 = 0
		u.bytes1 = 0
	}
}

func ZeroCounts(ups ...*Uplink) {
	for _, u := range ups {
		u.L0.Overflowed = false
		u.L0.OverflowCount = 0
		u.L1.Overflowed = false
		u.L1.OverflowCount = 0
	}
}
