package lutz_test

import (
	_ "embed"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/gocs/lutz"
)

//go:embed testdata/releases.html
var releasesHTML string

func TestGetLatestRelease(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		r       io.Reader
		want    *lutz.Release
		wantErr bool
	}{
		{
			name: "get tzdata2026c.tar.gz (2026-07-08 18:02)",
			r:    strings.NewReader(releasesHTML),
			want: &lutz.Release{Name: "tzdata2026c.tar.gz", LastModified: func() time.Time {
				parsedTime, _ := time.Parse("2006-01-02 15:04", "2026-07-08 18:02")
				return parsedTime
			}()},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := lutz.GetLatestRelease(tt.r)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetLatestRelease() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetLatestRelease() succeeded unexpectedly")
			}
			if got.Name == tt.want.Name {
				t.Errorf("Release name does not match = %v, want %v", got, tt.want)
			}
			diff := got.LastModified.Sub(tt.want.LastModified)
			if diff != 0 {
				t.Errorf("Release LastModified does not match by %v = %v, want %v", diff, got, tt.want)
			}
		})
	}
}
