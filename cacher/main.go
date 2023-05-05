package main

import (
	"fmt"
	"path/filepath"
	"sync"

	"github.com/comame/readenv-go"
)

var START_ANN_ID uint = 1
var MAX_ANN_ID uint = 30000
var THREADS_LIMIT = 10
var WORKDIR = "/tmp/dummy"

func main() {
	var env env
	readenv.Read(&env)

	outdir, err := filepath.Abs(env.OutDir)
	if err != nil {
		panic(err)
	}
	WORKDIR = outdir

	UA_MAIL = env.UaMail

	var wg sync.WaitGroup
	limit := make(chan struct{}, THREADS_LIMIT)

	for id := START_ANN_ID; id <= MAX_ANN_ID; id += 1 {
		limit <- struct{}{}
		wg.Add(1)
		go downloadAsync(id, limit, &wg)
	}

	wg.Wait()
}

func downloadAsync(annId uint, limit chan struct{}, wg *sync.WaitGroup) {
	for {
		err := download(annId)
		if err == nil {
			break
		}

		fmt.Println(err)
		fmt.Printf("Retrying: %d\n", annId)
	}

	wg.Done()
	<-limit
}

func download(annId uint) error {
	fmt.Printf("Start download: %d\n", annId)

	exists := getAnnIdIsCached(annId)
	if exists {
		fmt.Printf("  Skipped: %d\n", annId)
		return nil
	}

	res, exists, err := fetchAnisongdb(annId)
	if err != nil {
		return err
	}

	if !exists {
		fmt.Printf("  AnnId not exists: %d\n", annId)
		return nil
	}

	// 同一アニメ内の複数楽曲も並列ダウンロードする
	// 多少数が多くても気にせずに全件並列ダウンロードする
	// TODO: 後でコメントアウトを外す。今は anisongdb のレスポンスを保存したいだけ

	// var wg sync.WaitGroup
	// for _, obj := range res {
	// 	wg.Add(1)

	// 	filename, err := extractFilenameFromUrl(obj.CatboxAudioUrl)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	fmt.Println("  Downloading: " + filename)

	// 	prepareMusicFileDir(filename)

	// 	go downloadFile(filename, &wg)
	// }
	// wg.Wait()

	songs := make([]Song, 0)
	for _, v := range res {
		s, err := convertAnisongResponseToSong(v)
		if err != nil {
			return err
		}
		songs = append(songs, *s)
	}
	err = writeAnisongDbJson(annId, songs)
	if err != nil {
		return err
	}

	fmt.Printf("  Done: %d\n", annId)
	return nil
}

func downloadFile(filename string, wg *sync.WaitGroup) {
	for {
		err := downloadAndSaveCatboxFile(filename)
		if err == nil {
			break
		}
	}
	wg.Done()
}
