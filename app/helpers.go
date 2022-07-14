package app

import (
	"encoding/gob"
	"io"
	"log"
	"net/http"
	"os"
	"path"
)

func dumpGob(path string, data interface{}, logger *log.Logger) error {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer func() {
		if err := f.Close(); err != nil {
			logger.Printf("error closing file: %s", err.Error())
		}
	}()

	return gob.NewEncoder(f).Encode(data)
}

func readGob(path string, data interface{}, logger *log.Logger) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() {
		if err := f.Close(); err != nil {
			logger.Printf("error closing file: %s", err.Error())
		}
	}()

	return gob.NewDecoder(f).Decode(data)
}

func mkdir(path string) error {
	if _, err := os.Open(path); os.IsNotExist(err) {
		return os.MkdirAll(path, 0777)
	}
	return nil
}

func abspath(p string) (string, error) {
	if !path.IsAbs(p) {
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		p = path.Join(cwd, p)
	}
	return p, nil
}

func download(url string, path string, logger *log.Logger) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.Printf("error closing response body: %s", err.Error())
		}
	}()

	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		if err := out.Close(); err != nil {
			logger.Printf("error closing file: %s", err.Error())
		}
	}()

	_, err = io.Copy(out, resp.Body)
	return err
}
