// Package savefiles provides a wrapper for savefiles.io
package savefiles

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
	"github.com/kissanjamgit/ext"
	"resty.dev/v3"
)

type SaveFiles struct {
	Source string
}

func New(source string) ext.Site {
	return &SaveFiles{Source: source}
}

func (s *SaveFiles) Resource(_ *resty.Client) (cr ext.ContentResource, err error) {
	options := []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(30),
		tls_client.WithClientProfile(profiles.Chrome_103),
		tls_client.WithNotFollowRedirects(),
	}
	c, err := tls_client.NewHttpClient(tls_client.NewLogger(), options...)
	if err != nil {
		return
	}
	req, err := http.NewRequest(http.MethodGet, s.Source, nil)
	if err != nil {
		return
	}
	res, err := c.Do(req)
	if err != nil {
		return
	}
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return
	}
	os.WriteFile("content.html", data, 0o644)
	submatch := regexp.MustCompile(`sources:\s+\["([^"]*)"\]`).FindSubmatch(data)
	if len(submatch) < 2 {
		err = fmt.Errorf("len(submatch) < 2")
		return
	}
	uri := string(submatch[1])
	submatch = regexp.MustCompile(`<title>([^<]*)</title>`).FindSubmatch(data)
	if len(submatch) < 2 {
		err = fmt.Errorf("len(submatch) < 2")
		return
	}
	cr = ext.ContentResource{Name: string(submatch[1]), URL: uri}
	return
}

func (s *SaveFiles) Download(cr ext.ContentResource) (err error) {
	cmd := exec.Command("yt-dlp", cr.URL, "-o", cr.Name+".mp4")
	err = cmd.Run()
	return
}
