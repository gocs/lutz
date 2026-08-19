package lutz

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Release struct {
	Name         string
	LastModified time.Time
}

const ianaTimeLayout = "2006-01-02 15:04"

func GetLatestRelease(r io.Reader) (*Release, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, fmt.Errorf("error upon creating doc from reader")
	}
	release := Release{}
	// TODO: update test to make sure all cases covered
	for _, s := range doc.Find(`a[href$=".gz"]`).EachIter() {
		name := strings.TrimSpace(s.Text())
		if !strings.Contains(name, ".gz") || strings.Contains(name, ".asc") {
			continue
		}

		lastModified := strings.TrimSpace(s.Parent().Next().Text())
		lm, err := time.Parse(ianaTimeLayout, lastModified)
		if err != nil {
			continue // skip
		}

		// if choose the more recent LastModified
		if release.LastModified.Sub(lm) >= 0 {
			continue
		}
		release.Name = name
		release.LastModified = lm
	}

	return &release, nil
}
