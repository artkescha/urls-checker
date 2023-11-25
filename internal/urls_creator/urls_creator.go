package urls_creator

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
)

type UrlsCreator struct {
	urlsChan chan string
	config   Config
}

func New(config Config) *UrlsCreator {
	return &UrlsCreator{urlsChan: make(chan string), config: config}
}

func (c UrlsCreator) Start() chan string {
	domains, err := processDomains(c.config.domainsFile)
	if err != nil {
		fmt.Printf("read domains fail %s", err)
	}
	subDomains, err := processDomains(c.config.subDomainsFile)
	if err != nil {
		fmt.Printf("read subDomains fail %s", err)
	}
	links, err := processDomains(c.config.linksFile)
	if err != nil {
		fmt.Printf("read links fail %s", err)
	}
	go c.run(domains, subDomains, links)

	return c.urlsChan
}

func (c UrlsCreator) run(domains, subDomains, links []string) {
	defer close(c.urlsChan)

	for _, domain := range domains {
		for _, subDomain := range subDomains {
			for _, link := range links {
				c.urlsChan <- filepath.Join(domain, subDomain, link)
			}
		}
	}
}

func processDomains(domainsFile string) ([]string, error) {
	df, err := os.OpenFile(domainsFile, os.O_RDONLY, 0755)
	if err != nil {
		fmt.Println("domains file: ", err)
		return nil, err
	}
	defer df.Close()

	scaner := bufio.NewScanner(df)
	domains := make([]string, 0)
	for scaner.Scan() {
		d := scaner.Text()
		if d != "" {
			domains = append(domains, d)
		}
	}
	if err = scaner.Err(); err != nil {
		fmt.Println("domains file: ", err)
		return nil, err
	}

	return domains, nil
}
