// Package advd provides a wrapper for adultempire.com
package advd

import (
	"fmt"
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

func (a *Advd) Resource(_ *resty.Client) (cr ext.ContentResource, err error) {
	regID := regexp.MustCompile("[0-9]+")
	ID := regID.FindString(a.Source)
	regTitle := regexp.MustCompile("[^/]+.html")
	title := regTitle.FindString(a.Source)
	title = cleanTitle(title)
	hls := fmt.Sprintf("https://trailer.adultempire.com/hls/preview/%s/master.m3u8", ID)
	cr = ext.ContentResource{Name: title, URL: hls}
	return
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
