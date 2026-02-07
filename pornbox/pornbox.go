// Package pornbox provides a wrapper for pornbox.com
package pornbox

import (
	"fmt"
	"maps"
	"os/exec"
	"regexp"

	"github.com/kissanjamgit/ext"
	"github.com/tidwall/gjson"
	"resty.dev/v3"
)

// var header = map[string]string{
// 	"User-Agent":       "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:143.0) Gecko/20100101 Firefox/143.0",
// 	"Accept":           "*/*",
// 	"Accept-Language":  "en-US,en;q=0.5",
// 	"Accept-Encoding":  "gzip, deflate, br, zstd",
// 	"X-CSRF-Token":     "F6wyQTnF-rZeY2u7ZV8fUaY0dslunx1GQySQ",
// 	"X-Requested-With": "XMLHttpRequest",
// 	"Sec-GPC":          "1",
// 	"Connection":       "keep-alive",
// 	"Cookie":           "JDIALOG3=FIZT353D78913CYVAYLIVSOV4SGQM6MG516ZCS3SA5QEMVHWOA; http_referer=; entry_point=https%3A%2F%2Fpornbox.com%2Fapplication%2Fwatch-page%2F3842321; boxsessid=s%3APQIiI9skPgvgznr2YhxuACHu2WwahlCg.r1a4qD658kvqUWwtFlrGmTwWss7UvuwuR7%2BNUtkdfRo; version_website_id=j%3A%5B25%5D; agree18=1",
// 	"Sec-Fetch-Dest":   "empty",
// 	"Sec-Fetch-Mode":   "cors",
// 	"Sec-Fetch-Site":   "same-origin",
// }

type PornBox struct {
	Source string
}

func New(Source string) ext.Site {
	return &PornBox{Source}
}

func getHeader() map[string]string {
	header := maps.Clone(ext.Header)
	header["X-CSRF-Token"] = "F6wyQTnF-rZeY2u7ZV8fUaY0dslunx1GQySQ"
	header["Cookie"] = "JDIALOG3=FIZT353D78913CYVAYLIVSOV4SGQM6MG516ZCS3SA5QEMVHWOA; http_referer=; entry_point=https%3A%2F%2Fpornbox.com%2Fapplication%2Fwatch-page%2F3842321; boxsessid=s%3APQIiI9skPgvgznr2YhxuACHu2WwahlCg.r1a4qD658kvqUWwtFlrGmTwWss7UvuwuR7%2BNUtkdfRo; version_website_id=j%3A%5B25%5D; agree18=1"
	return header
}

// func New(Source string) ext.Site {
// 	return &PornBox{Source}
// }

func title(json string) (name string, err error) {
	value := gjson.Get(json, "scene_name").Value()
	var ok bool
	if name, ok = value.(string); !ok {
		err = fmt.Errorf("pornBox: (_name) `value.(string)` value: %v", value)
		return
	}
	return
}

func (p *PornBox) Resource(client *resty.Client) (cr ext.ContentResource, err error) {
	re := regexp.MustCompile(`\d+`)
	id := re.FindString(p.Source)
	url := fmt.Sprintf("https://pornbox.com/contents/%v", id)
	client.SetHeaders(getHeader())
	var res *resty.Response
	res, err = client.R().Get(url)
	if err != nil {
		return
	}
	value := gjson.Get(res.String(), "medias.@reverse.#(title==Trailer).media_id")
	mediaID, ok := value.Value().(float64)
	if !ok {
		err = fmt.Errorf("pornBox: (video) `value.(float64)` value: %v ", value)
		return
	}
	// fmt.Printf("mediaID: %d\n", int(mediaID))
	var getSrc *resty.Response
	getSrc, err = client.R().Get(fmt.Sprintf("https://pornbox.com/media/%d/stream", int(mediaID)))
	if err != nil {
		return
	}

	value = gjson.Get(getSrc.String(), `qualities.@reverse.#(size==1080p).src`)
	var src string
	src, ok = value.Value().(string)
	if !ok {
		err = fmt.Errorf("pornBox: (video) `value.(string)` value: %v ", value)
		return
	}
	var name string
	name, err = title(res.String())
	if err != nil {
		return
	}
	cr = ext.ContentResource{Name: name, URL: src}
	return
}

func (*PornBox) Download(cr ext.ContentResource) (err error) {
	cmd := exec.Command("yt-dlp", cr.URL, "-o", cr.Name)
	err = cmd.Run()
	return
}

// func Queue(input []string) (result []ext.ContentResource, errorList []string,
// ) {
// 	client := resty.New()
// 	defer client.Close()
// 	for _, url := range input {
// 		cr, err := video(url, client)
// 		if err != nil {
// 			errorList = append(errorList, fmt.Sprintf("ERROR: %s (%v)", url, err))
// 			continue
// 		}
// 		result = append(result, cr)
// 	}
// 	return
// }
