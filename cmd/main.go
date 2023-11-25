package main

import (
	"fmt"
	"urls-checker/internal/unzip"
	"urls-checker/internal/urls_checker"
	"urls-checker/internal/urls_creator"
	"urls-checker/internal/writer"

	"urls-checker/config"
)

func main() {
	// Обрабатываем параметры командной строки и загружаем конфигурацию
	config.ParseCmdLine()

	if *config.Opts.Unzip {
		unzip.UnzipArch("./subdomains.zip", "./subdomains.txt")
		unzip.UnzipArch("./domains.zip", "./domains.txt")
		unzip.UnzipArch("./links.zip", "./links.txt")

	}

	config := urls_creator.Config{
		DomainsFile:    "./domains.txt",
		SubDomainsFile: "./subdomains.txt",
		LinksFile:      "./links.txt",
	}

	creator := urls_creator.New(config)
	urlsChan := creator.Start()

	checker := urls_checker.New(urlsChan)
	sucUrlsChan, failUrlsChan := checker.Start(100)

	go writer.Writer("./", "fail.txt", failUrlsChan)
	writer.Writer("./", "sucsess.txt", sucUrlsChan)

	fmt.Printf("\n All data has been processed!")
}
