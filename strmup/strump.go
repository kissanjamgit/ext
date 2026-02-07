// Package strmup provides a wrapper for strmup.to
package strmup

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"reflect"
	"regexp"

	"github.com/kissanjamgit/ext"
	"resty.dev/v3"
)

type Strmup struct {
	Source string
}

func New(Source string) ext.Site {
	return &Strmup{Source: Source}
}

func (s *Strmup) Resource(client *resty.Client) (cr ext.ContentResource, err error) {
	id := regexp.MustCompile("[^/]+$").FindString(s.Source)
	return videoByID(id)
}

func videoByID(id string) (cr ext.ContentResource, err error) {
	client := resty.New()
	defer client.Close()
	url_ := fmt.Sprintf("https://strmup.to/ajax/stream?filecode=%s", id)
	resp, err := client.R().Get(url_)
	if err != nil {
		fmt.Println(err)
		return
	}
	var value any
	err = json.Unmarshal(resp.Bytes(), &value)
	if err != nil {
		return
	}
	map_, ok := value.(map[string]any)
	if !ok {
		err = fmt.Errorf("value is't map, type: %s, value: %v", reflect.TypeOf(value).String(), value)
		return
	}
	url, ok := map_["streaming_url"].(string)
	if !ok {
		err = fmt.Errorf("url.TypeOf != string value: %v", value)
		return
	}
	title, ok := map_["title"].(string)
	if !ok {
		err = fmt.Errorf("title.TypeOf != string value: %v", value)
		return
	}

	if title == "" {
		err = fmt.Errorf("title is empty")
		return
	}
	cr = ext.ContentResource{
		Name: title,
		URL:  url,
	}
	return
}

func (*Strmup) Download(cr ext.ContentResource) (err error) {
	cmd := exec.Command("yt-dlp", cr.URL, "-o", cr.Name)
	err = cmd.Run()
	return
}

// func main() {
// 	id := flag.String("id", "", "")
// 	url := flag.String("i", "", "")
// 	c := flag.Bool("c", false, "")
// 	name := flag.Bool("s", false, "")
//
// 	flag.Parse()
//
// 	var cr ext.ContentResource
// 	var err error
// 	go func() {
// 		if *id != "" {
// 			cr, err = videoByID(*id)
// 			if err != nil {
// 				fmt.Println(err)
// 				os.Exit(1)
// 				return
// 			}
// 			return
// 		}
// 		var strmup Strmup
// 		if *url != "" {
// 			strmup = New(*url)
// 			cr, err = strmup.Resource()
// 			if err != nil {
// 				fmt.Println(err)
// 				os.Exit(1)
// 				return
// 			}
// 			return
// 		}
// 		if *c {
// 			var str string
// 			str, err = clipboard.ReadAll()
// 			if err != nil {
// 				fmt.Println(err)
// 				os.Exit(1)
// 				return
// 			}
// 			strmup = New(str)
// 			cr, err = strmup.()
// 			if err != nil {
// 				fmt.Println(err)
// 				os.Exit(1)
// 				return
// 			}
// 			return
//
// 		}
// 	}()
//
// 	if *name {
// 		fmt.Printf("%s|%s", cr.Name, cr.URL)
// 		return
// 	}
// 	fmt.Println(cr.URL)
// }
