package main

type WeblinkResponse struct {
	Body WeblinkBody `json:"body"`
}

type WeblinkBody struct {
	WeblinkGet []WeblinkGet `json:"weblink_get"`
}

type WeblinkGet struct {
	Url string `json:"url"`
}

type Response struct {
	Body Body `json:"body"`
}

type Body struct {
	Size int64       `json:"size"`
	List []CloudFile `json:"list"`
}

type CloudFile struct {
	Size    int64  `json:"size"`
	Name    string `json:"name"`
	Weblink string `json:"weblink"`
	Type    string `json:"type"`
}

type FileNode interface {
	IsTree() bool
}

type File struct {
	Size    int64
	Name    string
	Weblink string
}

func (File) IsTree() bool {
	return false
}

type FileTree struct {
	Folder string
	Files  []FileNode
}

func (FileTree) IsTree() bool {
	return true
}
