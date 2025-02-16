package cmd

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/xremming/rulesraker/archiver"
)

var downloadAlways bool

func downloadRun(cmd *cobra.Command, args []string) error {
	fp, err := os.Open(filepath.Join(archiveDir, "metadata.json"))
	if err != nil {
		return err
	}

	var metadata archiver.Metadata
	err = json.NewDecoder(fp).Decode(&metadata)
	if err != nil {
		return err
	}

	for _, file := range metadata.Rules {
		if file.URL == nil {
			cmd.Println("skipping downloading of file as it does not have a URL", file.File)
			continue
		}

		url, err := url.Parse(*file.URL)
		if err != nil {
			panic(err)
		}

		path := filepath.Join(archiveDir, file.File)

		fileExists := false
		if _, err := os.Stat(path); err == nil {
			fileExists = true
		}

		headers := http.Header{}
		if !downloadAlways && file.ResponseMetadata != nil && fileExists {
			if file.ResponseMetadata.LastModified != "" {
				headers.Set("If-Modified-Since", file.ResponseMetadata.LastModified)
			}

			if file.ResponseMetadata.ETag != "" {
				headers.Set("If-None-Match", file.ResponseMetadata.ETag)
			}
		}

		req := http.Request{
			Method: http.MethodGet,
			URL:    url,
			Header: headers,
		}

		cmd.Println("downloading", file.URL)
		resp, err := http.DefaultClient.Do(&req)
		if err != nil {
			return err
		}

		if !downloadAlways && resp.StatusCode == http.StatusNotModified {
			cmd.Println("file not modified")
			continue
		}

		if resp.StatusCode != http.StatusOK {
			cmd.Printf("downloading %q returned a non 200 status code\n", file.URL)
			continue
		}

		fp, err := os.Create(path)
		if err != nil {
			return err
		}
		defer fp.Close()

		_, err = io.Copy(fp, resp.Body)
		if err != nil {
			return errors.Join(err, fp.Close())
		}
	}

	return nil
}

var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "download rules files based on the archive metadata",
	Args:  cobra.NoArgs,
	RunE:  downloadRun,
}

func init() {
	rootCmd.AddCommand(downloadCmd)

	downloadCmd.Flags().BoolVar(&downloadAlways, "always", false,
		"always download the files, even if they exist and are up to date",
	)
}
