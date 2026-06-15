package metadata

import (
	"encoding/xml"
	"unicode"
)

// ComicInfo represents the standard metadata structure for digital comics.
// Compliant with ComicInfo v2.0 and v2.1 (draft) standards.
// References:
// https://github.com/anansi-project/comicinfo/blob/main/drafts/v2.1.0/ComicInfo.xml
type ComicInfo struct {
	XMLName             xml.Name `xml:"ComicInfo"`
	XmlnsXsi            string   `xml:"xmlns:xsi,attr"`
	XmlnsXsd            string   `xml:"xmlns:xsd,attr"`
	Title               string   `xml:"Title,omitempty"`
	Series              string   `xml:"Series,omitempty"`
	OriginalTitle       string   `xml:"OriginalTitle,omitempty"`
	Number              string   `xml:"Number,omitempty"`
	Count               int      `xml:"Count,omitempty"`
	Volume              int      `xml:"Volume,omitempty"`
	AlternateSeries     string   `xml:"AlternateSeries,omitempty"`
	AlternateNumber     string   `xml:"AlternateNumber,omitempty"`
	AlternateCount      int      `xml:"AlternateCount,omitempty"`
	Summary             string   `xml:"Summary,omitempty"`
	Notes               string   `xml:"Notes,omitempty"`
	Year                int      `xml:"Year,omitempty"`
	Month               int      `xml:"Month,omitempty"`
	Day                 int      `xml:"Day,omitempty"`
	Writer              string   `xml:"Writer,omitempty"`
	Penciller           string   `xml:"Penciller,omitempty"`
	Inker               string   `xml:"Inker,omitempty"`
	Colorist            string   `xml:"Colorist,omitempty"`
	Letterer            string   `xml:"Letterer,omitempty"`
	CoverArtist         string   `xml:"CoverArtist,omitempty"`
	Editor              string   `xml:"Editor,omitempty"`
	Translator          string   `xml:"Translator,omitempty"`
	Publisher           string   `xml:"Publisher,omitempty"`
	Imprint             string   `xml:"Imprint,omitempty"`
	Genre               string   `xml:"Genre,omitempty"`
	Tags                string   `xml:"Tags,omitempty"`
	Web                 string   `xml:"Web,omitempty"`
	PageCount           int      `xml:"PageCount,omitempty"`
	LanguageISO         string   `xml:"LanguageISO,omitempty"`
	Format              string   `xml:"Format,omitempty"`
	BlackAndWhite       string   `xml:"BlackAndWhite,omitempty"`
	Manga               string   `xml:"Manga,omitempty"`
	Characters          string   `xml:"Characters,omitempty"`
	Teams               string   `xml:"Teams,omitempty"`
	Locations           string   `xml:"Locations,omitempty"`
	ScanInformation     string   `xml:"ScanInformation,omitempty"`
	StoryArc            string   `xml:"StoryArc,omitempty"`
	SeriesGroup         string   `xml:"SeriesGroup,omitempty"`
	AgeRating           string   `xml:"AgeRating,omitempty"`
	CommunityRating     float64  `xml:"CommunityRating,omitempty"`
	MainCharacterOrTeam string   `xml:"MainCharacterOrTeam,omitempty"`
	Review              string   `xml:"Review,omitempty"`
	GTIN                string   `xml:"GTIN,omitempty"`
	CustomElements      []any    `xml:",any"`
}

type CustomTag struct {
	XMLName xml.Name
	Value   string `xml:",chardata"`
}

// SetMangaType maps friendly types ("漫画", "小说") to ComicInfo standards
func (c *ComicInfo) SetMangaType(mangaType string) {
	if mangaType == "小说" {
		c.Manga = "No"
	} else {
		c.Manga = "YesAndRightToLeft"
	}
}

// SetCustomProviderTag injects a provider-specific URL tag
func (c *ComicInfo) SetCustomProviderTag(tagName, url string) {
	if url == "" || tagName == "" {
		return
	}

	tag := CustomTag{
		XMLName: xml.Name{Local: tagName},
		Value:   url,
	}

	// Remove existing same-named tags
	var cleaned []any
	for _, el := range c.CustomElements {
		if ct, ok := el.(CustomTag); ok && ct.XMLName.Local == tagName {
			continue
		}
		cleaned = append(cleaned, el)
	}
	c.CustomElements = append(cleaned, tag)
}

// SetGTIN sets and normalizes the GTIN (ISBN) field
func (c *ComicInfo) SetGTIN(gtin string) {
	c.GTIN = NormalizeISBN13(gtin)
}

// NormalizeISBN13 cleans a string and tries to ensure it's a valid ISBN-13 format
func NormalizeISBN13(input string) string {
	// Clean all non-digit/non-X characters
	var cleaned []rune
	for _, r := range input {
		if unicode.IsDigit(r) || unicode.ToUpper(r) == 'X' {
			cleaned = append(cleaned, unicode.ToUpper(r))
		}
	}
	
	s := string(cleaned)
	if len(s) == 10 {
		// Basic ISBN-10 to 13 conversion (prepend 978, simple version without checksum recalc for now)
		return "978" + s
	}
	return s
}
