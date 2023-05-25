package poster_tmdb

import (
	"regexp"
	"server/config"
	"server/log"
	"strings"
)

var link_search string = "/search/0/0/110/0/"
var tmdb_search string = "https://api.themoviedb.org/3/find/"
var tmdb_poster string = "https://image.tmdb.org/t/p/original"

func bodyget(in string) (string, error) {
	var str string
	var err error
	if strings.Contains(in, "rutor.lib") {
		str, err = GetNic(in, "", "")
	} else {
		str, err = Get(in)
	}
	return str, err
}

func GetPoster(name string) string {
	api_key, _ := config.ReadConfigParser("ApiKey")
	host, _ := config.ReadConfigParser("Host")
	if host == "" {
		host = "http://rutor.is"
	}
	if api_key == "" {
		return ""
	}
	link := host + link_search + name
	body, err := bodyget(link)
	if err != nil {
		return ""
	}
	gp, err := regexp.Compile("<a href=\"/torrent/([0-9]+)/([A-Za-z0-9)-_!':;~+=,.]+)\">")
	if err != nil {
		log.TLogln("Error compile regex %v", err)
	}
	out := gp.FindString(body)
	out = strings.ReplaceAll(out, "<a href=\"", "")
	out = strings.ReplaceAll(out, "\">", "")
	out = host + out
	body, err = bodyget(out)
	if err != nil {
		return ""
	}
	gp, err = regexp.Compile("<a href=\"http://www.imdb.com/title/([a-zA-Z0-9]+)/\"")
	if err != nil {
		log.TLogln("Error compile regex %v", err)
	}
	out = gp.FindString(body)
	out = strings.ReplaceAll(out, "/\"", "")
	out = strings.ReplaceAll(out, "\"", "")
	out = strings.ReplaceAll(out, "<a href=", "")
	out = strings.ReplaceAll(out, "http://www.imdb.com/title/", "")
	path := tmdb_search + out + "?api_key=" + api_key + "&external_source=imdb_id"
	body, err = bodyget(path)
	if err != nil {
		return ""
	}
	gp, err = regexp.Compile("\"poster_path\":\"([A-Za-z0-9)-_.]+)\"")
	if err != nil {
		log.TLogln("Error compile regex %v", err)
	}
	out = gp.FindString(body)
	out = strings.ReplaceAll(out, "\"poster_path\":\"", "")
	out = tmdb_poster + strings.ReplaceAll(out, "\"", "")
	return out
}
