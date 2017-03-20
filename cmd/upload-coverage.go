package cmd

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/codeclimate/test-reporter/version"
	"github.com/gobuffalo/envy"
	"github.com/spf13/cobra"
)

type Uploader struct {
	Input       string
	ReporterID  string
	EndpointURL string
}

var uploadOptions = Uploader{}

// uploadCoverageCmd represents the upload command
var uploadCoverageCmd = &cobra.Command{
	Use:   "upload-coverage",
	Short: "Upload pre-formatted coverage payloads to Code Climate servers.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return uploadOptions.Upload()
	},
}

func (u Uploader) Upload() error {
	if u.ReporterID == "" {
		return errors.New("you must supply a CC_TEST_REPORTER_ID ENV variable or pass it via the -r flag")
	}

	var err error
	var in io.Reader
	if u.Input == "-" {
		in = os.Stdin
	} else {
		in, err = os.Open(u.Input)
		if err != nil {
			return err
		}
	}

	c := http.Client{
		Timeout: 30 * time.Second,
	}
	req, err := http.NewRequest("POST", u.EndpointURL, in)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", fmt.Sprintf("TestReporter/%s (Code Climate, Inc.)", version.Version))
	req.Header.Set("Content-Type", "application/json")

	res, err := c.Do(req)
	if err != nil {
		return err
	}
	b, _ := ioutil.ReadAll(res.Body)
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("response from %s was %d: %s", u.EndpointURL, res.StatusCode, string(b))
	}
	fmt.Printf("Status: %d\n", res.StatusCode)
	fmt.Println(string(b))
	return nil
}

func init() {
	uploadCoverageCmd.Flags().StringVarP(&uploadOptions.Input, "input", "i", "coverage/codeclimate.json", "input path")
	uploadCoverageCmd.Flags().StringVarP(&uploadOptions.ReporterID, "id", "r", os.Getenv("CC_TEST_REPORTER_ID"), "reporter identifier")
	uploadCoverageCmd.Flags().StringVarP(&uploadOptions.EndpointURL, "endpoint", "e", envy.Get("CC_TEST_REPORTER_COVERAGE_ENDPOINT", "https://codeclimate.com/test_reports"), "endpoint to upload coverage information to")
	RootCmd.AddCommand(uploadCoverageCmd)
}
