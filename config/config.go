package config

// Обработка файла конфигурации

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

type (
	Config struct {
		ListenPort string `toml:"ListenPort"`
		Username   string `toml:"Username"`
		Password   string `toml:"Password"`
		DataDir    string `toml:"DataDir"`
		DebugLog   string `toml:"DebugLog"`
	}
)

var (
	Cfg Config
)

func LoadConfig(cfgfile string) error {
	file, err := os.OpenFile(cfgfile, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	dec := toml.NewDecoder(file)
	err = dec.Decode(&Cfg)
	return err
}

func DumpConfig() string {
	b, _ := json.MarshalIndent(Cfg, "", "\t")
	return string(b)
}

func OpenLog() (err error) {
	var logfile *os.File
	if Cfg.DebugLog == "" {
		logfile, err = os.OpenFile(os.DevNull, os.O_WRONLY, 0644)
	} else {
		filename := filepath.Join(Cfg.DataDir, Cfg.DebugLog)
		logfile, err = os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	}
	if err != nil {
		return
	}
	log.SetFlags(log.LstdFlags)
	log.SetOutput(logfile)
	return nil
}
