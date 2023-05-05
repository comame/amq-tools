package main

type env struct {
	OutDir string `env:"OUTDIR"`
	UaMail string `env:"UA_MAIL"`
}

type anisongApiRequest struct {
	AnnId           uint `json:"annId"`
	IgnoreDuplicate bool `json:"ignore_duplicate"`
	OpeningFilter   bool `json:"opening_filter"`
	EndingFilter    bool `json:"ending_filter"`
	InsertFilter    bool `json:"insert_filter"`
}

type anisongApiResponse struct {
	AnnId          uint     `json:"annId"`
	CatboxAudioUrl string   `json:"audio"`
	AnimeNameEn    string   `json:"animeENName"`
	AnimeNameJp    string   `json:"animeJPName"`
	AnimeNameAlt   []string `json:"animeAltName"`
	AnimeVintage   string   `json:"animeVintage"`
	SongName       string   `json:"songName"`
}
