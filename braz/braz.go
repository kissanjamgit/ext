// Package brazz provides a wrapper for brazz.com
package brazz

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/kissanjamgit/ext"
	"resty.dev/v3"
)

type Braz struct {
	Source string
}

func New(source string) ext.Site {
	return &Braz{source}
}

type jsonParse struct {
	Title string `json:"name"`
	URL   string `json:"contentUrl"`
}

func (j *jsonParse) toCR() ext.ContentResource {
	return ext.ContentResource{Name: j.Title, URL: j.URL}
}

func getName(client *resty.Client, source string) (name string, err error) {
	res, err := client.R().Get(source)
	if err != nil {
		return
	}
	submatch := regexp.MustCompile(`<script[^>]*type="application/ld\+json">([^<]+)</script>`).FindStringSubmatch(string(res.Bytes()))
	if len(submatch) < 2 {
		err = fmt.Errorf("len(submatch) < 2")
		return
	}
	var data map[string]any
	json.Unmarshal([]byte(submatch[1]), &data)
	name, ok := data["name"].(string)
	if !ok {
		err = fmt.Errorf("name not string")
	}
	return
}

func (b *Braz) Resource(client *resty.Client) (cr ext.ContentResource, err error) {
	u, err := url.Parse(b.Source)
	if err != nil {
		return
	}
	uri := b.Source
	switch u.Hostname() {

	case "newbrazz.com":
		break
	case "www.brazzers.com":
		{
			// path := strings.Split(u.Path, "/")
			// if len(path) < 3 {
			// 	err = fmt.Errorf("len(path) < 3")
			// 	return
			// }
			// uri = fmt.Sprintf("https://newbrazz.com/video/%s/", clean(path[len(path)-1]))
			name, err := getName(client, b.Source)
			if err != nil {
				return cr, err
			}
			uri = fmt.Sprintf("https://newbrazz.com/video/%s/", name)
		}

	case "bang-free.com":
		break
	case "bangbros.com":
		{

			name, err := getName(client, b.Source)
			if err != nil {
				return cr, err
			}

			uri = fmt.Sprintf("https://bang-free.com/video/%s/", name)
		}

	}

	res, err := client.R().Get(uri)
	if err != nil {
		return
	}

	submatch := regexp.MustCompile(`<script\s+type="application/ld\+json">([^<]*)</script>`).FindStringSubmatch(string(res.Bytes()))
	if len(submatch) < 2 {
		err = fmt.Errorf("len(submatch) < 2")
		return
	}
	var j jsonParse
	err = json.Unmarshal([]byte(submatch[1]), &j)
	if err != nil {
		return
	}
	cr = j.toCR()
	return
}

func (*Braz) Download(cr ext.ContentResource) (err error) {
	cmd := exec.Command("yt-dlp", cr.URL, "-o", cr.Name)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	return
}

func clean(s string) string {
	list := strings.Split(s, "-")
	var result strings.Builder
	for _, item := range list {
		result.WriteString(strings.ToUpper(item[0:1]) + item[1:] + " ")
	}
	str := result.String()
	str = strings.TrimSpace(str)
	return str
}
