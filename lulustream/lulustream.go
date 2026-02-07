// Package lulustream provides a wrapper for lulustream.com
package lulustream

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"

	"resty.dev/v3"

	"github.com/kissanjamgit/ext"
)

type Lulustream struct {
	Source string
}

func New(source string) ext.Site {
	return &Lulustream{Source: source}
}

func (l *Lulustream) Resource(client *resty.Client) (cr ext.ContentResource, err error) {
	res, err := client.R().SetHeaders(ext.Header).Get(l.Source)
	if err != nil {
		return
	}
	var URL string
	if urlSubmatch := regexp.MustCompile(`sources: \[{file:"(http[^"]+)"`).FindStringSubmatch(res.String()); len(urlSubmatch) < 2 {
		err = fmt.Errorf("len(urlSubmatch) <=2")
		return
	} else {
		URL = urlSubmatch[1]
	}

	var Name string
	nameRe := regexp.MustCompile(`<h1\s+class="h5">([^<]+)<`)
	if nameSubmatch := nameRe.FindStringSubmatch(res.String()); len(nameSubmatch) < 2 {
		err = fmt.Errorf("len(nameSubmatch) <=2")
		return
	} else {
		Name = nameSubmatch[1]
	}

	cr = ext.ContentResource{URL: URL, Name: Name}
	return
}

func (*Lulustream) Download(cr ext.ContentResource) (err error) {
	cmd := exec.Command("yt-dlp.exe", cr.URL, "-o", cr.Name)
	var header []string
	for k, v := range ext.Header {
		header = append(header, []string{"--add-header", fmt.Sprintf("%s: %s", k, v)}...)
	}
	cmd.Args = append(cmd.Args, header...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	return
}
