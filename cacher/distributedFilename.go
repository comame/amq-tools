package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
)

// AnnID の末尾 2 桁ごとにフォルダを分ける
func writeAnisongDbJson(annId uint, value []Song) error {
	annIdStr := fmt.Sprint(annId)

	dir := ("00" + annIdStr)[len(annIdStr):]
	dirPath := path.Join(WORKDIR, "annId", dir)

	err := os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		return err
	}

	fp := path.Join(dirPath, annIdStr+".json")
	f, err := os.Create(fp)
	if err != nil {
		return err
	}
	defer f.Close()

	valueBytes, err := json.Marshal(value)
	if err != nil {
		return err
	}

	f.Write(valueBytes)

	return nil
}

// 対象の AnnID の mp3 をダウンロード済みかどうかを返す
func getAnnIdIsCached(annId uint) bool {
	annIdStr := fmt.Sprint(annId)

	dir := ("00" + annIdStr)[len(annIdStr):]
	filepath := path.Join(WORKDIR, "annId", dir, annIdStr+".json")

	_, err := os.Stat(filepath)
	return err == nil
}

// ダウンロード先のフォルダがなければ作成する
func prepareMusicFileDir(filename string) error {
	p := calculateMusicFilePath(filename)
	d, _ := path.Split(p)

	err := os.MkdirAll(d, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

// ダウンロード先のパスを求める
func calculateMusicFilePath(filename string) string {
	dirname := filename[:2]
	return path.Join(WORKDIR, "file", dirname, filename)
}
