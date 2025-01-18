package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday/v2"
)

const header = `
<!DOCTYPE html>
<html>
  <head>
    <meta http-equiv="content-type" content="text/html; charset=utf-8">
    <title>Markdown Preview Tool</title>
  </head>
  <body>`

const footer = `
  </body>
</html>`

func main() {
	filename := flag.String("file", "", "Markdown file to preview")
	preview := flag.Bool("preview", false, "Auto preview the file")
	flag.Parse()

	if *filename == "" {
		flag.Usage()
		return
	}

	if err := run(*filename, os.Stdout, *preview); err != nil {
		log.Fatal(err)
	}
}

func run(filename string, out io.Writer, preview bool) error {
	input, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	htmlData := parseContent(input)

	temp, err := os.CreateTemp("", "mdp*.html")
	if err != nil {
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}

	outName := temp.Name()
	defer os.Remove(outName)

	fmt.Fprintln(out, outName)

	if err := saveHTML(outName, htmlData); err != nil {
		return err
	}

	if preview {
		previewFile(outName)
	}

	return nil
}

func parseContent(input []byte) []byte {
	output := blackfriday.Run(input)
	body := bluemonday.UGCPolicy().SanitizeBytes(output)

	var buffer bytes.Buffer

	buffer.WriteString(header)
	buffer.Write(body)
	buffer.WriteString(footer)

	return buffer.Bytes()
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
