package oodle

import (
	"bufio"
	"crypto/md5"
	"errors"
	"fmt"
	"github.com/itchio/lzma"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const baseUrl = "https://origin.warframe.com/origin/E926E926"

func Download() error {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", baseUrl, "index.txt.lzma"), nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(lzma.NewReader(resp.Body))

	var urlPath string
	for scanner.Scan() {
		if dllRe.Match(scanner.Bytes()) {
			sep := ","
			parts := strings.Split(scanner.Text(), sep)
			urlPath = strings.Join(parts[:len(parts)-1], sep)
			break
		}
	}

	err = scanner.Err()
	if err != nil {
		return err
	}

	if urlPath == "" {
		return errors.New("not found.")
	}

	req, err = http.NewRequest("GET", fmt.Sprintf("%s/%s", baseUrl, urlPath), nil)
	if err != nil {
		return err
	}

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	baseName := path.Base(urlPath)

	sep := "."
	parts := strings.Split(baseName, sep)
	if len(parts) != 4 {
		return errors.New("wrong file name format.")
	}

	//fileName := strings.Join(parts[:len(parts)-2], sep)
	md5sum := parts[len(parts)-2]

	filePath := getTempDllPath()
	err = os.MkdirAll(filepath.Dir(filePath), 0777)
	if err != nil {
		return err
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	md5Writer := md5.New()

	r := io.TeeReader(lzma.NewReader(resp.Body), file)

	for {
		b := make([]byte, 1)
		_, err = r.Read(b)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		_, err = md5Writer.Write(b)
		if err != nil {
			return err
		}
	}

	if fmt.Sprintf("%X", md5Writer.Sum(nil)) != md5sum {
		file.Close()
		os.Remove(file.Name())
		return errors.New("file is corrupted.")
	}

	return nil
}
