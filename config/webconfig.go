package config

// Работа с конфигурацией, управляемой через веб-интерфейс

import (
	"encoding/json"
	"errors"
	"github.com/pelletier/go-toml/v2"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type (
	Duration struct {
		D time.Duration
	}

	ErrorPolicy struct {
		DNS     int `toml:"dns" default:"-1"`
		Connect int `toml:"connect" default:"-1"`
		HTTPS   int `toml:"https" default:"-1"`
		HTTP    int `toml:"http" default:"-1"`
		Unknown int `toml:"unknown" default:"-1"`
	}

	WorkerConfig struct {
		Login      string      `toml:"login"`
		Password   string      `toml:"password"`
		Threads    int         `toml:"threads" default:"-1"`
		Timeout    Duration    `toml:"timeout"`
		JobLen     int         `toml:"job_len" default:"-1"`
		ErrorDelay Duration    `toml:"error_delay"`
		ErrorRetry ErrorPolicy `toml:"error_retry"`
	}

	WebConfig struct {
		DomainResendCount     int                     `toml:"DomainResendCount"`
		WorkerPollInterval    Duration                `toml:"WorkerPollInterval"`
		WorkerRetryCount      int                     `toml:"WorkerRetryCount"`
		WorkerRecheckInterval Duration                `toml:"WorkerRecheckInterval"`
		WorkerDefaults        WorkerConfig            `toml:"WorkerDefaults"`
		Workers               map[string]WorkerConfig `toml:"Workers"`
	}
)

var (
	MuxWCfg sync.RWMutex
	WCfg    WebConfig

	DefaultWebConfig = WebConfig{
		DomainResendCount:     1,
		WorkerPollInterval:    Duration{D: 10 * time.Second},
		WorkerRetryCount:      3,
		WorkerRecheckInterval: Duration{D: 2 * time.Minute},
		WorkerDefaults: WorkerConfig{
			Login:      "",
			Password:   "",
			Threads:    2,
			Timeout:    Duration{D: 10 * time.Second},
			JobLen:     100,
			ErrorDelay: Duration{D: 500 * time.Millisecond},
			ErrorRetry: ErrorPolicy{
				DNS:     1,
				Connect: 1,
				HTTPS:   1,
				HTTP:    1,
				Unknown: 1,
			},
		},
	}
)

func (this *Duration) UnmarshalTOML(x interface{}) (err error) {
	switch xx := x.(type) {
	case string:
		this.D, err = time.ParseDuration(xx)
	case int64:
		this.D = time.Duration(xx)
		err = nil
	default:
		err = errors.New("invalid time interval")
	}
	return
}

func (this Duration) MarshalTOML() ([]byte, error) {
	return []byte("\"" + this.D.String() + "\""), nil
}

func (this Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(this.D.String())
}

func LoadWebConfig() error {
	filename := filepath.Join(Cfg.DataDir, "config.toml")
	file, err := os.OpenFile(filename, os.O_RDONLY, 0644)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("Config file %s not exists, create default", filename)
			WCfg = DefaultWebConfig
			WCfg.Workers = make(map[string]WorkerConfig)
			return SaveWebConfig()
		}
		return err
	}
	defer file.Close()
	dec := toml.NewDecoder(file)
	err = dec.Decode(&WCfg)
	return err
}

func SaveWebConfig() error {
	filename := filepath.Join(Cfg.DataDir, "config.toml")
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	enc := toml.NewEncoder(file)
	err = enc.Encode(&WCfg)
	return err
}

func setDefaultString(dst *string, val string) {
	if *dst == "" {
		*dst = val
	}
}

func setDefaultInt(dst *int, val int) {
	if *dst < 0 {
		*dst = val
	}
}

func setDefaultDuration(dst *Duration, val time.Duration) {
	if dst.D == 0 {
		dst.D = val
	}
}
