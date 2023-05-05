package main

type Song struct {
	AnimeId   uint         `json:"animeId"`
	AnimeName AnimeName    `json:"animeName"`
	Vintage   AnimeVintage `json:"vintage"`
	Name      string       `json:"name"`
	Filename  string       `json:"filename"`
	Type      string       `json:"type"`
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
