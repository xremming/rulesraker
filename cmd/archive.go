package cmd

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/spf13/cobra"
	"github.com/xremming/rulesraker/archiver"
)

var (
	archiveDir           string
	dateMargin           int
	latestUpdateOverride FlagDate
	onlyMetadata         bool
)

func archiveRun(cmd *cobra.Command, args []string) error {
	now := time.Now().UTC()

	metadata := archiver.ArchiveMetadata{
		// This is the first known date that comprehensive rules are
		// available from the Wizards of the Coast website.
		LatestUpdate: archiver.JSONDate(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)),
	}

	err := archiver.ReadMetadata(filepath.Join(archiveDir, "metadata.json"), &metadata)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}

	newMetadata := archiver.ArchiveMetadata{
		LatestUpdate: archiver.JSONDate(now),
	}

	startFrom := time.Time(metadata.LatestUpdate).AddDate(0, 0, -dateMargin)
	if !time.Time(latestUpdateOverride).IsZero() {
		startFrom = time.Time(latestUpdateOverride)
	}

	runUntil := now.AddDate(0, 0, dateMargin)
	for dateToCheck := startFrom; dateToCheck.Before(runUntil); dateToCheck = dateToCheck.AddDate(0, 0, 1) {
		logPrefix := dateToCheck.Format("2006-01-02")

		var data archiver.ArchivedFiles

		var existingData *archiver.ArchivedFiles
		for i := range metadata.Files {
			if time.Time(metadata.Files[i].Date).Equal(dateToCheck) {
				cmd.Println(logPrefix, "existing data found for the date, rechecking it")
				existingData = &metadata.Files[i]
				break
			}
		}

		if existingData != nil {
			data = *existingData
		} else {
			data = archiver.ArchivedFiles{
				Date: archiver.JSONDate(dateToCheck),
			}
		}

		newData, err := archiver.Check(data)
		if err != nil {
			if errors.Is(err, archiver.ErrNotFound) {
				cmd.Println(logPrefix, "no rules found")
				continue
			}
			return err
		}

		cmd.Println(logPrefix, "rules found")
		newMetadata.Files = append(newMetadata.Files, newData)
	}

	sort.Slice(newMetadata.Files, func(i, j int) bool {
		return time.Time(newMetadata.Files[i].Date).Before(time.Time(newMetadata.Files[j].Date))
	})

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

	if onlyMetadata {
		return nil
	}

	return nil
}

var archiveCmd = &cobra.Command{
	Use:   "archive",
	Short: "Archive rules from wizards.com",
	Args:  cobra.NoArgs,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return os.MkdirAll(archiveDir, 0o755)
	},
	RunE: archiveRun,
}

func init() {
	rootCmd.AddCommand(archiveCmd)
	archiveCmd.Flags().StringVarP(&archiveDir, "archive-dir", "a", "archive",
		"directory to store stores the archived rules",
	)
	archiveCmd.MarkFlagDirname("archive-dir")

	archiveCmd.Flags().IntVarP(&dateMargin, "date-margin", "m", 3,
		"number of days to check before the latest update and after the current date",
	)

	archiveCmd.Flags().VarP(&latestUpdateOverride, "latest-update", "l",
		"override the latest update date, useful when needing to backfill data",
	)

	archiveCmd.Flags().BoolVar(&onlyMetadata, "only-metadata", false,
		"only update the metadata without downloading the files",
	)
}
