package main

import (
	"fmt"
	"urls-checker/internal/urls_checker"
	"urls-checker/internal/urls_creator"
	"urls-checker/internal/writer"

	"urls-checker/config"
)

func main() {
	// Обрабатываем параметры командной строки и загружаем конфигурацию
	config.ParseCmdLine()
	//err := config.LoadConfig(*config.Opts.ConfigFile)
	//if err != nil {
	//	fmt.Println("Config error: ", err)
	//	return
	//}
	//if *config.Opts.ConfigTest {
	//	fmt.Println(config.DumpConfig())
	//	return
	//}
	//
	//err = config.OpenLog()
	//if err != nil {
	//	fmt.Println("Open log error: ", err)
	//	return
	//}
	//
	//err = config.LoadWebConfig()
	//if err != nil {
	//	fmt.Println("Web config error: ", err)
	//	return
	//}

	config := urls_creator.Config{
		DomainsFile:    "./domains.txt",
		SubDomainsFile: "./subdomains.txt",
		LinksFile:      "./links.txt",
	}

	creator := urls_creator.New(config)
	urlsChan := creator.Start()

	checker := urls_checker.New(urlsChan)
	sucUrlsChan := checker.Start(100)

	writer.Writer("./", "sucsess.txt", sucUrlsChan)

	fmt.Printf("\n All data has been processed!")
}
