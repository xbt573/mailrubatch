package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func GetResponse(folder string) (Response, error) {
	req, err := http.Get(fmt.Sprintf("https://cloud.mail.ru/api/v2/folder?weblink=%v", folder))
	if err != nil {
		return Response{}, err
	}
	defer req.Body.Close()

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return Response{}, err
	}

	var res Response
	err = json.Unmarshal(body, &res)
	if err != nil {
		return Response{}, err
	}

	return res, nil
}

func GetWeblink() (string, error) {
	req, err := http.Get("https://cloud.mail.ru/api/v2/dispatcher")
	if err != nil {
		return "", err
	}
	defer req.Body.Close()

	data, err := io.ReadAll(req.Body)
	if err != nil {
		return "", err
	}

	var res WeblinkResponse
	err = json.Unmarshal(data, &res)
	if err != nil {
		return "", err
	}

	return res.Body.WeblinkGet[0].Url, nil
}
