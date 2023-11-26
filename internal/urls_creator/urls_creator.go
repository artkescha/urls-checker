package urls_creator

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
)

type UrlsCreator struct {
	wg       sync.WaitGroup
	urlsChan chan string
	config   Config
}

func New(config Config) *UrlsCreator {
	return &UrlsCreator{urlsChan: make(chan string), config: config}
}

func (c *UrlsCreator) Start() chan string {
	var err error
	var domains, subDomains, links []string

	c.wg.Add(1)
	go func() {
		domains, err = processDomains(c.config.DomainsFile)
		if err != nil {
			fmt.Printf("read domains fail %s", err)
		}
		c.wg.Done()
	}()
	c.wg.Add(1)
	go func() {
		subDomains, err = processDomains(c.config.SubDomainsFile)
		if err != nil {
			fmt.Printf("read subDomains fail %s", err)
		}
		c.wg.Done()
	}()
	c.wg.Add(1)
	go func() {
		links, err = processDomains(c.config.LinksFile)
		if err != nil {
			fmt.Printf("read links fail %s", err)
		}
		c.wg.Done()
	}()

	c.wg.Wait()

	go c.run(domains, subDomains, links)

	return c.urlsChan
}

func (c *UrlsCreator) run(domains, subDomains, links []string) {
	defer close(c.urlsChan)

	for _, domain := range domains {
		for _, subDomain := range subDomains {
			for _, link := range links {

				subDomain = strings.Trim(subDomain, ".")
				url_, _ := url.JoinPath(domain, link)
				url_ = subDomain + "." + url_

				fmt.Printf("%s\n", url_)
				c.urlsChan <- url_
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
