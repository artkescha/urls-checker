package config

// Обработка параметров командной строки

import (
	"flag"
)

type (
	CmdOptions struct {
		ConfigFile         *string
		ConfigTest         *bool
		DomainsArchPath    *string
		SubDomainsArchPath *string
		LinksArchPath      *string
		Unzip              *bool
	}
)

var (
	Opts = CmdOptions{
		ConfigFile:         flag.String("config", "config.toml", "configuration file"),
		ConfigTest:         flag.Bool("cfgtest", false, "test configuration file and exit"),
		DomainsArchPath:    flag.String("domains_path", "./domains.zip", "path to the domains archive"),
		SubDomainsArchPath: flag.String("subdomains_path", "./subdomains.zip", "path to the subdomains archive"),
		LinksArchPath:      flag.String("Links_path", "./links.zip", "path to the links archive"),
		Unzip:              flag.Bool("unzip", false, "unzip .zip archives of domains, subdomains and links"),
	}
)

func ParseCmdLine() {
	flag.Parse()
}
