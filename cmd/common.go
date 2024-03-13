package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type FileLoader struct {
	gitToken string
}

func NewFileLoader(gitToken string) FileLoader {
	return FileLoader{gitToken: gitToken}
}

func (fl FileLoader) GetFile(uri string) ([]byte, error) {
	if strings.HasPrefix(uri, "http") {
		return fl.getHTTPFile(uri)
	}

	if _, err := os.Stat(uri); os.IsNotExist(err) {
		return nil, fmt.Errorf("file %s does not exist", uri)
	}

	return os.ReadFile(uri)
}

func (fl FileLoader) getHTTPFile(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("PRIVATE-TOKEN", fl.gitToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("the URL %s response with code not espected: %d", url, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
