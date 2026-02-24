// Package blacked is a package for getting the
package blacked

import (
	"fmt"
	"io"
	"maps"
	"net/url"
	"os"
	"os/exec"
	"regexp"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
	"github.com/kissanjamgit/ext"
	"resty.dev/v3"
)

type Blacked struct {
	Source string
}

func New(Source string) ext.Site {
	return &Blacked{Source: Source}
}

func getBody(videoID string) map[string]any {
	body := map[string]any{
		"operationName": "getTokenTour",
		"query": `query getTokenTour($videoId: ID!, $device: Device!) {
        generateVideoToken(input: {videoId: $videoId, device: $device}) {
            p270 { token cdn __typename }
            p360 { token cdn __typename }
            p480 { token cdn __typename }
            p480l { token cdn __typename }
            p720 { token cdn __typename }
            p1080 { token cdn __typename }
            p2160 { token cdn __typename }
            hls { token cdn __typename }
            __typename
        }
    }`,
		"variables": map[string]any{
			"videoId": videoID,
			"device":  "trailer",
		},
	}
	return body
}

func getHeader(uri *url.URL) map[string]string {
	header := maps.Clone(ext.Header)
	header["Referer"] = uri.String()
	header["Origin"] = fmt.Sprintf("%s://%s", uri.Scheme, uri.Hostname())
	header["Cookie"] = "consent=0; nats=NjI3LjYxLjE1LjUxLjAuMC4wLjAuMA; nats_cookie=No%2BReferring%2BURL; nats_unique=NjI3LjYxLjE1LjUxLjAuMC4wLjAuMA; nats_sess=5c23e5d65e2e21e9803d498da8774913; nats_landing=No%2BLanding%2BPage%2BURL; vuid=e4aea734-ac27-4f3b-816d-4d301d8cd4f9; last_video_performer_page_viewed=video%3Amarried-redhead-lauren-gets-dped"
	return header
	// return map[string]string{
	// 	"User-Agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:143.0) Gecko/20100101 Firefox/143.0",
	// 	"Accept":          "application/json",
	// 	"Accept-Language": "en-US,en;q=0.5",
	// 	"Referer":         uri.String(),
	// 	"content-type":    "application/json",
	// 	"Origin":          fmt.Sprintf("%s://%s", uri.Scheme, uri.Hostname()),
	// 	"Sec-GPC":         "1",
	// 	"Connection":      "keep-alive",
	// 	"Cookie":          "consent=0; nats=NjI3LjYxLjE1LjUxLjAuMC4wLjAuMA; nats_cookie=No%2BReferring%2BURL; nats_unique=NjI3LjYxLjE1LjUxLjAuMC4wLjAuMA; nats_sess=5c23e5d65e2e21e9803d498da8774913; nats_landing=No%2BLanding%2BPage%2BURL; vuid=e4aea734-ac27-4f3b-816d-4d301d8cd4f9; last_video_performer_page_viewed=video%3Amarried-redhead-lauren-gets-dped",
	// 	"Sec-Fetch-Dest":  "empty",
	// 	"Sec-Fetch-Mode":  "cors",
	// 	"Sec-Fetch-Site":  "same-origin",
	// 	"Priority":        "u=4",
	// 	"TE":              "trailers",
	// }
}

func getVideoMetaData(uri *url.URL, _ *resty.Client) (id string, title string, err error) {
	options := []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(30),
		tls_client.WithClientProfile(profiles.Chrome_144),
		tls_client.WithNotFollowRedirects(),
	}
	c, err := tls_client.NewHttpClient(tls_client.NewLogger(), options...)
	if err != nil {
		fmt.Println(err)
		return
	}

	req, err := http.NewRequest(http.MethodGet, uri.String(), nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	req.Header = http.Header{
		"accept":          {},
		"accept-language": {"de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7"},
		"user-agent":      {"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36"},
		http.HeaderOrderKey: {
			"accept",
			"accept-language",
			"user-agent",
		},
	}
	r, err := c.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}

	var data []byte
	data, err = io.ReadAll(r.Body)
	os.WriteFile("content.html", data, 0o644)
	// panic("unimplemented")
	// res, err := client.R().Get(uri.String())
	// if err != nil {
	// 	return
	// }
	// data = res.String()
	regVideoID := regexp.MustCompile(`"videoId":"([^"]+)"`)
	subVideoID := regVideoID.FindStringSubmatch(string(data))
	if len(subVideoID) < 1 {
		err = fmt.Errorf("(getVideoID) `len(subVideoID)<1` subVideoID: %v", subVideoID)
		return
	}
	id = subVideoID[1]

	regTitle := regexp.MustCompile(`"title":"([^"]+)"`)
	subTitle := regTitle.FindStringSubmatch(string(data))
	if len(subTitle) < 1 {
		err = fmt.Errorf("(getVideoMetaData) `len(title)<1` title: %v", subTitle)
		return
	}
	title = subTitle[1]

	return
}

func (b *Blacked) Resource(client *resty.Client) (cr ext.ContentResource, err error) {
	uri, err := url.Parse(b.Source)
	if err != nil {
		return
	}
	videoID, title, err := getVideoMetaData(uri, client)
	if err != nil {
		return
	}
	res, err := client.R().SetHeaders(getHeader(uri)).SetBody(getBody(videoID)).Post(
		fmt.Sprintf("%s://%s/graphql", uri.Scheme, uri.Hostname()))
	if err != nil {
		return
	}
	reg := regexp.MustCompile(`"p1080":{"token":"([^"]+)"`)

	sub := reg.FindStringSubmatch(res.String())
	if len(sub) < 1 {
		err = fmt.Errorf("(Video) len(sub)<1, sub: %v", sub)
		return
	}
	video := sub[1]
	if video == "" || title == "" {
		err = fmt.Errorf(`(Video) 'video == "" || title == ""' video: %s, title: %s`, video, title)
	}
	cr = ext.ContentResource{
		Name: title,
		URL:  video,
	}
	return
}

func (*Blacked) Download(cr ext.ContentResource) (err error) {
	cmd := exec.Command("yt-dlp", cr.URL, "-o", cr.Name+".mp4")
	err = cmd.Run()
	return
}
