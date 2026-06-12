package models

import "gorm.io/gorm"

type MangaSeries struct {
	gorm.Model
	Path            string      `json:"path" gorm:"uniqueIndex"`
	Title           string      `json:"title"`
	OriginalTitle   string      `json:"originalTitle"`
	Series          string      `json:"series"`
	AlternateSeries string      `json:"alternateSeries"`
	Author          string      `json:"author"`
	Publisher       string      `json:"publisher"`
	Genre           string      `json:"genre"`
	Summary         string      `json:"summary"`
	Year            int         `json:"year"`
	Month           int         `json:"month"`
	Day             int         `json:"day"`
	Web             string      `json:"web"`
	Manga           string      `json:"manga"` // Yes/No
	AgeRating       string      `json:"ageRating"`
	Status          string      `json:"status"` // e.g., Scraped, Missing
	Books           []MangaBook `json:"books" gorm:"foreignKey:SeriesID"`
}

type MangaBook struct {
	gorm.Model
	SeriesID        uint   `json:"seriesId"`
	Path            string `json:"path" gorm:"uniqueIndex"`
	Filename        string `json:"filename"`
	Title           string `json:"title"`
	OriginalTitle   string `json:"originalTitle"`
	Series          string `json:"series"`
	Number          string `json:"number"`
	Author          string `json:"author"`
	Publisher       string `json:"publisher"`
	Genre           string `json:"genre"`
	Volume          int    `json:"volume"`
	Year            int    `json:"year"`
	Month           int    `json:"month"`
	Day             int    `json:"day"`
	Web             string `json:"web"`
	PageCount       int    `json:"pageCount"`
	Manga           string `json:"manga"` // Yes/No
	AgeRating       string `json:"ageRating"`
	Characters      string `json:"characters"`
	Teams           string `json:"teams"`
	Status          string `json:"status"`
	Summary         string `json:"summary"`
}


type LibraryFolder struct {
	gorm.Model
	Path string `json:"path" gorm:"uniqueIndex"`
}
