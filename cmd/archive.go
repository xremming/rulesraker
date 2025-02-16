package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"github.com/xremming/rulesraker/archiver"
)

var (
	dateMargin        int
	dateStartOverride FlagDate
	dateEndOverride   FlagDate
	jobLimit          int
)

func archiveRun(cmd *cobra.Command, args []string) error {
	now := time.Now().UTC()

	fp, err := os.Open(filepath.Join(archiveDir, "metadata.json"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("metadata.json not found in %s", archiveDir)
		}

		return err
	}
	defer fp.Close()

	var metadata archiver.Metadata
	err = json.NewDecoder(fp).Decode(&metadata)
	if err != nil {
		return err
	}

	newMetadata := metadata
	newMetadata.LatestUpdate = archiver.JSONDate(now)

	startDate := time.Time(metadata.LatestUpdate).AddDate(0, 0, -dateMargin)
	if !time.Time(dateStartOverride).IsZero() {
		startDate = time.Time(dateStartOverride)
	}

	endDate := now.AddDate(0, 0, dateMargin+1)
	if !time.Time(dateEndOverride).IsZero() {
		endDate = time.Time(dateEndOverride).AddDate(0, 0, 1)
	}

	jobs := make(chan archiver.PossibleURL)

	go func() {
		for dateToCheck := startDate; dateToCheck.Before(endDate); dateToCheck = dateToCheck.AddDate(0, 0, 1) {
			urlsToCheck := metadata.URLFormats.PossibleURLs(dateToCheck)
			for _, urlToCheck := range urlsToCheck {
				jobs <- urlToCheck
			}
		}
		close(jobs)
	}()

	results := make(chan archiver.Rule)
	var wg sync.WaitGroup

	for range jobLimit {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for possibleURL := range jobs {
				logPrefix := possibleURL.Date.Format("2006-01-02")

				if !possibleURL.Available {
					cmd.Println(logPrefix, "possible url marked as unavailable, skipping", possibleURL.URL.String())
					continue
				}

				cmd.Println(logPrefix, "checking the following URL for rules:", possibleURL.URL.String())

				resp, err := http.DefaultClient.Head(possibleURL.URL.String())
				if err != nil {
					cmd.Println(logPrefix, "failed to check:", err)
					continue
				}

				if resp.StatusCode == http.StatusNotFound {
					continue
				}

				if resp.StatusCode != http.StatusOK {
					cmd.Println(logPrefix, "failed to check:", resp.Status)
					continue
				}

				cmd.Println(logPrefix, "rules found")
				url := possibleURL.URL.String()
				results <- archiver.Rule{
					Date:   archiver.JSONDate(possibleURL.Date),
					Format: possibleURL.Format,
					File: fmt.Sprintf(
						"%s/%s.%s",
						possibleURL.Format,
						possibleURL.Date.Format("2006-01-02"),
						possibleURL.Format,
					),
					URL: &url,
					ResponseMetadata: &archiver.ResponseMetadata{
						ContentLength: resp.ContentLength,
						ContentType:   resp.Header.Get("Content-Type"),
						LastModified:  resp.Header.Get("Last-Modified"),
						ETag:          resp.Header.Get("ETag"),
					},
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for result := range results {
		newMetadata.Rules = append(newMetadata.Rules, result)
	}

	newMetadata.PrepareForEncoding()

	tmpFile, err := os.CreateTemp(archiveDir, "metadata-*.json")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

	encoder := json.NewEncoder(tmpFile)
	encoder.SetIndent("", "  ")

	err = encoder.Encode(&newMetadata)
	if err != nil {
		tmpFile.Close()
		return err
	}

	err = tmpFile.Close()
	if err != nil {
		return err
	}

	err = os.Rename(tmpFile.Name(), filepath.Join(archiveDir, "metadata.json"))
	if err != nil {
		return err
	}

	return nil
}

var archiveCmd = &cobra.Command{
	Use:   "archive",
	Short: "Scrape and archive rules files from wizards.com based on a configuration",
	Args:  cobra.NoArgs,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return os.MkdirAll(archiveDir, 0o755)
	},
	RunE: archiveRun,
}

func init() {
	rootCmd.AddCommand(archiveCmd)
	archiveCmd.Flags().SortFlags = false

	archiveCmd.Flags().IntVarP(&dateMargin, "date-margin", "m", 3,
		"number of days to check before the latest update and after the current date",
	)
	archiveCmd.Flags().VarP(&dateStartOverride, "date-start", "s",
		"override the start date",
	)
	archiveCmd.Flags().VarP(&dateEndOverride, "date-end", "e",
		"override the end date",
	)
	archiveCmd.Flags().IntVarP(&jobLimit, "jobs", "j", 4,
		"number of concurrent jobs to run",
	)
}
