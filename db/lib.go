package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path"
	"regexp"
	"strings"
	"unicode"
)

type animeNameSearchRequest struct {
	AnimeName string `json:"animeName"`
}

type vintageSearchRequest struct {
	Season *Season `json:"season"`
	Year   uint    `json:"year"`
}

type searchResultResponse struct {
	Values []Song `json:"values"`
}

func listJsonFilePaths() ([]string, error) {
	ent, err := os.ReadDir(path.Join(WORKDIR, "annId"))
	if err != nil {
		return nil, err
	}

	dirs := make([]string, 0)
	for _, v := range ent {
		if v.IsDir() {
			dirs = append(dirs, v.Name())
		}
	}

	files := make([]string, 0)
	for _, d := range dirs {
		p := path.Join(WORKDIR, "annId", d)
		ents, err := os.ReadDir(p)
		if err != nil {
			return nil, err
		}

		for _, ent := range ents {
			files = append(files, path.Join(p, ent.Name()))
		}
	}

	return files, nil
}

func normalizeAnimeName(str string) string {
	isAcceptable := func(c rune) bool {
		return c <= unicode.MaxASCII
	}

	result := make([]string, 0)

	str = strings.ToLower(str)

	for _, c := range str {
		if isAcceptable(c) {
			result = append(result, string(c))
		} else {
			result = append(result, " ")
		}
	}

	str = strings.Join(result, "")

	reg := regexp.MustCompile(`\s+`)
	str = reg.ReplaceAllString(str, " ")

	return str
}

func searchByAnimeName(ctx context.Context, db *sql.DB, req animeNameSearchRequest, sqlWhereCond string, queryArgs []any) *searchResultResponse {
	stmt, err := db.Prepare(fmt.Sprintf(`
				SELECT anime.id AS anime_id, anime_name.name AS anime_name, anime_name.raw_name AS anime_raw_name, anime.year, anime.season, song.name AS song_name, song.filename, song.type AS song_type, song.id AS song_id, anime_name.type AS anime_name_type
				FROM song
				LEFT OUTER JOIN anime ON song.anime_id = anime.id
				LEFT OUTER JOIN anime_name ON anime.id = anime_name.anime_id
				WHERE %s
			`, sqlWhereCond))
	if err != nil {
		panic(err)
	}

	rows, err := stmt.QueryContext(ctx, queryArgs...)
	if err != nil {
		log.Println(err)
		return nil
	}

	rawSongs := make([]rawSong, 0)
	for rows.Next() {
		var animeId uint
		var animeName string
		var animeRawName string
		var year uint
		var season Season
		var songName string
		var filename string
		var songType string
		var songId uint
		var animeNameType string

		err = rows.Scan(&animeId, &animeName, &animeRawName, &year, &season, &songName, &filename, &songType, &songId, &animeNameType)
		if err != nil {
			log.Println(err)
			return nil
		}

		rawSongs = append(rawSongs, rawSong{
			AnimeId:       animeId,
			AnimeName:     animeName,
			AnimeRawName:  animeRawName,
			Year:          year,
			Season:        season,
			SongName:      songName,
			Filename:      filename,
			SongType:      songType,
			SongId:        songId,
			AnimeNameType: animeNameType,
		})
	}

	songsMap := make(map[uint]Song)
	for _, rawSong := range rawSongs {
		curr, ok := songsMap[rawSong.SongId]

		if ok {
			switch rawSong.AnimeNameType {
			case "en":
				curr.AnimeName.En = rawSong.AnimeName
			case "jp":
				curr.AnimeName.Jp = rawSong.AnimeName
			case "alt":
				curr.AnimeName.Alt = append(curr.AnimeName.Alt, rawSong.AnimeName)
			default:
				panic("unreachable")
			}
		} else {
			song := new(Song)
			song.AnimeId = rawSong.AnimeId
			song.Vintage.Season = rawSong.Season
			song.Vintage.Year = rawSong.Year
			song.Name = rawSong.SongName
			song.Filename = rawSong.Filename
			song.Type = rawSong.SongType

			song.AnimeName.Alt = make([]string, 0)

			switch rawSong.AnimeNameType {
			case "en":
				song.AnimeName.En = rawSong.AnimeName
			case "jp":
				song.AnimeName.Jp = rawSong.AnimeName
			case "alt":
				song.AnimeName.Alt = append(song.AnimeName.Alt, rawSong.AnimeName)
			default:
				panic("unreachable")
			}

			songsMap[rawSong.SongId] = *song
		}
	}

	for songId := range songsMap {
		animeId := songsMap[songId].AnimeId

		stmt, err = db.Prepare(`
				SELECT name, raw_name, type
				FROM anime_name
				WHERE anime_id=?
			`)
		if err != nil {
			log.Println(err)
			return nil
		}

		rows, err = stmt.QueryContext(ctx, animeId)
		if err != nil {
			log.Println(err)
			return nil
		}

		for rows.Next() {
			var name string
			var rawName string
			var nameType string
			err = rows.Scan(&name, &rawName, &nameType)
			if err != nil {
				log.Println(err)
				return nil
			}

			song := songsMap[songId]
			switch nameType {
			case "en":
				song.AnimeName.En = name
			case "jp":
				song.AnimeName.Jp = name
			case "alt":
				song.AnimeName.Alt = append(song.AnimeName.Alt, name)
			default:
				panic("unreachable")
			}

			songsMap[songId] = song
		}
	}

	songs := make([]Song, 0)
	for k := range songsMap {
		songs = append(songs, songsMap[k])
	}

	return &searchResultResponse{
		Values: songs,
	}
}
