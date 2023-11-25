package unzip

import (
	"archive/zip"
	"bufio"
	"fmt"
	"os"
)

func UnzipArch(target, dst string) error {
	// открытие zip архива
	r, err := zip.OpenReader(target)
	if err != nil {
		return err
	}

	defer func() {
		r.Close()
	}()

	if len(r.File) == 0 {
		return fmt.Errorf("upload archive is empty")
	}

	saveFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer saveFile.Close()

	lines := 0
	for _, f := range r.File {
		file, err := f.Open()
		if err != nil {
			file.Close()
			return err
		}
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			_, err = saveFile.Write(scanner.Bytes())
			if err != nil {
				file.Close()
				return err
			}
			saveFile.Write([]byte("\n"))
			lines++
		}
		file.Close()
		err = scanner.Err()
		if err != nil {
			return err
		}
	}
	return nil
}
