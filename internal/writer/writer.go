package writer

import (
	"fmt"
	"log"
	"os"
)

func Writer(prefix string, filename string, ch chan string) {
	fp, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Println(prefix, err)
		log.Fatal(prefix, err)
	}
	defer fp.Close()
	for s := range ch {
		_, err = fp.WriteString(s + "\n")
		if err != nil {
			fmt.Println(prefix, err)
			log.Fatal(prefix, err)
		}
	}
}
