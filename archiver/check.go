package archiver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"
)

type ArchiveMetadata struct {
	LatestUpdate JSONDate
	Files        []ArchivedFiles `json:",omitempty"`
}

func ReadMetadata(path string, metadata *ArchiveMetadata) error {
	fp, err := os.Open(path)
	if err != nil {
		return err
	}
	defer fp.Close()

	err = json.NewDecoder(fp).Decode(&metadata)
	if err != nil {
		return err
	}

	return nil
}

type ArchivedFiles struct {
	Date JSONDate
	TXT  *ArchivedCompRulesData
	PDF  *ArchivedCompRulesData
	DOCX *ArchivedCompRulesData
}

type ArchivedCompRulesData struct {
	Path string

	URL           string
	ContentLength int64
	ContentType   string
	LastModified  string
	ETag          string
}

var ErrNotFound = fmt.Errorf("no rules found")

func Check(data ArchivedFiles) (ArchivedFiles, error) {
	out := ArchivedFiles{
		Date: data.Date,
	}

	for _, ext := range []string{"txt", "pdf", "docx"} {
		url, err := url.Parse(fmt.Sprintf(
			"https://media.wizards.com/%d/downloads/MagicCompRules%%20%s.%s",
			time.Time(data.Date).Year(),
			time.Time(data.Date).Format("20060102"),
			ext,
		))
		if err != nil {
			panic(err)
		}

		req := http.Request{
			Method: http.MethodHead,
			URL:    url,
		}

		resp, err := http.DefaultClient.Do(&req)
		if err != nil {
			return out, err
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusNotFound {
			continue
		}

		if resp.StatusCode != http.StatusOK {
			return out, fmt.Errorf("failed to check %q: %s", url, resp.Status)
		}

		data := ArchivedCompRulesData{
			Path: fmt.Sprintf("%s.%s", time.Time(data.Date).Format("2006-01-02"), ext),

			URL:           url.String(),
			ContentLength: resp.ContentLength,
			ContentType:   resp.Header.Get("Content-Type"),
			LastModified:  resp.Header.Get("Last-Modified"),
			ETag:          resp.Header.Get("ETag"),
		}

		switch ext {
		case "txt":
			out.TXT = &data
		case "pdf":
			out.PDF = &data
		case "docx":
			out.DOCX = &data
		}
	}

	if out.TXT == nil && out.PDF == nil && out.DOCX == nil {
		return out, ErrNotFound
	}

	return out, nil
}

// Use when downloading the file.

// var (
// 	ifModifiedSince string
// 	ifNoneMatch     string
// )

// switch ext {
// case "txt":
// 	if data.TXT != nil {
// 		ifModifiedSince = data.TXT.LastModified.Format(http.TimeFormat)
// 		ifNoneMatch = data.TXT.ETag
// 	}
// case "pdf":
// 	if data.PDF != nil {
// 		ifModifiedSince = data.PDF.LastModified.Format(http.TimeFormat)
// 		ifNoneMatch = data.PDF.ETag
// 	}
// case "docx":
// 	if data.DOCX != nil {
// 		ifModifiedSince = data.DOCX.LastModified.Format(http.TimeFormat)
// 		ifNoneMatch = data.DOCX.ETag
// 	}
// }

// headers := http.Header{}
// if ifModifiedSince != "" {
// 	headers.Set("If-Modified-Since", ifModifiedSince)
// }
// if ifNoneMatch != "" {
// 	headers.Set("If-None-Match", ifNoneMatch)
// }
