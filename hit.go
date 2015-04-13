package main

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/nbio/httpcontext"
	"github.com/oschwald/geoip2-golang"
	"net"
	"net/http"
	//"strings"
	//"errors"
	"time"
)

type hit struct {
	Id          uint64 `gorm:"primary_key"`
	Date        time.Time
	Podcast     string
	Episode     string
	Ip          string
	Host        string
	UserAgent   string
	Referer     string
	Country     string
	CountryCode string
	City        string
	TimeZone    string
	Coordinates string
}

func RecordHit(r *http.Request) error {
	var fromHost string
	//fromIp := strings.SplitN(r.RemoteAddr, ":", 2)[0]
	fromIp := r.Header.Get("X-Real-IP")
	//return errors.New(fromIp)
	//fromIp = "109.190.73.59"
	//fromIp = "88.178.118.205"
	hosts, err := net.LookupAddr(fromIp)
	if err == nil && len(hosts) > 0 {
		fromHost = hosts[0]
	} else {
		fromHost = ""
	}
	// geoip
	var country, countryCode, city, timeZone, coordinates string
	db, err := geoip2.Open(getBasePath() + "/GeoLite2-City.mmdb")
	if err == nil {
		defer db.Close()
		// If you are using strings that may be invalid, check that ip is not nil
		ip := net.ParseIP(fromIp)
		record, err := db.City(ip)
		if err == nil {
			city = record.City.Names["en"]
			country = record.Country.Names["en"]
			countryCode = record.Country.IsoCode
			timeZone = record.Location.TimeZone
			coordinates = fmt.Sprintf("%f, %f", record.Location.Latitude, record.Location.Longitude)
		}
	}

	return DB.Create(hit{
		Date:        time.Now(),
		Podcast:     httpcontext.Get(r, "params").(httprouter.Params).ByName("podcast"),
		Episode:     httpcontext.Get(r, "params").(httprouter.Params).ByName("episode"),
		Ip:          fromIp,
		Host:        fromHost,
		UserAgent:   r.UserAgent(),
		Referer:     r.Referer(),
		Country:     country,
		CountryCode: countryCode,
		City:        city,
		TimeZone:    timeZone,
		Coordinates: coordinates,
	}).Error

}
