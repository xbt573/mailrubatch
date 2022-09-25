package main

import "fmt"

type Progressbar struct {
	Total       int64
	Current     int64
	FileName    string
	Percent     int64
	CurrentFile int
	FilesCount  int
}

func NewProgressbar(fileName string, current, total int64, file, files int) *Progressbar {
	p := &Progressbar{
		Total:       total,
		Current:     current,
		FileName:    fileName,
		CurrentFile: file,
		FilesCount:  files,
	}
	p.Percent = p.GetPercent()

	return p
}

func (p *Progressbar) GetPercent() int64 {
	return int64((float32(p.Current) / float32(p.Total)) * 100)
}

func (p *Progressbar) Play(n int) {
	p.Current = int64(n)

	last := p.Percent
	p.Percent = p.GetPercent()

	if last == p.Percent {
		return
	}

	fmt.Printf("\r%3d%% %v %v/%v", p.Percent, p.FileName, p.CurrentFile, p.FilesCount)
}

func (p *Progressbar) Finish() {
	fmt.Println()
}

func (b *Progressbar) Write(p []byte) (n int, err error) {
	n = len(p)
	b.Play(int(b.Current) + n)
	return
}
