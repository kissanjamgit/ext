// Package kk provides a wrapper for kink.com
package kk

import (
	"encoding/json"
	"flag"
	"fmt"
	"maps"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	// "github.com/atotto/clipboard"
	"github.com/kissanjamgit/ext"
	"resty.dev/v3"
)

// var header = map[string]string{
// 	"User-Agent":                "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:143.0) Gecko/20100101 Firefox/143.0",
// 	"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
// 	"Accept-Language":           "en-US,en;q=0.5",
// 	"Sec-GPC":                   "1",
// 	"Connection":                "keep-alive",
// 	"Upgrade-Insecure-Requests": "1",
// 	"Sec-Fetch-Dest":            "document",
// 	"Sec-Fetch-Mode":            "navigate",
// 	"Sec-Fetch-Site":            "same-origin",
// 	"Priority":                  "u=0, i",
// 	"TE":                        "trailers",
// }

type KK struct {
	Source string
}

func New(source string) ext.Site {
	return &KK{Source: source}
}

func clean(str string) []string {
	str = strings.TrimSpace(str)
	reg := regexp.MustCompile("\n| +")
	str = reg.ReplaceAllString(str, " ")
	return strings.Split(str, " ")
}

type jsonValue struct {
	ContentURL string `json:"contentUrl"`
	Title      string `json:"name"`
	UploadDate string `json:"uploadDate"`
}

func (k *KK) Resource(client *resty.Client) (cr ext.ContentResource, err error) {
	client.SetHeaders(ext.Header)
	res, err := client.R().Get(k.Source)
	if err != nil {
		fmt.Println(err)
		return
	}
	reg := regexp.MustCompile(`<script type="application/ld\+json">(.*?)</script>`)
	valu := reg.FindStringSubmatch(res.String())
	if len(valu) < 2 {
		err = fmt.Errorf("len(valu) < 2")
		return
	}
	// var jsonValue map[string]any
	var jsonValue jsonValue
	json.Unmarshal([]byte(valu[1]), &jsonValue)
	t, err := time.Parse(time.RFC3339Nano, "2011-07-27T19:00:00.000Z")
	if err != nil {
		return
	}
	timeString := t.Format("Jan 02 2006")
	Name := fmt.Sprintf("Kink - %s - %s", jsonValue.Title, timeString)
	cr = ext.ContentResource{Name: Name, URL: jsonValue.ContentURL}
	return
}

func main() {
	inputArgs := flag.String("i", "", "")
	clipboardArg := flag.Bool("c", false, "")
	cookieArg := flag.String("cookie", "", "")
	// download := flag.Bool("d", false, "")
	flag.Parse()

	if *inputArgs == "" && !*clipboardArg {
		fmt.Println(`*inputArgs == "" && !*clipboard `)
		os.Exit(1)
	}
	// if *clipboardArg {
	// 	value, err := clipboard.ReadAll()
	// 	if err != nil {
	// 		os.Exit(1)
	// 	}
	// 	inputArgs = &value
	// }

	header := maps.Clone(ext.Header)
	if *cookieArg != "" {
		header["Cookie"] = *cookieArg
	}

	// input := clean(*inputArgs)
	// var errorList []string
	// var res []ext.ContentResource
	// var buff strings.Builder

	// client := resty.New()
	// defer client.Close()
	// client.SetHeaders(header)
	// buff.WriteString("#EXTM3U\n")
	// for _, url := range input {
	// 	cr, err := getContentResource(url, client)
	// 	if err != nil {
	// 		errorList = append(errorList, err.Error())
	// 		continue
	// 	}
	// 	res = append(res, cr)
	// }
	// if !*download {
	// 	for _, cr := range res {
	// 		fmt.Fprintf(&buff, "#EXTINF:-1,%s\n%s\n", cr.Name, cr.URL)
	// 	}
	// 	fmt.Println(buff.String())
	// } else {
	//
	// for _, cr := range res {
	// 			cmdString := fmt.Sprintf(`ext.py -i "%s" -o "%s"`, cr.URL, cr.Name)
	// 			cmd := exec.Command("fish.exe", "-c", cmdString)
	// 			cmd.Stdout = os.Stdout
	// 			cmd.Stderr = os.Stderr
	// 			err := cmd.Run()
	// 			if err == nil {
	// 				continue
	// 			}
	// 	errorList = append(errorList, cr.URL)
	// }
	// }
	// if len(errorList) == 0 && len(res) > 0 {
	// 	os.Exit(0)
	// }
	// var errorHls strings.Builder
	// for _, hls := range errorList {
	// 	fmt.Fprintf(&errorHls, " %s\n", hls)
	// }
	// fmt.Printf("ERROR:\n%s", errorHls.String())
	// os.Exit(1)
}

func (*KK) Download(cr ext.ContentResource) (err error) {
	cmd := exec.Command("yt-dlp.exe", "-i", cr.URL, "-o", cr.Name)
	err = cmd.Run()
	return
}
