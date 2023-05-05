package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
)

var ANISONGDB_HOST = "https://anisongdb.com"
var CATBOX_HOST = "https://files.catbox.moe"
var UA_MAIL string

// 第 2 引数は annId が存在するかどうかを返す
func fetchAnisongdb(annId uint) ([]anisongApiResponse, bool, error) {
	body := anisongApiRequest{
		AnnId:           annId,
		IgnoreDuplicate: false,
		OpeningFilter:   true,
		EndingFilter:    true,
		InsertFilter:    true,
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, false, err
	}

	req, err := http.NewRequest("POST", ANISONGDB_HOST+"/api/annId_request", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, false, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "go-cacher "+UA_MAIL)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, false, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("ansiongdb returns non-200 status code")
	}

	resBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, false, err
	}

	var resObj []anisongApiResponse
	err = json.Unmarshal(resBytes, &resObj)
	if err != nil {
		return nil, false, err
	}

	if len(resObj) == 0 {
		return nil, false, nil
	}

	// mp3 がないときはスキップする
	result := make([]anisongApiResponse, 0)
	for _, v := range resObj {
		if v.CatboxAudioUrl == "" {
			continue
		}
		result = append(result, v)
	}

	return result, true, nil
}

func extractFilenameFromUrl(urlStr string) (string, error) {
	url, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}

	_, filename := path.Split(url.Path)

	return filename, nil
}

func downloadAndSaveCatboxFile(filename string) error {
	req, err := http.NewRequest("GET", CATBOX_HOST+"/"+filename, nil)
	if err != nil {
		return err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	p := calculateMusicFilePath(filename)
	file, err := os.Create(p)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, res.Body)
	if err != nil {
		return err
	}

	return nil
}
