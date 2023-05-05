package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/comame/readenv-go"

	_ "embed"

	_ "github.com/mattn/go-sqlite3"
)

type env struct {
	WorkDir string `env:"WORKDIR"`
}

var WORKDIR = ""

//go:embed index.html
var INDEX_HTML string

func main() {
	var env env
	readenv.Read(&env)

	outdir, err := filepath.Abs(env.WorkDir)
	if err != nil {
		panic(err)
	}
	WORKDIR = outdir

	cmd, exists := getCommand()

	if !exists {
		startServer()
	}

	if cmd == "prepare" {
		prepareDb()
	}
}

type rawSong struct {
	AnimeId       uint
	AnimeName     string
	AnimeRawName  string
	Year          uint
	Season        Season
	SongName      string
	Filename      string
	SongType      string
	SongId        uint
	AnimeNameType string
}

func startServer() {
	dbPath := path.Join(WORKDIR, "anisong.db")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(INDEX_HTML))
	})
	http.HandleFunc("/api/search/byAnimeName", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var req animeNameSearchRequest
		err = json.Unmarshal(body, &req)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		res := searchByAnimeName(r.Context(), db, req, "anime_name.name LIKE ?", []any{"%" + req.AnimeName + "%"})

		if res == nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		resBytes, err := json.Marshal(res)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Write(resBytes)
	})

	log.Println("Start http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func prepareDb() {
	dbPath := path.Join(WORKDIR, "anisong.db")

	err := os.Remove(dbPath)
	if err != nil {
		panic(err)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE song (
		id auto increment integer primary key,
		name text not null,
		filename text not null,
		anime_id integer not null,
		type text not null
	)`)
	if err != nil {
		panic(err)
	}

	_, err = db.Exec(`CREATE TABLE anime (
		id integer primary key not null,
		year integer not null,
		season integer not null
	)`)
	if err != nil {
		panic(err)
	}

	_, err = db.Exec(`CREATE TABLE anime_name (
		anime_id integer not null,
		name text not null,
		raw_name text not null,
		type text not null
	)`)
	if err != nil {
		panic(err)
	}

	files, err := listJsonFilePaths()
	if err != nil {
		panic(err)
	}

	songs := make([]Song, 0)

	for _, filename := range files {
		f, err := os.Open(filename)
		if err != nil {
			panic(err)
		}
		defer f.Close()

		bytes, err := io.ReadAll(f)
		if err != nil {
			panic(err)
		}

		var s []Song
		json.Unmarshal(bytes, &s)

		songs = append(songs, s...)
	}

	commitOrRollback := func(tx *sql.Tx) {
		err = tx.Commit()
		if err != nil {
			log.Println(err)
			err = tx.Rollback()
			if err != nil {
				panic(err)
			}
		}
	}

	addSong := func(db *sql.DB, song Song, i uint) {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println(song)
				panic(err)
			}
		}()

		tx, err := db.Begin()
		if err != nil {
			panic(err)
		}
		defer commitOrRollback(tx)

		// song を INSERT
		stmt, err := tx.Prepare(`
			INSERT INTO song (id, name, filename, anime_id, type)
			VALUES (?, ?, ?, ?, ?)
		`)
		if err != nil {
			panic(err)
		}
		_, err = stmt.Exec(i, song.Name, song.Filename, song.AnimeId, song.Type)
		if err != nil {
			panic(err)
		}

		// anime が既に存在するかどうか
		stmt, err = tx.Prepare(`
			SELECT count(*) FROM anime WHERE id=?
		`)
		if err != nil {
			panic(err)
		}
		rows, err := stmt.Query(song.AnimeId)
		if err != nil {
			panic(err)
		}
		defer rows.Close()
		rows.Next()
		var animeCount uint
		err = rows.Scan(&animeCount)
		if err != nil {
			panic(err)
		}
		animeExists := animeCount != 0

		if !animeExists {
			// anime を追加
			stmt, err = tx.Prepare(`
				INSERT INTO anime (id, year, season) VALUES (?, ?, ?)
			`)
			if err != nil {
				panic(err)
			}
			_, err := stmt.Exec(song.AnimeId, song.Vintage.Year, song.Vintage.Season)
			if err != nil {
				panic(err)
			}

			// anime_name を追加
			addAnimeName := func(tx *sql.Tx, animeId uint, name string, nameType string) {
				stmt, err := tx.Prepare(`
					INSERT INTO anime_name (anime_id, name, raw_name, type) VALUES (?, ?, ?, ?)
				`)
				if err != nil {
					panic(err)
				}

				_, err = stmt.Exec(animeId, normalizeAnimeName(name), name, nameType)
				if err != nil {
					panic(err)
				}
			}
			addAnimeName(tx, song.AnimeId, song.AnimeName.En, "en")
			addAnimeName(tx, song.AnimeId, song.AnimeName.Jp, "jp")
			for _, v := range song.AnimeName.Alt {
				addAnimeName(tx, song.AnimeId, v, "alt")
			}
		}
	}

	for i, song := range songs {
		addSong(db, song, uint(i))

		fmt.Printf("Done: %d / %d\n", i, len(songs))
	}
}

func getCommand() (string, bool) {
	args := os.Args

	if len(args) >= 2 {
		return args[1], true
	}
	return "", false
}
