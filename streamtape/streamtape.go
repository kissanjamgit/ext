// Package streamtape provides a wrapper for streamtape.com
package streamtape

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/kissanjamgit/ext"
	"resty.dev/v3"
)

type StreamTape struct {
	Source string
}

func New(source string) ext.Site {
	return &StreamTape{Source: source}
}

func (s *StreamTape) Resource(client *resty.Client) (cr ext.ContentResource, err error) {
	res, err := client.R().
		SetHeaders(ext.Header).
		Get(s.Source)
	if err != nil {
		return
	}
	var title string
	titleRe := regexp.MustCompile(`(?i)<h2>(.*?)</h2>`)
	if titleMatches := titleRe.FindStringSubmatch(res.String()); len(titleMatches) <= 1 {
		err = fmt.Errorf("len(titleMatches) <=1")
		return
	} else {
		title = strings.TrimSpace(titleMatches[1])
	}

	re := regexp.MustCompile(`innerHTML\W*([^\'\\"]*)\W*xcd([^bd][^\\\'"]*)`)
	matches := re.FindStringSubmatch(res.String())

	if len(matches) < 3 {
		err = fmt.Errorf("len(matches) < 3")
		return
	}
	videoURL := "https://" + matches[1] + matches[2]
	cr = ext.ContentResource{
		Name: title,
		URL:  videoURL,
	}
	return
}

func (*StreamTape) Download(cr ext.ContentResource) (err error) {
	cmd := exec.Command("yt-dlp.exe", cr.URL, "-o", cr.Name)
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	return
}
