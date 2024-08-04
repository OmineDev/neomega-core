package download_wrapper

import (
	"bytes"
	"io"
	"net/http"
)

func DownloadMicroContent(sourceUrl string) ([]byte, error) {
	resp, err := http.Get(sourceUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	contents := bytes.NewBuffer([]byte{})
	if _, err := io.Copy(contents, resp.Body); err == nil {
		return contents.Bytes(), nil
	} else {
		return nil, err
	}
}
