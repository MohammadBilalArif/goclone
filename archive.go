package main

import (
	"archive/tar"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
)

func archiveFiles(root string, c chan fileHeader, w io.ReadWriter, output chan io.Reader) {
	tw := tar.NewWriter(w)

	var fileCount int

	root = path.Clean(root)

	for info := range c {
		fileCount++

		fp, err := os.Open(path.Join(root, info.Path, info.Name))
		if err != nil {
			log.Println("[!] attempted to archive invalid file", path.Join(info.Path, info.Name), err)
			continue
		}

		data, err := ioutil.ReadAll(fp)
		fp.Close()
		if err != nil {
			log.Println("[!] error reading entire file", path.Join(info.Path, info.Name), err)
			continue
		}

		hdr := &tar.Header{
			Name: path.Join(info.Path, info.Name),
			Mode: info.Mode,
			Size: int64(len(data)),
		}

		if err := tw.WriteHeader(hdr); err != nil {
			log.Println("[!] unable to write tar file header", err)
			return
		}

		if _, err := tw.Write([]byte(data)); err != nil {
			log.Println("[!] unable to write tar file body", err)
			return
		}
	}

	if fileCount == 0 {
		return
	}

	if err := tw.Close(); err != nil {
		log.Println("[!] unable to close tar archive", err)
		return
	}

	output <- w
}

func expandFiles(root string, tarFile io.Reader) {
	tr := tar.NewReader(tarFile)

	root = path.Clean(root)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Println("[!] error expanding tar file", err)
			break
		}
		log.Println("Making Directories:", path.Dir(path.Join(root, hdr.Name)))
		log.Println("Creating File:", path.Join(root, hdr.Name))

		// ensure all paths exist
		if err := os.MkdirAll(path.Dir(path.Join(root, hdr.Name)), 0755); err != nil {
			log.Println("[!] error creating directories for tar expansion", err)
			continue
		}
		fp, err := os.OpenFile(path.Join(root, hdr.Name), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.FileMode(hdr.Mode))
		if err != nil {
			log.Println("[!] error opening expanded file for writing", err)
			continue
		}

		if _, err := io.Copy(fp, tr); err != nil {
			log.Println("[!] unable to write expanded file", err)
			continue
		}

		log.Println("[*] found expanded file", hdr.Name)
	}
}
