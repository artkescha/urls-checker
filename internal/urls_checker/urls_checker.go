package urls_checker

import (
	"fmt"
	"net/http"
	"sync"
)

type Checker struct {
	wg          sync.WaitGroup
	inUrlsChan  chan string
	sucUrlsChan chan string
	client      *http.Client
}

func New(urlsChan chan string) *Checker {
	return &Checker{inUrlsChan: urlsChan, sucUrlsChan: make(chan string), client: http.DefaultClient}
}

func (c *Checker) Start(workersCount int) chan string {
	for i := 0; i < workersCount; i++ {
		c.wg.Add(1)
		go c.run()
	}
	go func() {
		c.wg.Wait()
		close(c.sucUrlsChan)
	}()
	return c.sucUrlsChan
}

func (c *Checker) run() {
	for url := range c.inUrlsChan {
		resp, err := c.client.Get(url)
		if err != nil {
			fmt.Printf("http get request fail %s", err)
			continue
		}
		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			fmt.Printf("http get request fail status %d", resp.StatusCode)
			continue
		}

		c.sucUrlsChan <- url
	}
	c.wg.Done()
}
