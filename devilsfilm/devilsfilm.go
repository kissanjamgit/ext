// Package devilsfilm provides a wrapper for devilsfilm.com etc
package devilsfilm

import (
	"encoding/json"
	"fmt"
	"maps"
	"net/url"
	"os/exec"
	"regexp"
	"strings"

	"github.com/kissanjamgit/ext"
	"resty.dev/v3"
)

type devilsfilmAPIEnv struct {
	API struct {
		Algolia struct {
			AppID  string `json:"applicationID"`
			APIKey string `json:"apiKey"`
		} `json:"algolia"`
	} `json:"api"`
}

type DevilsFilm struct {
	Source string
}

func New(Source string) ext.Site {
	return &DevilsFilm{Source: Source}
}

func form(source string) (str string, err error) {
	re := regexp.MustCompile(`video/([^/]+)\S+/([0-9]+)`)
	var available string
	var clipID string
	if submatch := re.FindStringSubmatch(source); len(submatch) < 3 {
		err = fmt.Errorf("len(submatch) <3 len(submatch): %d", len(submatch))
		return
	} else {
		available = submatch[1]
		clipID = submatch[2]
	}
	str = `{"requests":[{"indexName":"all_scenes","params":"clickAnalytics=true&facetFilters=%5B%5B%22availableOnSite%3A` + available + "%22%2C%22availableOnSite%3Asquirtalicious%22%5D%2C%5B%22clip_id%3A" + clipID + `%22%5D%5D&facets=%5B%5D&hitsPerPage=1&tagFilters="},{"indexName":"all_scenes","params":"analytics=false&clickAnalytics=false&facetFilters=%5B%5B%22clip_id%3A` + clipID + `%22%5D%5D&facets=availableOnSite&hitsPerPage=0&page=0"},{"indexName":"all_scenes","params":"analytics=false&clickAnalytics=false&facetFilters=%5B%5B%22availableOnSite%3A` + available + `%22%2C%2222%5D%5D&facets=clip_id&hitsPerPage=0&page=0"}]}`
	return
}

func (d *DevilsFilm) Resource(client *resty.Client) (cr ext.ContentResource, err error) {
	Name, err := title(d.Source)
	if err != nil {
		return
	}
	res, err := client.SetHeaders(ext.Header).R().Get(d.Source)
	if err != nil {
		return
	}
	var match []byte
	if submatch := regexp.MustCompile(`window.env\s+=\s+([^;]+)`).FindSubmatch(res.Bytes()); len(submatch) < 2 {
		err = fmt.Errorf("len(match) < 2, len(match): %d", len(submatch))
		return
	} else {
		match = submatch[1]
	}
	var apiEnv devilsfilmAPIEnv
	err = json.Unmarshal(match, &apiEnv)
	if err != nil {
		return
	}

	header, err := postHeader(d.Source, apiEnv)
	if err != nil {
		return
	}
	data, err := form(d.Source)
	if err != nil {
		return
	}
	res, err = client.R().SetHeaders(header).SetBody(data).Post(
		fmt.Sprintf(
			"https://%s-dsn.%s.net/1/indexes/*/queries?%s",
			strings.ToLower(apiEnv.API.Algolia.AppID),
			"algolia",
			url.QueryEscape("x-algolia-agent=Algolia for JavaScript (4.22.1); Browser; JS Helper (3.16.3)"),
		))
	if err != nil {
		return
	}
	URL, err := quality(res.String(), "1080p")
	if err != nil {
		return
	}
	cr = ext.ContentResource{
		URL:  URL,
		Name: Name,
	}
	return
}

func postHeader(uri string, apiEnv devilsfilmAPIEnv) (header map[string]string, err error) {
	URL, err := url.Parse(uri)
	if err != nil {
		return
	}
	host := fmt.Sprintf("https://%s/", URL.Hostname())
	header = maps.Clone(ext.Header)
	header["x-algolia-api-key"] = apiEnv.API.Algolia.APIKey
	header["x-algolia-application-id"] = apiEnv.API.Algolia.AppID
	header["content-type"] = "application/x-www-form-urlencoded"
	header["Referer"] = host
	header["Origin"] = host
	return
}

func quality(raw string, queryEqualToORLessThen string) (url string, err error) {
	var quality struct {
		Trailers map[string]string `json:"trailers"`
	}
	match := regexp.MustCompile(`"trailers":{[^}]+}`).FindString(raw)
	if match == "" {
		err = fmt.Errorf("match == `` ")
		return
	}
	err = json.Unmarshal([]byte("{"+match+"}"), &quality)
	if err != nil {
		return
	}
	i := len(ext.Quality) - 1
	for ; i > 0; i-- {
		if ext.Quality[i] == queryEqualToORLessThen {
			break
		}
	}
	err = fmt.Errorf("for the query item doesn't exsist which is equal or less then ")
	for ; i > 0; i-- {
		url = quality.Trailers[ext.Quality[i]]
		if url == "" {
			continue
		}
		return url, nil
	}

	return
}

func title(url string) (title string, err error) {
	if submatch := regexp.MustCompile("/([^/]+)/[0-9]+").FindStringSubmatch(url); len(submatch) < 2 {
		err = fmt.Errorf("len(submatch) <2, len(submatch): %d", len(submatch))
		return
	} else {
		title = submatch[1]
	}
	title = regexp.MustCompile(`[-]+`).ReplaceAllString(title, " ")
	return
}

func (*DevilsFilm) Download(cr ext.ContentResource) (err error) {
	cmd := exec.Command("yt-dlp", cr.URL, "-o", cr.Name)
	err = cmd.Run()
	return
}
