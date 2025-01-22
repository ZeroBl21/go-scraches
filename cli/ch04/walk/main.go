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
	name   string
	ext    string
	size   int64
	list   bool
	delete bool

	wLog    io.Writer
	archive string
}

func main() {
	c := config{}

	root := flag.String("root", ".", "Root directory to start")
	logFile := flag.String("log", "", "Log deletes to this file")

	flag.StringVar(&c.archive, "archive", "", "Archive directory")
	flag.StringVar(&c.name, "name", "", "File name to filter out")
	flag.StringVar(&c.ext, "ext", "", "File extension to filter out")
	flag.Int64Var(&c.size, "size", 0, "Minimum file size")

	flag.BoolVar(&c.list, "list", false, "List files only")
	flag.BoolVar(&c.delete, "del", false, "Delete files")

	flag.Parse()

	var (
		f   = os.Stdout
		err error
	)
	if *logFile != "" {
		f, err = os.OpenFile(*logFile, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
	}
	c.wLog = f

	if err := run(*root, os.Stdout, c); err != nil {
		log.Fatal(err)
	}
}

func run(root string, out io.Writer, cfg config) error {
	deleteLogger := log.New(cfg.wLog, "DELETED FILE: ", log.LstdFlags)

	return filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if filterOut(path, cfg.ext, cfg.name, cfg.size, info) {
			return nil
		}

		if cfg.list {
			return listFile(path, out)
		}

		if cfg.archive != "" {
			if err := archiveFile(cfg.archive, root, path); err != nil {
				return err
			}
		}

		if cfg.delete {
			return deleteFile(path, deleteLogger)
		}

		return listFile(path, out)
	})
}
