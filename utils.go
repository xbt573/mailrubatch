package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
)

var allSize int64
var downloadedSize int64

func GetPercent() int {
	if allSize == 0 {
		return 100
	}

	if downloadedSize == 0 {
		return 0
	}

	return int(downloadedSize / (allSize / 100))
}

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
	}

	allSize = res.Body.Size
	return tree, nil
}

func DownloadFiles(downloadFolder string, tree *FileTree) {
	for _, file := range tree.Files {
		if file.IsTree() {
			folder := file.(*FileTree).Folder
			log.Printf("found folder %v, recursively downloading it...", folder)

			err := os.Mkdir(folder, os.ModePerm)
			if err != nil {
				log.Println(err)
			}

			DownloadFiles(folder, file.(*FileTree))
			continue
		}

		DownloadFile(downloadFolder, file.(File))
	}
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

	_, err = io.Copy(f, res.Body)
	if err != nil {
		log.Printf("download %v: %v", fmt.Sprintf("%v/%v", folder, file.Name), err)
		os.Remove(fmt.Sprintf("%v/%v", folder, file.Name))
		return
	}

	downloadedSize += file.Size
	log.Printf("download %v: finished (%v%%)", fmt.Sprintf("%v/%v", folder, file.Name), GetPercent())
}
