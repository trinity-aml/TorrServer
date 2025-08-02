package poster_tmdb

import (
	"fmt"
	tmdb "github.com/cyruzin/golang-tmdb"
	"github.com/essentialkaos/translit"
	"math"
	"regexp"
	"server/config"
	"server/log"
	"strconv"
	"strings"
	"time"
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
	time2, _ = regexp.Compile("[0-9][0-9][0-9][0-9]")
	sq, _    = regexp.Compile("\\[[a-zA-Zа-яА-Я0-9-[:space:]+.,].+\\]")
	ss, _    = regexp.Compile("[Ss][0-9][0-9]")
	ee, _    = regexp.Compile("[Ee][0-9][0-9]")
	ss2, _   = regexp.Compile("[0-9][0-9][a-zA-Zа-яА-Я][0-9][0-9][-][0-9][0-9]")
	ss3, _   = regexp.Compile("[Ss][Ee][Zz][Oo][Nn][[:space:]][0-9]")
	ss4, _   = regexp.Compile("[Ss][Ee][Rr][Ii][Ii][[:space:]][0-9]")
	ss5, _   = regexp.Compile("[Ss][0-9][0-9][-][0-9][0-9]")
	ss6, _   = regexp.Compile("[0-9][0-9][-][0-9][0-9]")
	latin, _ = regexp.Compile("[A-Za-z]+")
	cyr, _   = regexp.Compile("[А-Яа-я]+")
)

func getYear(str string) int {
	layout := "2006-01-02"
	date, err := time.Parse(layout, str)
	if err != nil {
		return 0
	}
	year, _, _ := date.Date()
	return year
}

func compareStr(name string, str string) int {
	i := 0
	k := 0
	reg, err := regexp.Compile("[^a-zA-Zа-яА-Я0-9]+")
	if err == nil {
		name = reg.ReplaceAllString(name, "")
		str = reg.ReplaceAllString(str, "")
	} else {
		fmt.Println(err)
	}
	name = strings.ToLower(name)
	str = strings.ToLower(str)
	if int(math.Abs(float64(len(name)-len(str)))) > 3 {
		return 0
	}
	for _, a := range name {
		for _, b := range str {
			if a == b {
				i = 1
				k = 1
				break
			} else {
				k = 0
			}
		}
		if k == 0 {
			i = 0
			break
		}
	}
	return i
}

func getUtil(str string, tv bool, movie bool, y int) string {

	var media_type string
	var poster = "https://image.tmdb.org/t/p/original"
	var year = 0
	var year_1 = 0
	var year_2 = 0

	api_key := config.ReadConfigParser2("Api_key")
	if api_key == "" {
		api_key = config.RandApiKey()
	}

	tmdbClient, err := tmdb.Init(api_key)

	if err != nil {
		fmt.Println(err)
	}

	options := make(map[string]string)
	options["language"] = "ru-RU"

	search, err2 := tmdbClient.GetSearchMulti(str, nil)
	release_d := ""
	fist_air := ""
	comp := 0
	exit := ""
	bypass := ""
	bypassA := ""

	if err2 == nil {
		if tv == true {
			media_type = "tv"
		} else if movie == true {
			media_type = "movie"
		}
		for o, v := range search.Results {
			release_d = v.ReleaseDate
			fist_air = v.FirstAirDate
			if release_d != "" {
				year_1 = getYear(release_d)
			}
			if fist_air != "" {
				year_2 = getYear(fist_air)
			}
			if year_1 != 0 {
				year = year_1
			} else if year_2 != 0 {
				year = year_2
			} else {
				year = 0
			}
			if y == 0 || (int(math.Abs(float64(y-year))) == 1) {
				y = year
			}
			if v.Name != "" {
				comp = compareStr(str, v.Name)
			} else if v.OriginalName != "" {
				comp = compareStr(str, v.OriginalName)
			} else if search.Results[o].Title != "" {
				comp = compareStr(str, v.Title)
			}
			if v.MediaType == media_type && y == year && comp == 1 && v.PosterPath != "" {
				log.TLogln("Poster:", poster+v.PosterPath)
				exit = poster + v.PosterPath
				break
			} else if v.MediaType != media_type && y == year && comp == 1 && v.PosterPath != "" {
				if media_type == "tv" {
					if v.MediaType == "movie" {
						log.TLogln("Poster:", poster+v.PosterPath)
						bypassA = poster + v.PosterPath
					}
				} else if media_type == "movie" {
					if v.MediaType == "tv" {
						log.TLogln("Poster:", poster+v.PosterPath)
						bypassA = poster + v.PosterPath
					}
				}
			} else if v.MediaType == media_type && y == year && comp == 0 && v.PosterPath != "" {
				log.TLogln("Poster:", poster+v.PosterPath)
				bypass = poster + v.PosterPath
			}
		}
		if exit != "" {
			return exit
		} else if bypassA != "" {
			return bypassA
		} else if bypass != "" {
			return bypass
		} else {
			return ""
		}
	} else {
		fmt.Println(err2)
		return ""
	}
}

func GetPoster(name string) string {
	var year int
	var movie = true
	var tv = false
	var en_rus int

	name = strings.ReplaceAll(name, ".", " ")
	name = strings.ReplaceAll(name, "_", " ")
	name = strings.ReplaceAll(name, "|", "")
	name = strings.ReplaceAll(name, "(", "")
	name = strings.ReplaceAll(name, ")", "")
	name = strings.ReplaceAll(name, " RUS ", " ")
	name = strings.ReplaceAll(name, " ENG ", " ")

	nameMass := strings.Split(name, " ")

	if len(ss.FindString(name)) > 0 || len(ee.FindString(name)) > 0 || len(ss2.FindString(name)) > 0 || len(ss3.FindString(name)) > 0 || len(ss4.FindString(name)) > 0 || len(ss5.FindString(name)) > 0 || len(ss6.FindString(name)) > 0 {
		tv = true
		movie = false
		log.TLogln("Serial: ", tv)
	} else {
		log.TLogln("Movie: ", movie)
	}

	for _, word := range nameMass {
		if len(time2.FindString(word)) > 0 && time2.FindString(word) != "1080" && time2.FindString(word) != "2160" {
			if strings.Contains(word, "-") {
				for _, a := range word {
					year, _ = strconv.Atoi(time2.FindString(strconv.Itoa(int(a))))
					break
				}
			} else {
				year, _ = strconv.Atoi(time2.FindString(word))
			}
		}
	}

	log.TLogln("YEAR:", year)

	for i, word := range nameMass {
		for _, word2 := range tor_q {
			if strings.EqualFold(word, word2) {
				for l := i; l < len(nameMass); l++ {
					nameMass[l] = ""
				}
				nameMass = nameMass[:i]
			}
		}
		if len(time2.FindString(word)) > 0 {
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
				en_rus = 1
				break
			}
		}
		if en_rus == 1 {
			nameMassNew = strings.Trim(nameMassNew, " ")
		} else {
			for _, word2 := range nameMass {
				nameMassNew = word2
				en_rus = 0
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

	e := getUtil(nameMassNew, tv, movie, year)
	if e != "" {
		return e
	} else {
		e = getUtil(nameMassNew2, tv, movie, year)
		if e != "" {
			return e
		} else {
			e = getUtil(nameMassNew, tv, movie, 0)
			if e != "" {
				return e
			} else {
				e = getUtil(nameMassNew2, tv, movie, 0)
				if e != "" {
					return e
				} else {
					return ""
				}
			}
		}
	}
}
