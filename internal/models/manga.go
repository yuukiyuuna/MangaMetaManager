package models

import "gorm.io/gorm"

type MangaBase struct {
	Title           string `json:"title"`
	OriginalTitle   string `json:"originalTitle"`
	Series          string `json:"series"`
	Author          string `json:"author"`
	Translator      string `json:"translator"`
	Publisher       string `json:"publisher"`
	Genre           string `json:"genre"`
	Tags            string `json:"tags"`
	Summary         string `json:"summary"`
	Year            int    `json:"year"`
	Month           int    `json:"month"`
	Day             int    `json:"day"`
	Web             string `json:"web"`
	Language        string `json:"language"`
	Type            string `json:"type" gorm:"index"`      // 漫画 or 小说
	AgeRating       string `json:"ageRating"`
	Status          string `json:"status" gorm:"index"`    // e.g., Scraped, Missing
	LastError       string `json:"lastError"`
	SeriesGroup     string `json:"seriesGroup"`
	AlternateSeries string `json:"alternateSeries"`
	AlternateNumber string `json:"alternateNumber"`
	StoryArc        string `json:"storyArc"`
}

type MangaSeries struct {
	gorm.Model
	MangaBase
	Path            string      `json:"path" gorm:"uniqueIndex"`
	AlternateSeries string      `json:"alternateSeries"`
	Books           []MangaBook `json:"books" gorm:"foreignKey:SeriesID;constraint:OnDelete:CASCADE"`
}

type MangaBook struct {
	gorm.Model
	MangaBase
	SeriesID        uint   `json:"seriesId" gorm:"index"`
	Path            string `json:"path" gorm:"uniqueIndex"`
	Filename        string `json:"filename"`
	Number          string `json:"number"`
	Volume          int    `json:"volume"`
	PageCount       int    `json:"pageCount"`
	Characters      string `json:"characters"`
	Teams           string `json:"teams"`
}


type LibraryFolder struct {
	gorm.Model
	Path string `json:"path" gorm:"uniqueIndex"`
}
