// Package ext
package ext

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"resty.dev/v3"
)

var Header = map[string]string{
	"User-Agent":                "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:135.0) Gecko/20100101 Firefox/135.0",
	"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
	"Accept-Language":           "en-US,en;q=0.5",
	"Accept-Encoding":           "gzip, deflate",
	"DNT":                       "1",
	"Sec-GPC":                   "1",
	"Connection":                "keep-alive",
	"Upgrade-Insecure-Requests": "1",
	"Sec-Fetch-Dest":            "document",
	"Sec-Fetch-Mode":            "navigate",
	"Sec-Fetch-Site":            "cross-site",
	"Priority":                  "u=0, i",
	"TE":                        "trailers",
}

var (
	Quality = []string{"240p", "480p", "540p", "720p", "1080p", "4k"}
	PrtFmt  = `{"name":%(title)#j, "url":%(url)#j}`
)

type YtdlpResource struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

func (y *YtdlpResource) ToContentResource() ContentResource {
	return ContentResource{
		Name: y.Name,
		URL:  y.URL,
	}
}

type Site interface {
	Resource(*resty.Client) (ContentResource, error)
	Download(ContentResource) error
}

type ContentResource struct {
	Name string
	URL  string
}

type SiteAlter struct {
	Source string
}

func NewSiteAlter(Source string) Site {
	return &SiteAlter{Source}
}

func (s *SiteAlter) Resource(*resty.Client) (cr ContentResource, err error) {
	var header []string
	// header = append(header, "--playlist-items", "1")
	for k, v := range Header {
		header = append(header, []string{"--add-headers", fmt.Sprintf("%s: %s", k, v)}...)
	}
	arg := []string{"yt-dlp", s.Source, "--print", PrtFmt}
	arg = append(arg, header...)
	cmd := exec.Command("yt-dlp")
	cmd.Args = arg
	output, err := cmd.Output()
	if err != nil {
		return
	}
	ytResource := YtdlpResource{}
	err = json.Unmarshal(output, &ytResource)
	if err != nil {
		return
	}
	cr = ytResource.ToContentResource()
	return
}

func (s *SiteAlter) Download(cr ContentResource) (err error) {
	// var header []string
	// for k, v := range Header {
	// 	header = append(header, []string{"--add-header", fmt.Sprintf("%s: %s", k, v)}...)
	// }
	arg := []string{"yt-dlp", "-v", s.Source, "-o", cr.Name + ".mp4"}
	// arg = append(arg, header...)
	cmd := exec.Command("yt-dlp")
	cmd.Stderr = os.Stderr
	cmd.Args = arg
	// cmd.Args = header
	err = cmd.Run()
	return
}
