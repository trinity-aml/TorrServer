package poster_tmdb

import (
	"fmt"
	tmdb "github.com/cyruzin/golang-tmdb"
	"github.com/essentialkaos/translit"
	"regexp"
	"server/config"
	"server/log"
	"strconv"
	"strings"
)

var tor_q = []string{
	"Blu-ray",
	"BDRemux",
	"BDRip",
	"HDRip",
	"WEB-DL",
	"WEBDL",
	"XviD",
	"WEB-DLRip",
	"WEBRip",
	"HDTV",
	"HDTVRip",
	"DVD",
	"DVD9",
	"DVD5",
	"DVDRip",
	"DVDScr",
	"DVDRemux",
	"DVB",
	"SATRip",
	"IPTVRip",
	"TVRip",
	"VHSRip",
	"TS",
	"CAMRip",
	"2160p",
	"1080p",
	"1080i",
	"720p",
	"576p",
	"576i",
	"480p",
	"400p",
}

var (
	time, _  = regexp.Compile("[0-9][0-9][0-9][0-9]")
	sq, _    = regexp.Compile("\\[[a-zA-Zа-яА-Я0-9-[:space:]+.,].+\\]")
	ss, _    = regexp.Compile("[Ss][0-9][0-9]")
	ee, _    = regexp.Compile("[Ee][0-9][0-9]")
	ss2, _   = regexp.Compile("[0-9][0-9][a-zA-Zа-яА-Я][0-9][0-9][-][0-9][0-9]:space:[a-zA-Zа-яА-Я][a-zA-Zа-яА-Я]:space:[0-9][0-9]")
	ss3, _   = regexp.Compile("[Ss][Ee][Zz][Oo][Nn]:space:[0-9]")
	ss4, _   = regexp.Compile("[Ss][Ee][Rr][Ii][Ii]:space:[0-9]")
	ss5, _   = regexp.Compile("[Ss][0-9][0-9][-][0-9][0-9]")
	latin, _ = regexp.Compile("[A-Za-z]+")
	cyr, _   = regexp.Compile("[А-Яа-я]+")
)

func getUtil(str string, tv bool, movie bool) string {

	var media_type string
	var poster = "https://image.tmdb.org/t/p/original"

	tmdbClient, err := tmdb.Init(config.ReadConfigParser2("Api_key"))

	if err != nil {
		fmt.Println(err)
	}

	options := make(map[string]string)
	options["language"] = "ru-RU"

	search, err2 := tmdbClient.GetSearchMulti(str, options)

	if err2 == nil {
		if tv == true {
			media_type = "tv"
		} else if movie == true {
			media_type = "movie"
		}
		for o := 0; o < int(search.TotalResults); o++ {
			if search.Results[o].MediaType == media_type {
				log.TLogln("Poster:", poster+search.Results[o].PosterPath)
				return poster + search.Results[o].PosterPath
			}
		}
	} else {
		fmt.Println(err2)
		return ""
	}

	return ""
}

func GetPoster(name string) string {
	var year = 0
	var movie = true
	var tv = false

	name = strings.ReplaceAll(name, ".", " ")
	name = strings.ReplaceAll(name, "_", " ")
	name = strings.ReplaceAll(name, "|", "")
	name = strings.ReplaceAll(name, "(", "")
	name = strings.ReplaceAll(name, ")", "")
	name = strings.ReplaceAll(name, " RUS ", " ")
	name = strings.ReplaceAll(name, " ENG ", " ")

	nameMass := strings.Split(name, " ")

	for _, word := range nameMass {
		if len(time.FindString(word)) > 0 {
			if strings.Contains(word, "-") {
				for _, a := range word {
					year, _ = strconv.Atoi(time.FindString(strconv.Itoa(int(a))))
					break
				}
			} else {
				year, _ = strconv.Atoi(time.FindString(word))
			}
		}
		if len(ss.FindString(word)) > 0 || len(ee.FindString(word)) > 0 || len(ss2.FindString(word)) > 0 || len(ss3.FindString(word)) > 0 || len(ss4.FindString(word)) > 0 || len(ss5.FindString(word)) > 0 {
			tv = true
			movie = false
		}
	}

	for i, word := range nameMass {
		for _, word2 := range tor_q {
			if strings.EqualFold(word, word2) {
				for l := i; l < len(nameMass); l++ {
					nameMass[l] = ""
				}
				nameMass = nameMass[:i]
			}
		}
		if len(time.FindString(word)) > 0 {
			for m := i; m < len(nameMass); m++ {
				nameMass[m] = ""
			}
			nameMass = nameMass[:i]
		}
		if len(ss.FindString(word)) > 0 || len(ee.FindString(word)) > 0 {
			for k := i; k < len(nameMass); k++ {
				nameMass[k] = ""
			}
			nameMass = nameMass[:i]
		}
	}

	nameMassNew := strings.Join(nameMass, " ")

	if len(nameMassNew) > 0 {
		nameMassNew = strings.Trim(nameMassNew, " ")
	}

	nameMassNew = strings.ReplaceAll(nameMassNew, sq.FindString(nameMassNew), "")

	if strings.Contains(nameMassNew, "/") {
		nameMass = strings.Split(nameMassNew, "/")
		for _, word := range nameMass {
			out := latin.FindString(word)
			if len(out) > 2 {
				nameMassNew = word
				break
			}
		}
		if len(nameMassNew) > 0 {
			nameMassNew = strings.Trim(nameMassNew, " ")
		} else {
			for _, word2 := range nameMass {
				nameMassNew = word2
				break
			}
			if len(nameMassNew) > 0 {
				nameMassNew = strings.Trim(nameMassNew, " ")
			}
		}
	}

	log.TLogln(nameMassNew)

	var nameMassNew2 string

	if len(cyr.FindString(nameMassNew)) > 0 {
		nameMassNew2 = translit.EncodeToISO9B(nameMassNew)
		log.TLogln(nameMassNew2)
	}

	if len(getUtil(nameMassNew+strconv.Itoa(year), tv, movie)) > 0 {
		return getUtil(nameMassNew+strconv.Itoa(year), tv, movie)
	} else if len(getUtil(nameMassNew, tv, movie)) > 0 {
		return getUtil(nameMassNew, tv, movie)
	} else if len(getUtil(nameMassNew2, tv, movie)) > 0 {
		return getUtil(nameMassNew2, tv, movie)
	} else {
		return ""
	}
}
