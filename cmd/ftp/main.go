package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/gocs/lutz"
	"github.com/jlaffaye/ftp"
)

const (
	releasesHost    = "ftp.iana.org:21"
	releasesPath    = "/tz/tzdata-latest.tar.gz"
	timezoneFileOut = "tz"
)

// dl download the timezone file from the url to the writer

func dl(ctx context.Context, w io.Writer, host, path string) error {
	c, err := ftp.Dial(host, ftp.DialWithContext(ctx))
	if err != nil {
		return fmt.Errorf("error dialing ftp: %w", err)
	}
	defer func() {
		if err := c.Quit(); err != nil {
			log.Printf("error ftp quitting: %v", err)
		}
	}()

	if err = c.Login("anonymous", "anonymous"); err != nil {
		return fmt.Errorf("error ftp login: %w", err)
	}
	r, err := c.Retr(path)
	if err != nil {
		return fmt.Errorf("error ftp retrieval: %w", err)
	}
	defer r.Close()

	_, err = io.Copy(w, r)
	if err != nil {
		return fmt.Errorf("error copying response: %w", err)
	}

	return nil
}

func main() {
	ctx := context.Background()

	// download the selected file
	// open the input file for reading from the source
	fileIn := bytes.NewBuffer(nil)
	if err := dl(ctx, fileIn, releasesHost, releasesPath); err != nil {
		log.Fatalf("error downloading file: %v", err)
	}

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
