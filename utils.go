package main

import (
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
	for _, file := range tree.Files {
		if file.IsTree() {
			folder := file.(*FileTree).Folder

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

	progress := NewProgressbar(
		fmt.Sprintf("%v/%v", folder, file.Name),
		0,
		file.Size,
		currentFile,
		allFiles,
	)

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
