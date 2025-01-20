package main

import (
	"flag"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

type config struct {
	ext    string
	size   int64
	list   bool
	delete bool
}

func main() {
	c := config{}

	root := flag.String("root", ".", "Root directory to start")

	flag.BoolVar(&c.list, "list", false, "List files only")
	flag.BoolVar(&c.delete, "del", false, "Delete files")
	flag.StringVar(&c.ext, "ext", "", "File extension to filter out")
	flag.Int64Var(&c.size, "size", 0, "Minimum file size")

	flag.Parse()

	if err := run(*root, os.Stdout, c); err != nil {
		log.Fatal(err)
	}
}

func run(root string, out io.Writer, cfg config) error {
	return filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if filterOut(path, cfg.ext, cfg.size, info) {
			return nil
		}

		if cfg.list {
			return listFile(path, out)
		}

		if cfg.delete {
			return deleteFile(path)
		}

		return listFile(path, out)
	})
}
