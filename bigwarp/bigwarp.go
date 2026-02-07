package bigwarp

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	. "github.com/kissanjamgit/ext"
)

type BigWarp struct {
	Source string
}

func (b *BigWarp) Resource() (cr ContentResource, err error) {
	var headerList []string
	for k, v := range Header {
		headerList = append(headerList, []string{"--add-header", fmt.Sprintf("%s: %s", k, v)}...)
	}
	cmd := exec.Command("yt-dlp.exe", b.Source, "--extractor-args", `generic:impersonate`, "--print", PrtFmt)
	cmd.Args = append(cmd.Args, headerList...)
	var yt YtdlpResource
	cmd.Stderr = os.Stderr
	output, err := cmd.Output()
	if err != nil {
		return
	}
	fmt.Println(string(output))
	err = json.Unmarshal(output, &yt)
	if err != nil {
		return
	}
	cr = yt.ToContentResource()
	return
}

func (b *BigWarp) Download(cr ContentResource) (err error) {
	cmd := exec.Command("yt-dlp.exe", cr.URL, "-o", cr.Name)
	err = cmd.Run()
	return
}
