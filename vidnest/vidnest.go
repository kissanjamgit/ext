// Package vidnest provides a wrapper for vidnest.io
package vidnest

import (
	"encoding/json"
	"os"
	"os/exec"

	"github.com/kissanjamgit/ext"
	"resty.dev/v3"
)

type Vidnest struct {
	Source string
}

func New(source string) ext.Site {
	return &Vidnest{Source: source}
}

func (v *Vidnest) Resource(client *resty.Client) (cr ext.ContentResource, err error) {
	cmd := exec.Command("yt-dlp", v.Source, "--print", ext.PrtFmt)
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

func (*Vidnest) Download(cr ext.ContentResource) (err error) {
	cmd := exec.Command("yt-dlp", cr.URL, "-o", cr.Name+".mp4")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	return
}
