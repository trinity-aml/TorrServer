package api

import (
	"net/http"
	poster_tmdb "server/poster"
	"strings"

	"server/log"
	set "server/settings"
	"server/torr"
	"server/web/api/utils"

	"github.com/gin-gonic/gin"
)

var tor_q = []string{
	"Blu-ray",
	"BDRemux",
	"BDRip",
	"HDRip",
	"WEB-DL",
	"WEB-DLRip",
	"HDTV",
	"HDTVRip",
	"DVD9",
	"DVD5",
	"DVDRip",
	"DVDScr",
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
	"576i",
	"480p",
}

// torrentUpload godoc
//
//	@Summary		Add .torrent file
//	@Description	Only one file support.
//
//	@Tags			API
//
//	@Param			file	formData	file	true	"Torrent file to insert"
//	@Param			save	formData	string	false	"Save to DB"
//	@Param			title	formData	string	false	"Torrent title"
//	@Param			category	formData	string	false	"Torrent category"
//	@Param			poster	formData	string	false	"Torrent poster"
//	@Param			data	formData	string	false	"Torrent data"
//
//	@Accept			multipart/form-data
//
//	@Produce		json
//	@Success		200	{object}	state.TorrentStatus	"Torrent status"
//	@Router			/torrent/upload [post]
func torrentUpload(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	defer form.RemoveAll()

	save := len(form.Value["save"]) > 0
	title := ""
	if len(form.Value["title"]) > 0 {
		title = form.Value["title"][0]
	}
	category := ""
	if len(form.Value["category"]) > 0 {
		category = form.Value["category"][0]
	}
	poster := ""
	if len(form.Value["poster"]) > 0 {
		poster = form.Value["poster"][0]
	}
	data := ""
	if len(form.Value["data"]) > 0 {
		data = form.Value["data"][0]
	}
	var tor *torr.Torrent
	for name, file := range form.File {
		log.TLogln("add .torrent", name)

		torrFile, err := file[0].Open()
		if err != nil {
			log.TLogln("error upload torrent:", err)
			continue
		}
		defer torrFile.Close()

		spec, err := utils.ParseFile(torrFile)
		if err != nil {
			log.TLogln("error upload torrent:", err)
			continue
		}

		tor, err = torr.AddTorrent(spec, title, poster, data, category)

		if tor.Data != "" && set.BTsets.EnableDebug {
			log.TLogln("torrent data:", tor.Data)
		}
		if tor.Category != "" && set.BTsets.EnableDebug {
			log.TLogln("torrent category:", tor.Category)
		}

		if err != nil {
			log.TLogln("error upload torrent:", err)
			continue
		}

		go func() {
			if !tor.GotInfo() {
				log.TLogln("error add torrent:", "timeout connection torrent")
				return
			}

			if tor.Title == "" {
				tor.Title = tor.Name()
			}
			log.TLogln(spec.DisplayName)

			if tor.Poster == "" {
				if strings.Contains(tor.Title, ".") {
					tor.Title = strings.ReplaceAll(tor.Title, "_", ".")
					nameMass := strings.Split(tor.Title, ".")
					for i, word := range nameMass {
						for _, word2 := range tor_q {
							if word == word2 {
								if i == len(nameMass)-1 {
									break
								}
								if nameMass[i+1] == "2160p" || nameMass[i+1] == "1080p" || nameMass[i+1] == "720p" || nameMass[i+1] == "1080i" {
									for k := i + 2; k < len(nameMass); k++ {
										nameMass[k] = ""
									}
									nameMass = nameMass[:i+2]
								} else {
									for l := i + 1; l < len(nameMass); l++ {
										nameMass[l] = ""
									}
									nameMass = nameMass[:i+1]
								}
								break
							}
						}
					}
					tor.Title = strings.Join(nameMass, " ")
					log.TLogln("Title: ", tor.Title)
				}

				tor.Title = strings.Trim(tor.Title, " ")
				log.TLogln("Title: ", tor.Title)
				tor_poster := poster_tmdb.GetPoster(tor.Title)
				if tor_poster != "" {
					tor.Poster = tor_poster
				}
			}

			if save {
				torr.SaveTorrentToDB(tor)
			}
		}()

		break
	}
	c.JSON(200, tor.Status())
}
