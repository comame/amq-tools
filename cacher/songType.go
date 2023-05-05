package main

import (
	"errors"
	"strconv"
	"strings"
)

type Song struct {
	AnimeName AnimeName    `json:"animeName"`
	Vintage   AnimeVintage `json:"vintage"`
	Name      string       `json:"name"`
	Filename  string       `json:"filename"`
}

type AnimeName struct {
	En  string   `json:"en"`
	Jp  string   `json:"jp"`
	Alt []string `json:"alt"`
}

type AnimeVintage struct {
	Year   uint   `json:"year"`
	Season Season `json:"season"`
}

type Season uint

const (
	Spring Season = iota
	Summer
	Fall
	Winter
)

func convertAnisongResponseToSong(res anisongApiResponse) (*Song, error) {
	parseVintage := func(str string) (*AnimeVintage, error) {
		result := new(AnimeVintage)

		spl := strings.Split(str, " ")
		if len(spl) != 2 {
			return nil, errors.New("InvalidVintageFormat")
		}

		y, err := strconv.ParseUint(spl[1], 10, 32)
		if err != nil {
			return nil, err
		}

		var s Season
		switch spl[0] {
		case "Spring":
			s = Spring
		case "Summer":
			s = Summer
		case "Fall":
			s = Fall
		case "Winter":
			s = Winter
		default:
			return nil, errors.New("InvalidVintageFormat")
		}

		result.Year = uint(y)
		result.Season = s

		return result, nil
	}

	s := new(Song)

	name := AnimeName{
		En:  res.AnimeNameEn,
		Jp:  res.AnimeNameJp,
		Alt: make([]string, 0),
	}
	if res.AnimeNameAlt != nil {
		name.Alt = res.AnimeNameAlt
	}
	s.AnimeName = name

	s.Name = res.SongName

	filename, err := extractFilenameFromUrl(res.CatboxAudioUrl)
	if err != nil {
		return nil, err
	}
	s.Filename = filename

	vint, err := parseVintage(res.AnimeVintage)
	if err != nil {
		s.Vintage = AnimeVintage{
			Year:   1970,
			Season: Spring,
		}
	} else {
		s.Vintage = *vint
	}

	return s, nil
}
