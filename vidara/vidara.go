// package Vidara provides a wrapper for vidara.so
package vidara

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/kissanjamgit/ext"
	"resty.dev/v3"
)

type Vidara struct {
	Source string
}

func New(source string) ext.Site {
	return &Vidara{Source: source}
}

type jsonParse struct {
	URL  string `json:"streaming_url"`
	Name string `json:"title"`
}

func (v *Vidara) Resource(client *resty.Client) (cr ext.ContentResource, err error) {
	var ID string
	if l := strings.Split(v.Source, "/"); len(l) == 0 {
		err = fmt.Errorf("len(l) == 0")
		return
	} else {
		ID = l[len(l)-1]
	}

	res, err := client.R().Get(fmt.Sprintf("https://vidara.so/api/stream?filecode=%s", ID))
	if err != nil {
		return
	}
	var data jsonParse
	err = json.Unmarshal(res.Bytes(), &data)
	if err != nil {
		return
	}
	if data.URL == "" {
		err = fmt.Errorf(`data.URL == ""`)
		return
	}
	var name string
	if data.Name != "" {
		name = data.Name
	} else {
		name = fmt.Sprintf("Vidara_%s", ID)
	}

	cr = ext.ContentResource{URL: data.URL, Name: name}
	return
}

func (*Vidara) Download(cr ext.ContentResource) (err error) {
	cmd := exec.Command("yt-dlp", cr.URL, "-o", cr.Name+".mp4")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	return
}
