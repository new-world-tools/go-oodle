package oodle

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const dllUrl = "https://github.com/new-world-tools/go-oodle/releases/download/untagged-60360a4ffdf148b31c76/oo2core_9_win64.dll"

func Download() error {
	req, err := http.NewRequest("GET", dllUrl, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	filePath := getTempDllPath()
	err = os.MkdirAll(filepath.Dir(filePath), 0777)
	if err != nil {
		return err
	}

	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
