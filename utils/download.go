package utils

import (
	"archive/zip"
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

// DownloadAndUnzipFile download the zip file from the URL and put it's content in the file
func DownloadAndUnzipFile(ctx context.Context, client HTTPInterface, file *os.File, URL string) error {

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, URL, nil)
	if err != nil {
		return err
	}

	var resp *http.Response
	if resp, err = client.Do(req); err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	zipFileName := filepath.Base(file.Name()) + ".zip"

	var zipFile *os.File
	zipFile, err = ioutil.TempFile("", zipFileName)
	if err != nil {
		return err
	}
	defer func() {
		_ = os.Remove(zipFile.Name())
	}()

	// Write the body to file
	_, err = io.Copy(zipFile, resp.Body)
	if err != nil {
		return err
	}

	var reader *zip.ReadCloser
	reader, err = zip.OpenReader(zipFile.Name())
	if err != nil {
		return err
	}
	defer func() {
		_ = reader.Close()
	}()

	if len(reader.File) == 1 {
		in, _ := reader.File[0].Open()
		defer func() {
			_ = in.Close()
		}()
		_, err = io.Copy(file, in)
	}

	return err
}
