// Package savefiles provides a wrapper for savefiles.io
package savefiles

import (
	"encoding/json"
	"os/exec"

	"github.com/kissanjamgit/ext"
	"resty.dev/v3"
)

type SaveFiles struct {
	Source string
}

func New(source string) ext.Site {
	return &SaveFiles{Source: source}
}

func (s *SaveFiles) Resource(client *resty.Client) (cr ext.ContentResource, err error) {
	cmd := exec.Command("yt-dlp", s.Source, "--print", ext.PrtFmt)
	output, err := cmd.Output()
	if err != nil {
		return
	}
	var ytResource ext.YtdlpResource
	err = json.Unmarshal(output, &ytResource)
	if err != nil {
		return
	}
	cr = ytResource.ToContentResource()
	return
}

func (s *SaveFiles) Download(cr ext.ContentResource) (err error) {
	cmd := exec.Command("yt-dlp", cr.URL, "-o", cr.Name)
	err = cmd.Run()
	return
}
