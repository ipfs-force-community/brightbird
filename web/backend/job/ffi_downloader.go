package job

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/google/go-github/v50/github"
)

type FFIDownloader interface {
	DownloadFFI(ctx context.Context, releaseTag string) (string, error)
}

type ffiDownloader struct {
	tempPath string
	token    string
}

func NewFFIDownloader(token string) FFIDownloader {
	return &ffiDownloader{
		tempPath: os.TempDir(),
		token:    token,
	}
}

func (downloader ffiDownloader) DownloadFFI(ctx context.Context, releaseTag string) (string, error) {
	fileName := releaseTag + "-filecoin-ffi-Linux-standard.tar.gz"
	filePath := path.Join(downloader.tempPath, fileName)

	_, err := os.Stat(filePath)
	if err == nil {
		return filePath, nil
	}
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}

	client := github.NewTokenClient(ctx, downloader.token)
	tag, _, err := client.Repositories.GetReleaseByTag(ctx, "filecoin-project", "filecoin-ffi", releaseTag[0:16])
	if err != nil {
		return "", err
	}

	var linuxAssert *github.ReleaseAsset
	for _, assert := range tag.Assets {
		if assert.Name != nil && strings.Contains(*assert.Name, "Linux") && assert.URL != nil {
			linuxAssert = assert
		}
	}
	if linuxAssert == nil {
		return "", fmt.Errorf("linux release for tag %s not exit", releaseTag)
	}

	body, _, err := client.Repositories.DownloadReleaseAsset(ctx, "filecoin-project", "filecoin-ffi", *linuxAssert.ID, client.Client())
	if err != nil {
		return "", err
	}

	fs, err := os.Create(filePath)
	if err != nil {
		return "", err
	}

	_, err = io.Copy(fs, body)
	if err != nil {
		return "", err
	}
	return filePath, nil
}

func uncompressFFI(tarPath string, dst string) error {
	gzipStream, err := os.Open(tarPath)
	if err != nil {
		return err
	}

	uncompressedStream, err := gzip.NewReader(gzipStream)
	if err != nil {
		return err
	}

	tarReader := tar.NewReader(uncompressedStream)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		switch header.Typeflag {
		case tar.TypeReg:
			_, fileName := path.Split(header.Name)
			dstPath := path.Join(dst, fileName)
			outFile, err := os.Create(dstPath)
			if err != nil {
				return err
			}

			if _, err := io.Copy(outFile, tarReader); err != nil {
				return err
			}
			err = outFile.Close()
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("extractTarGz: uknown type: %d in %s", header.Typeflag, header.Name)
		}

	}
	return nil
}
