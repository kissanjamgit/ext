// Package fidelity provides
package fidelity

import (
	"encoding/json"
	"fmt"
	"html"
	"os"
	"os/exec"
	"regexp"

	"github.com/kissanjamgit/ext"
	"resty.dev/v3"
)

type Fidelity struct {
	Source string
}

func New(source string) ext.Site {
	return &Fidelity{Source: source}
}

type quality struct {
	URL        string `json:"url"`
	Resolution string `json:"resolution"`
}

func (f *Fidelity) Resource(client *resty.Client) (cr ext.ContentResource, err error) {
	res, err := client.R().Get(f.Source)
	if err != nil {
		return
	}
	// os.WriteFile("content.html", res.Bytes(), 0o644)
	submatch := regexp.MustCompile(`data-sources="(.*)"`).FindStringSubmatch(string(res.Bytes()))
	if len(submatch) < 2 {
		err = fmt.Errorf("len(submatch) < 2")
		return
	}
	data := html.UnescapeString(submatch[1])
	var qlist []quality
	json.Unmarshal([]byte(data), &qlist)

	maxQuality := "1080p"

	m := map[string]quality{}
	for _, i := range qlist {
		m[i.Resolution] = i
	}

	var q quality
	for index := len(qlist) - 1; index >= 0; index-- {
		q = qlist[index]
		if q.Resolution != maxQuality {
			continue
		}
		break
	}

	submatch = regexp.MustCompile(`<title>([^<]*)</title`).FindStringSubmatch(res.String())
	if len(submatch) < 2 {
		err = fmt.Errorf("len(submatch) < 2")
		return
	}
	cr = ext.ContentResource{Name: submatch[1], URL: q.URL}
	return
}

func (f *Fidelity) Download(cr ext.ContentResource) (err error) {
	cmd := exec.Command("yt-dlp", cr.URL, "-o", cr.Name)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	return
}
