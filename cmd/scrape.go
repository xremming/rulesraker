package cmd

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	"github.com/spf13/cobra"
)

const rulesURL = "https://magic.wizards.com/en/rules"

var ruleLinksRegexp = regexp.MustCompile(`"([^"]+\.(docx|pdf|txt))"`)

func download(url, ext string) error {
	resp, err := http.DefaultClient.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("downloading %q returned a non 200 status code", url)
	}

	fp, err := os.CreateTemp(os.TempDir(), fmt.Sprintf("MagicCompRules_*.%s", ext))
	if err != nil {
		return err
	}

	_, err = io.Copy(fp, resp.Body)
	if err != nil {
		return errors.Join(err, fp.Close())
	}

	err = fp.Close()
	if err != nil {
		return err
	}

	return os.Rename(fp.Name(), filepath.Join(dataDir, fmt.Sprintf("MagicCompRules.%s", ext)))
}

func scrapeRun(cmd *cobra.Command, args []string) error {
	cmd.Printf("getting rules index page %q\n", rulesURL)
	resp, err := http.DefaultClient.Get(rulesURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	toDownload := make(map[string]string)

	matches := ruleLinksRegexp.FindAllStringSubmatch(string(body), -1)
	for _, match := range matches {
		ext := match[2]
		url := match[1]

		_, ok := toDownload[ext]
		if ok {
			return fmt.Errorf(
				"format %s is defined multiple times on %s, not sure which to download",
				ext, rulesURL,
			)
		}

		toDownload[ext] = url
	}

	var errDownload error
	for ext, url := range toDownload {
		cmd.Printf("downloading %4s from %q\n", ext, url)
		errDownload = errors.Join(errDownload, download(url, ext))
	}

	if errDownload != nil {
		return errDownload
	}

	return nil
}

var scrapeCmd = &cobra.Command{
	Use:     "scrape",
	Aliases: []string{"s"},
	Short:   "Scrape the latest comprehensive rules from wizards.com",
	Args:    cobra.NoArgs,
	RunE:    scrapeRun,
}

func init() {
	rootCmd.AddCommand(scrapeCmd)
}
