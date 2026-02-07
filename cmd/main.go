package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/kissanjamgit/ext"
	"github.com/kissanjamgit/ext/advd"
	"github.com/kissanjamgit/ext/blacked"
	"github.com/kissanjamgit/ext/devilsfilm"
	"github.com/kissanjamgit/ext/kk"
	"github.com/kissanjamgit/ext/lulustream"
	"github.com/kissanjamgit/ext/pornbox"
	"github.com/kissanjamgit/ext/savefiles"
	"github.com/kissanjamgit/ext/streamtape"
	"github.com/kissanjamgit/ext/strmup"
	"github.com/kissanjamgit/ext/vidnest"
	"resty.dev/v3"
)

var convergTree = map[string]func(string) ext.Site{
	"lulustream": lulustream.New,
	"luluvid":    lulustream.New,
	"savefiles":  savefiles.New,
	"vidnest":    vidnest.New,
	"streamtape": streamtape.New,
	"strmup":     strmup.New,

	"devilsfilm":         devilsfilm.New,
	"puretaboo":          devilsfilm.New,
	"mommysboy":          devilsfilm.New,
	"mommysgirl":         devilsfilm.New,
	"girlsway":           devilsfilm.New,
	"21sextury":          devilsfilm.New,
	"outofthefamily":     devilsfilm.New,
	"moderndaysins":      devilsfilm.New,
	"evilangel":          devilsfilm.New,
	"tabooheat":          devilsfilm.New,
	"accidentalgangbang": devilsfilm.New,

	"kink":           kk.New,
	"adultdvdempire": advd.New,

	"blacked":    blacked.New,
	"tushy":      blacked.New,
	"vixen":      blacked.New,
	"blackedraw": blacked.New,
	"tushyraw":   blacked.New,
	"deeper":     blacked.New,
	"slayed":     blacked.New,
	"milfy":      blacked.New,
	"wifey":      blacked.New,

	"pornbox":  pornbox.New,
	"analvids": pornbox.New,
	"pissvids": pornbox.New,
}

func domainOnly(raw string) (str string, err error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", err
	}

	host := strings.ToLower(u.Hostname()) // strips port automatically

	index := []int{}
	for i, s := range host {
		if s != '.' {
			continue
		}
		index = append(index, i)
	}
	if l := len(index); l != 1 && l != 2 {
		err = fmt.Errorf("len(index) != 1 or 2")
		return
	} else {
		switch l {
		case 1:
			str = host[:index[0]]
			return
		case 2:
			str = host[index[0]+1 : index[1]]
			return
		}
	}

	return
}

func handle(item string) (s ext.Site, err error) {
	domain, err := domainOnly(item)
	if err != nil {
		return
	}
	sfunc, ok := convergTree[domain]
	if !ok {
		sfunc = ext.NewSiteAlter
	}
	s = sfunc(item)
	return
}

func httpSplit(str string) []string {
	str = strings.ReplaceAll(str, "http", " http")
	str = strings.TrimSpace(str)
	value := strings.Split(str, " ")
	return value
}

type inputError struct {
	input string
	err   error
}

func prtErrorList(list []inputError) {
	if len(list) == 0 {
		return
	}
	var buf strings.Builder
	for i, item := range list {
		fmt.Fprintf(&buf, "%d: %s\n", i, item.err)
	}
	fmt.Fprint(os.Stderr, buf.String())
}

func main() {
	inputArg := flag.String("i", "", "input")
	dragNDrop := flag.Bool("dd", false, "drag and drop")
	download := flag.Bool("d", false, "download")
	clipboard := flag.Bool("c", false, "copy from clipboard")
	flag.Parse()

	argExclution := 0
	for _, i := range []bool{*dragNDrop, *download, *clipboard} {
		if !i {
			continue
		}
		argExclution++
	}
	if argExclution > 1 {
		err := fmt.Errorf("only one can be selected at an instance dragNDrop, download and clipboard")
		panic(err)
	}

	var input []string

	if *inputArg != "" {
		input = httpSplit(*inputArg)
	} else if *dragNDrop {
		reader := bufio.NewReader(os.Stdin)
		str, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}
		input = httpSplit(str)
	}
	var inputErrorList []inputError

	client := resty.New()
	defer client.Close()
	if *download {
		for _, item := range input {
			domain, err := domainOnly(item)
			if err != nil {
				inputErrorList = append(inputErrorList, inputError{input: item, err: err})
				continue
			}
			sfunc, ok := convergTree[domain]
			if !ok {
				sfunc = ext.NewSiteAlter
			}
			site := sfunc(item)
			cr, err := site.Resource(client)
			if err != nil {
				inputErrorList = append(inputErrorList, inputError{input: item, err: err})
				continue
			}
			err = site.Download(cr)
			if err != nil {
				inputErrorList = append(inputErrorList, inputError{input: item, err: err})
				continue
			}

		}
		prtErrorList(inputErrorList)
		return
	}
	CR := make(chan ext.ContentResource)

	go func() {
		semaphore := make(chan struct{}, 10)
		wg := sync.WaitGroup{}
		for _, item := range input {
			semaphore <- struct{}{}
			s, err := handle(item)
			if err != nil {
				inputErrorList = append(inputErrorList, inputError{input: item, err: err})
				<-semaphore
				continue
			}

			wg.Go(func() {
				defer func() {
					<-semaphore
				}()

				cr, err := s.Resource(client)
				if err != nil {
					inputErrorList = append(inputErrorList, inputError{input: item, err: err})
					return
				}
				CR <- cr
			},
			)
		}
		close(semaphore)

		wg.Wait()
		close(CR)
	}()

	cr := <-CR
	fmt.Printf("#EXTM3U\n#EXTINF:-1,%s\n%s\n", cr.Name, cr.URL)
	for cr = range CR {
		fmt.Printf("#EXTINF:-1,%s\n%s\n", cr.Name, cr.URL)
	}
	prtErrorList(inputErrorList)
}
