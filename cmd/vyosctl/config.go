package main

import (
	"strconv"
	"time"

	"github.com/BurntSushi/toml"
)

type (
	Duration struct {
		time.Duration
	}

	CfgVyos struct {
		Addr string `toml:"addr"`
		Key  string `toml:"key"`
	}

	Config struct {
		Vyos          CfgVyos   `toml:"vyos"`
		Beeline       *Uplink   `toml:"beeline"`
		TTK           *Uplink   `toml:"ttk"`
		OverflowCount uint      `toml:"overflow_count"`
		CheckPeriod   *Duration `toml:"check_period"`
		Path          string
	}
)

func (d *Duration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	return err
}

func (t *Threshold) UnmarshalText(text []byte) error {
	var err error
	t.Limit, err = strconv.Atoi(string(text))
	return err
}

func LoadConfig(fileName string) (conf *Config) {
	if _, err := toml.DecodeFile(fileName, &conf); err != nil {
		panic(err)
	}
	conf.Path = fileName
	return
}
