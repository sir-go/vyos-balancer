package main

import (
	"flag"
	"os"
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

func ConfigInit() *Config {
	fCfgPath := flag.String("c", "config.toml", "path to conf file")
	flag.Parse()

	conf := new(Config)
	file, err := os.Open(*fCfgPath)
	if err != nil {
		panic(err)
	}

	defer func() {
		if file == nil {
			return
		}
		if err = file.Close(); err != nil {
			panic(err)
		}
	}()

	if _, err = toml.DecodeFile(*fCfgPath, &conf); err != nil {
		panic(err)
	}
	conf.Path = *fCfgPath
	return conf
}
