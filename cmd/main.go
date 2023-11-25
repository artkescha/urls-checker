package main

import (
	"fmt"
	"log"
	"urls-checker/internal/urls_checker"
	"urls-checker/internal/urls_creator"
	"urls-checker/internal/writer"

	"urls-checker/config"
	"urls-checker/internal/webui"
)

func main() {
	// Обрабатываем параметры командной строки и загружаем конфигурацию
	config.ParseCmdLine()
	err := config.LoadConfig(*config.Opts.ConfigFile)
	if err != nil {
		fmt.Println("Config error: ", err)
		return
	}
	if *config.Opts.ConfigTest {
		fmt.Println(config.DumpConfig())
		return
	}

	err = config.OpenLog()
	if err != nil {
		fmt.Println("Open log error: ", err)
		return
	}

	err = config.LoadWebConfig()
	if err != nil {
		fmt.Println("Web config error: ", err)
		return
	}

	creator := urls_creator.New(urls_creator.Config{})
	urlsChan := creator.Start()

	checker := urls_checker.New(urlsChan)
	sucUrlsChan := checker.Start(100)

	go writer.Writer("./", "sucsess.txt", sucUrlsChan)

	log.Fatal(webui.StartListener(config.Cfg.ListenPort, config.Cfg.Username, config.Cfg.Password))
}
