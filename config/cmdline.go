package config

// Обработка параметров командной строки

import (
	"flag"
)

type (
	CmdOptions struct {
		ConfigFile *string
		ConfigTest *bool
	}
)

var (
	Opts = CmdOptions{
		ConfigFile: flag.String("config", "config.toml", "configuration file"),
		ConfigTest: flag.Bool("cfgtest", false, "test configuration file and exit"),
	}
)

func ParseCmdLine() {
	flag.Parse()
}
