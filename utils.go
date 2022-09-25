package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
)

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
		})
	}

	return tree, nil
}

func DownloadFiles(downloadFolder string, tree *FileTree) {
	for _, file := range tree.Files {
		if file.IsTree() {
			folder := file.(*FileTree).Folder
			log.Printf("found folder %v, recursively downloading it...", folder)

			err := os.Mkdir(folder, os.ModePerm)
			if err != nil {
				panic(err)
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

	data, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("download %v: %v", fmt.Sprintf("%v/%v", folder, file.Name), err)
		return
	}

	err = os.WriteFile(fmt.Sprintf("%v/%v", folder, file.Name), data, os.ModePerm)
	if err != nil {
		log.Printf("download %v: %v", fmt.Sprintf("%v/%v", folder, file.Name), err)
		return
	}

	log.Printf("download %v: finished", fmt.Sprintf("%v/%v", folder, file.Name))
}
