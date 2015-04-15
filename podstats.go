package main

import (
	"fmt"
	"github.com/codegangsta/negroni"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/julienschmidt/httprouter"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/nbio/httpcontext"
	"log"
	"net/http"
	"os"
	//"path"
	//"net"
	"encoding/json"
	"github.com/unrolled/render"
	"html/template"
	"path/filepath"
	//"sort"
	"strconv"
)

var (
	DB *gorm.DB
	R  *render.Render
)

// Commons data for template
type tplBase struct {
	Title       string
	Js          template.JS
	MoreScripts []string
}

// main launches HTTP server
func main() {
	if err := initDb(); err != nil {
		log.Fatalln(err)
	}

	R = render.New(render.Options{
		Layout: "layout",
	})

	router := httprouter.New()
	router.HandlerFunc("GET", "/ping", func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("pong"))
	})

	// Routes

	// Home
	router.GET("/", wrapHandler(handlerHome))

	// Get podcast
	router.GET("/p/:podcast/:episode", wrapHandler(getEpisode))

	// Podcast stats
	//router.GET("/s/:podcast", wrapHandler(showPodcastStats))

	// Admin: add podcast
	router.GET("/a/new", wrapHandler(newEpisode))
	router.POST("/a/add", wrapHandler(addEpisode))

	// Server
	n := negroni.New(negroni.NewRecovery(), negroni.NewStatic(http.Dir("public")))
	n.UseHandler(router)
	addr := fmt.Sprintf("127.0.0.1:3333")
	log.Fatalln(http.ListenAndServe(addr, n))
}

// jsonResponse represents a json reponse
type jsonResponse struct {
	Success bool   `json:"success"`
	Msg     string `json:"msg"`
}

// Handlers
// home
func handlerHome(w http.ResponseWriter, r *http.Request) {
	type data struct {
		Base     tplBase
		Episodes []episode
	}

	episodes := []episode{}
	// get all episode
	err := DB.Where("podcast = ?", "tmail").Order("episode desc").Find(&episodes).Error
	//rows, err := DB.Raw("select distinct episode from episodes where podcast = ?", "tmail").Rows()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	for i, ep := range episodes {
		c := 0
		// get stats for this episode
		row := DB.Raw("select count(distinct ip) from hits where podcast = ? and episode = ?", "tmail", ep.Episode).Row()
		err := row.Scan(&c)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}
		episodes[i].PlayCount = ep.CounterDiff + c
	}

	d := &data{
		Base: tplBase{
			Title:       "Podstats: Analytics for podcasters",
			Js:          ``,
			MoreScripts: []string{},
		},
		Episodes: episodes,
	}
	R.HTML(w, http.StatusOK, "index", d)
}

// getEpisode records hit and return episode
func getEpisode(w http.ResponseWriter, r *http.Request) {

	episodeNumber, err := strconv.ParseUint(httpcontext.Get(r, "params").(httprouter.Params).ByName("episode"), 10, 32)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))

		return
	}
	ep, err := GetEpisodeByPodcastEpisodeNumber(httpcontext.Get(r, "params").(httprouter.Params).ByName("podcast"), uint(episodeNumber))
	if err == gorm.RecordNotFound {
		http.NotFound(w, r)
		return
	}

	/*if err := RecordHit(r); err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}*/
	go RecordHit(r)
	w.Header().Set("Location", ep.Link)
	w.WriteHeader(302)
}

// admin add episode
func newEpisode(w http.ResponseWriter, r *http.Request) {
	type data struct {
		Base tplBase
	}

	d := &data{Base: tplBase{
		Title:       "Podstats: Add a new episode",
		Js:          ``,
		MoreScripts: []string{"add_episode.js"},
	},
	}

	R.HTML(w, http.StatusOK, "addEpisode", d)

}

// Add an episode (ajax)
func addEpisode(w http.ResponseWriter, r *http.Request) {
	// get body
	ep := episode{}
	// nil body
	if r.Body == nil {
		R.JSON(w, 422, "empty body")
		return
	}
	if err := json.NewDecoder(r.Body).Decode(&ep); err != nil {
		R.JSON(w, http.StatusOK, &jsonResponse{false, "unable to get JSON body. " + err.Error()})
		return
	}
	// create record in DB
	if err := ep.CreateInDb(); err != nil {
		R.JSON(w, http.StatusOK, &jsonResponse{false, err.Error()})
		return
	}

	R.JSON(w, http.StatusOK, &jsonResponse{true, "Episode Added"})
}

// showPodcastStats return podcast stats
/*func showPodcastStats(w http.ResponseWriter, r *http.Request) {
	// Episodes SELECT COUNT(DISTINCT episode) FROM hits WHERE podcast = ``
	var episodes sort.StringSlice
	episodes = []string{}
	var episode string
	//row := DB.Table("hits").Find("DISTINCT episode").Where("podcast=?", httpcontext.Get(r, "params").(httprouter.Params).ByName("podcast"))
	rows, err := DB.Raw("select distinct episode from hits where podcast = ?", httpcontext.Get(r, "params").(httprouter.Params).ByName("podcast")).Rows()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
	defer rows.Close()
	for rows.Next() {
		rows.Scan(&episode)
		episodes = append(episodes, episode)
	}

	sort.Sort(episodes)

	stats := make(map[string]int)
	var c int
	for _, ep := range episodes {
		row := DB.Raw("select count(distinct ip) from hits where podcast = ? and episode = ?", httpcontext.Get(r, "params").(httprouter.Params).ByName("podcast"), ep).Row()
		err := row.Scan(&c)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}
		stats[ep] = c
	}
	out := ""
	for ep, c := range stats {
		out += fmt.Sprintf("%s: %d lecture(s)\n", ep, c)
	}

	w.Write([]byte(out))
}*/

// wrapHandler puts httprouter.Params in query context
// in order to keep compatibily with http.Handler
func wrapHandler(h func(http.ResponseWriter, *http.Request)) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		httpcontext.Set(r, "params", ps)
		h(w, r)
	}
}

// DB

func initDb() error {
	db, err := gorm.Open("sqlite3", getBasePath()+"/podstats.db")
	if err != nil {
		return err
	}
	DB = &db
	DB.LogMode(false)

	if err = DB.AutoMigrate(&hit{}, &episode{}).Error; err != nil {
		return err
	}
	db.Model(&episode{}).AddUniqueIndex("idx_podcast_episode", "podcast", "episode")
	return nil
}

// getBasePath is a helper for retrieving app path
func getBasePath() string {
	p, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	return p
}
