package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/gocs/lutz"
)

const (
	timezoneFileURL = "https://data.iana.org/time-zones/releases"
	timezoneFileIn  = "tzdata2025c.tar.gz"
	timezoneFileOut = "tz"
)

// dl download the timezone file from the url to the writer
func dl(ctx context.Context, w io.Writer) error {
	urlPath, err := url.JoinPath(timezoneFileURL, timezoneFileIn)
	if err != nil {
		return fmt.Errorf("error joining url: %w", err)
	}
	request, err := http.NewRequestWithContext(ctx, "GET", urlPath, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return fmt.Errorf("error doing request: %w", err)
	}
	defer response.Body.Close()
	_, err = io.Copy(w, response.Body)
	if err != nil {
		return fmt.Errorf("error copying response body: %w", err)
	}
	return nil
}

func main() {
	ctx := context.Background()
	// open the input file for reading from the source
	fileIn := bytes.NewBuffer(nil)
	if err := dl(ctx, fileIn); err != nil {
		log.Fatalf("error downloading file: %v", err)
	}
	// fileIn, err := os.Open(timezoneFileIn)
	// if err != nil {
	// 	log.Fatalf("error opening file: %v", err)
	// }
	// defer fileIn.Close()

	// open the output file for writing to the destination
	fileOut, err := os.Create(timezoneFileOut)
	if err != nil {
		log.Fatalf("error creating file: %v", err)
	}
	defer fileOut.Close()

	if err := lutz.GetLookupTable(fileIn, fileOut); err != nil {
		log.Fatalf("error getting lookup table: %v", err)
	}

}
