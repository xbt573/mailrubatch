package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
)

var currentFile = 1
var allFiles int

func GetFiles(folder string) (*FileTree, error) {
	tree := new(FileTree)

	res, err := GetResponse(url.QueryEscape(folder))
	if err != nil {
		return nil, err
	}

	for _, file := range res.Body.List {
		if file.Type == "folder" {
			files, err := GetFiles(file.Weblink)
			if err != nil {
				return nil, err
			}

			files.Folder = file.Name

			tree.Files = append(tree.Files, files)
			continue
		}

		tree.Files = append(tree.Files, File{
			Name:    file.Name,
			Weblink: file.Weblink,
			Size:    file.Size,
		})
		allFiles++
	}

	return tree, nil
}

func DownloadFiles(downloadFolder string, tree *FileTree) {
	if downloadFolder != "." {
		os.Mkdir(downloadFolder, os.ModePerm)
	}

	for _, file := range tree.Files {
		if file.IsTree() {
			folder := file.(*FileTree).Folder

			DownloadFiles(fmt.Sprintf("%v/%v", downloadFolder, folder), file.(*FileTree))
			continue
		}

		DownloadFile(downloadFolder, file.(File))
	}

	currentFile = 1
}

func DownloadFile(folder string, file File) {
	weblink, err := GetWeblink()
	if err != nil {
		log.Printf("download %v: %v", fmt.Sprintf("%v/%v", folder, file.Name), err)
		return
	}

	res, err := http.Get(fmt.Sprintf("%v/%v", weblink, file.Weblink))
	if err != nil {
		log.Printf("download %v: %v", fmt.Sprintf("%v/%v", folder, file.Name), err)
		return
	}
	defer res.Body.Close()

	f, err := os.Create(fmt.Sprintf("%v/%v", folder, file.Name))
	if err != nil {
		log.Printf("download %v: %v", fmt.Sprintf("%v/%v", folder, file.Name), err)
		return
	}
	defer f.Close()

	progress := NewProgressbar(
		fmt.Sprintf("%v/%v", folder, file.Name),
		0,
		file.Size,
		currentFile,
		allFiles,
	)
	progress.Play(0)

	_, err = io.Copy(io.MultiWriter(f, progress), res.Body)
	if err != nil {
		progress.Finish()
		log.Printf("download %v: %v", fmt.Sprintf("%v/%v", folder, file.Name), err)
		os.Remove(fmt.Sprintf("%v/%v", folder, file.Name))
		return
	}

	currentFile++
	progress.Finish()
}

func RemoveFiles(basePath string, tree *FileTree) {
	for _, node := range tree.Files {
		if node.IsTree() {
			subTree := node.(*FileTree)

			RemoveFiles(
				fmt.Sprintf("%v/%v", basePath, subTree.Folder),
				subTree,
			)
			continue
		}

		file := node.(File)
		os.Remove(
			fmt.Sprintf("%v/%v", basePath, file.Name),
		)
	}

	if basePath != "." {
		os.Remove(basePath)
	}
}

func FlatTree(tree *FileTree) []string {
	flatFiles := make([]string, 0)

	for _, node := range tree.Files {
		if node.IsTree() {
			subTree := node.(*FileTree)
			files := FlatTree(subTree)

			for _, subFile := range files {
				flatFiles = append(
					flatFiles,
					fmt.Sprintf("%v/%v", tree.Folder, subFile),
				)
			}
			continue
		}

		file := node.(File)
		flatFiles = append(
			flatFiles,
			fmt.Sprintf("%v/%v", tree.Folder, file.Name),
		)
	}

	return flatFiles
}

func ArchiveFiles(basePath string, tree *FileTree) {
	flatFiles := FlatTree(tree)

	var path = basePath
	if path == "." {
		path = "archive"
	}

	archive, err := os.Create(
		fmt.Sprintf("%v.tar.gz", path),
	)
	if err != nil {
		log.Fatalln(err)
	}
	defer archive.Close()

	gzipBuffer := gzip.NewWriter(archive)
	defer gzipBuffer.Close()

	tarBuffer := tar.NewWriter(gzipBuffer)
	defer tarBuffer.Close()

	for _, path := range flatFiles {
		file, err := os.Open(path)
		if err != nil {
			log.Fatalln(err)
		}
		defer file.Close()

		stat, err := file.Stat()
		if err != nil {
			log.Fatalln(err)
		}

		header := &tar.Header{
			Name:    path,
			Mode:    0666,
			Size:    stat.Size(),
			ModTime: stat.ModTime(),
			Uid:     os.Getuid(),
			Gid:     os.Getgid(),
		}

		if err := tarBuffer.WriteHeader(header); err != nil {
			log.Fatalln(err)
		}

		progress := NewProgressbar(
			path,
			0,
			stat.Size(),
			currentFile,
			allFiles,
		)

		if _, err := io.Copy(io.MultiWriter(progress, tarBuffer), file); err != nil {
			progress.Finish()
			log.Fatalln(err)
		}

		currentFile++
		progress.Finish()
	}
}
