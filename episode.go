package main

type episode struct {
	Id          uint   `gorm:"primary_key"`
	Podcast     string `sql:"not null"; json:"podcast"`
	Episode     uint   `sql:"not null"; json:"episode"`
	Title       string `sql:"not null"; json:"title"`
	Link        string `sql:"not null"; json:"link"`
	CounterDiff int    `json: "counter_diff"`
	PlayCount   int    `json: "playcount"`
}

// CreateInDb create a new record in DB
func (e *episode) CreateInDb() error {
	return DB.Create(e).Error
}

// GetByPodcastEpisode returns an episode by its podcast name and episode number
func GetEpisodeByPodcastEpisodeNumber(podcast string, episodeNumber uint) (episode, error) {
	ep := episode{}
	err := DB.Where("podcast = ? and episode = ?", podcast, episodeNumber).Find(&ep).Error
	return ep, err
}
