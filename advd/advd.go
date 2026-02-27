// Package advd provides a wrapper for adultempire.com
package advd

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/kissanjamgit/ext"
	"resty.dev/v3"
)

func cleanTitle(str string) string {
	str = strings.ReplaceAll(str, "-", " ")
	str = strings.TrimRight(str, ".html")
	return str
}

type Advd struct {
	Source string
}

func New(source string) ext.Site {
	return &Advd{Source: source}
}

func clip(a *Advd, client *resty.Client, ID string) (cr ext.ContentResource, err error) {
	res, err := client.R().SetHeader("Cookie", "ageConfirmed=true; defaults={}").Get(a.Source)
	if err != nil {
		return
	}

	// submatch := regexp.MustCompile(`<link rel="alternate" type="[^"]" title="([^"]+)" href="([^"])"`).FindStringSubmatch(res.String())
	os.WriteFile("content.html", res.Bytes(), 0o644)
	submatch := regexp.MustCompile(`<h1 class="clip-page__detail__title__primary">([^<]*)</h1>`).FindStringSubmatch(res.String())
	fmt.Println(submatch)
	if len(submatch) < 2 {
		err = fmt.Errorf("(clip) `len(submatch) < 3`")
		return
	}
	name := submatch[1]
	submatch = regexp.MustCompile(`<iframe frameborder="0"[^>]+src="([^"]*)"`).FindStringSubmatch(res.String())
	if len(submatch) < 2 {
		err = fmt.Errorf("(clip) `len(submatch) < 2`")
		return
	}
	uri, err := url.Parse(submatch[1])
	if err != nil {
		return
	}
	movieID := uri.Query().Get("item_id")
	if movieID == "" {
		err = fmt.Errorf("(clip) `movieID == ''`")
		return
	}
	cr = ext.ContentResource{Name: name, URL: fmt.Sprintf("https://internal-video.adultempire.com/hls/previewscene/%s/%s/index-f2-v1-a1.m3u8", movieID, ID)}
	return
}

func movie(a *Advd, client *resty.Client, ID string) (cr ext.ContentResource, err error) {
	res, err := client.R().SetHeader("Cookie", "ageConfirmed=true; defaults={}").Get(a.Source)
	if err != nil {
		return
	}
	submatch := regexp.MustCompile(`<h1[^>]+>([^<]*)</h1>`).FindSubmatch(res.Bytes())
	if len(submatch) < 2 {
		err = fmt.Errorf("(movie) `len(submatch) < 2`")
		return
	}
	name := strings.TrimSpace(string(submatch[1]))
	if submatch := regexp.MustCompile(`Label="Studio">([^<]+)</a>`).FindStringSubmatch(res.String()); len(submatch) == 2 {
		name = name + " - " + submatch[1]
	}
	cr = ext.ContentResource{Name: name, URL: fmt.Sprintf("https://trailer.adultempire.com/hls/preview/%s/master.m3u8", ID)}
	return
}

func (a *Advd) Resource(client *resty.Client) (cr ext.ContentResource, err error) {
	if submatch := regexp.MustCompile("/clip/([0-9]+)/").FindStringSubmatch(a.Source); len(submatch) > 0 {
		if len(submatch) < 2 {
			err = fmt.Errorf("(Resource) `len(submatch) < 2`")
			return
		}
		cr, err = clip(a, client, submatch[1])
		return
	} else if match := regexp.MustCompile("[0-9]+").FindString(a.Source); match != "" {
		cr, err = movie(a, client, match)
		return
	}
	err = fmt.Errorf("if Matched Failed")
	return

	// regID := regexp.MustCompile("[0-9]+")
	// ID := regID.FindString(a.Source)
	// regTitle := regexp.MustCompile("[^/]+.html")
	// title := regTitle.FindString(a.Source)
	// title = cleanTitle(title)
	// hls := fmt.Sprintf("https://trailer.adultempire.com/hls/preview/%s/master.m3u8", ID)
	// cr = ext.ContentResource{Name: title, URL: hls}
}

// func clean(str string) []string {
// 	str = strings.TrimSpace(str)
// 	reg := regexp.MustCompile("\n| +")
// 	str = reg.ReplaceAllString(str, " ")
// 	return strings.Split(str, " ")
// }

// func main() {
// 	inputArgs := flag.String("i", "", "")
// 	clipboardArg := flag.Bool("c", false, "")
// 	download := flag.Bool("d", false, "")
// 	flag.Parse()
//
// 	if *inputArgs == "" && !*clipboardArg {
// 		fmt.Println(`*inputArgs == ""`)
// 		os.Exit(1)
// 	}
//
// 	if *clipboardArg {
// 		value, err := clipboard.ReadAll()
// 		if err != nil {
// 			os.Exit(1)
// 		}
// 		inputArgs = &value
// 	}
// 	input := clean(*inputArgs)
// 	var errorList []string
// 	var res []ContentResource
// 	var buff strings.Builder
// 	buff.WriteString("#EXTM3U\n")
// 	for _, url := range input {
// 		res = append(res, getContentResource(url))
// 	}
// 	if !*download {
// 		for _, cr := range res {
// 			fmt.Fprintf(&buff, "#EXTINF:-1,%s\n%s\n", cr.Name, cr.URL)
// 		}
// 		fmt.Println(buff.String())
// 	} else {
// 		for _, cr := range res {
// 			cmdString := fmt.Sprintf(`ext.py -i "%s" -o "%s"`, cr.URL, cr.Name)
// 			cmd := exec.Command("fish.exe", "-c", cmdString)
// 			cmd.Stdout = os.Stdout
// 			cmd.Stderr = os.Stderr
// 			err := cmd.Run()
// 			if err == nil {
// 				continue
// 			}
// 			errorList = append(errorList, cr.URL)
// 		}
// 	}
// 	if len(errorList) == 0 && len(res) > 0 {
// 		os.Exit(0)
// 	}
// 	var errorHls strings.Builder
// 	for _, hls := range errorList {
// 		fmt.Fprintf(&errorHls, " %s\n", hls)
// 	}
// 	fmt.Printf("ERROR:\n%s", errorHls.String())
// 	os.Exit(1)
// }

func (*Advd) Download(cr ext.ContentResource) (err error) {
	cmd := exec.Command("yt-dlp", cr.URL, "-o", cr.Name+".mp4")
	err = cmd.Run()
	return
}
