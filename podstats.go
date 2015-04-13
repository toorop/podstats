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
	"path/filepath"
	"sort"
)

var (
	DB *gorm.DB
)

// main launches HTTP server
func main() {
	if err := initDb(); err != nil {
		log.Fatalln(err)
	}

	router := httprouter.New()
	router.HandlerFunc("GET", "/ping", func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("pong"))
	})

	// Routes
	// Get podcast
	router.GET("/p/:podcast/:episode", wrapHandler(getEpisode))

	// Stats podcast
	router.GET("/s/:podcast", wrapHandler(showPodcastStats))

	// Server
	n := negroni.New(negroni.NewRecovery(), negroni.NewLogger())
	n.UseHandler(router)
	addr := fmt.Sprintf("127.0.0.1:3333")
	log.Fatalln(http.ListenAndServe(addr, n))
}

// Handlers
// getEpisode records hit and return episode
func getEpisode(w http.ResponseWriter, r *http.Request) {
	/*if err := RecordHit(r); err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}*/
	go RecordHit(r)
	//w.Write([]byte("OK"))
	location := `http://podcasts.toorop.fr/` + httpcontext.Get(r, "params").(httprouter.Params).ByName("podcast") + `/` + httpcontext.Get(r, "params").(httprouter.Params).ByName("episode")
	w.Header().Set("Location", location)
	w.WriteHeader(302)
}

func showPodcastStats(w http.ResponseWriter, r *http.Request) {
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
}

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
	return DB.AutoMigrate(&hit{}).Error
}

// getBasePath is a helper for retrieving app path
func getBasePath() string {
	p, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	return p
}
