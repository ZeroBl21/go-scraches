package main

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday/v2"
)

const defaultTemplate = `
<!DOCTYPE html>
<html>
  <head>
    <meta http-equiv="content-type" content="text/html; charset=utf-8">
    <title>{{.Title}}</title>
  </head>
  <body>
	{{.Body}}
  </body>
</html>`

type content struct {
	Title string
	Body  template.HTML
}

func main() {
	filename := flag.String("file", "", "Markdown file to preview")
	preview := flag.Bool("preview", false, "Auto preview the file")
	tName := flag.String("t", "", "Alternative template name")
	flag.Parse()

	if *filename == "" {
		flag.Usage()
		return
	}

	if err := run(*filename, *tName, os.Stdout, *preview); err != nil {
		log.Fatal(err)
	}
}

func run(filename, tName string, out io.Writer, preview bool) error {
	input, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	htmlData, err := parseContent(input, tName)
	if err != nil {
		return err
	}

	temp, err := os.CreateTemp("", "mdp*.html")
	if err != nil {
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}

	outName := temp.Name()

	fmt.Fprintln(out, outName)

	if err := saveHTML(outName, htmlData); err != nil {
		return err
	}

	if !preview {
		return nil
	}

	defer os.Remove(outName)

	return previewFile(outName)
}

func parseContent(input []byte, tName string) ([]byte, error) {
	output := blackfriday.Run(input)
	body := bluemonday.UGCPolicy().SanitizeBytes(output)

	templ, err := template.New("mdp").Parse(defaultTemplate)
	if err != nil {
		return nil, err
	}

	if tName != "" {
		templ, err = template.ParseFiles(tName)
		if err != nil {
			return nil, err
		}
	}

	c := content{
		Title: "Markdown Preview Tool",
		Body:  template.HTML(body),
	}

	var buffer bytes.Buffer

	if err := templ.Execute(&buffer, c); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func saveHTML(outName string, data []byte) error {
	return os.WriteFile(outName, data, 0644)
}

func previewFile(fname string) error {
	cName := ""
	cParams := []string{}

	switch runtime.GOOS {
	case "linux":
		cName = "xdg-open"
	case "windows":
		cName = "cmd.exe"
	case "darwin":
		cName = "open"
	default:
		return fmt.Errorf("OS not supportedcmd.exe")
	}

	cParams = append(cParams, fname)

	cPath, err := exec.LookPath(cName)
	if err != nil {
		return err
	}

	err = exec.Command(cPath, cParams...).Run()

	time.Sleep(2 * time.Second)

	return err
}
