package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
)

const (
	GAME_ID_REGEX   = `href="/boardgame/(\d*)/`
	GAME_ID_URL     = "https://boardgamegeek.com/browse/boardgame/page/%d"
	GAME_RATING_URL = "https://www.boardgamegeek.com/xmlapi2/thing?id=%d&ratingcomments=1&page=%d"
)

func getGameIds() ([]string, error) {
	ids := make([]string, 500)

	//r, err := regexp.Compile("https:\\/\\/boardgamegeek\\.com\\/boardgame\\/(\\d*)\\/[\\w\\d-]*")
	r, err := regexp.Compile(GAME_ID_REGEX)
	if err != nil {
		panic(err)
	}
	for i := 1; i <= 5; i++ {
		url := fmt.Sprintf(GAME_ID_URL, i)
		response, err := http.Get(url)
		if err != nil {
			return ids, err
		}
		defer response.Body.Close()

		contents, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return ids, err
		}

		matches := r.FindAllStringSubmatch(string(contents), -1)
		ids = append(ids, matches[1]...)
	}

	return ids, nil
}

func getGameRatingPage(id int, page int) (*items, error) {
	data := items{}

	url := fmt.Sprintf(GAME_RATING_URL, id, page)
	log.Printf("Fetching %s", url)

	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	xmlContents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	err = xml.Unmarshal(xmlContents, &data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func getGameName(game item) string {
	for i := range game.Names {
		if game.Names[i].Type == "primary" {
			return game.Names[i].Name
		}
	}
	return "UNKNOWN"
}

func main() {
	gameIds, err := getGameIds()
	if err != nil {
		log.Fatal(err)
		return
	}

	for id := range gameIds {
		data, err := getGameRatingPage(id, 1)
		if err != nil {
			log.Println(err)
			continue
		}
		log.Println(data)

		if len(data.Items) == 0 {
			log.Println("No data found for id %d", id)
			continue
		}

		game := data.Items[0]
		gameName := getGameName(game)

		totalComments, err := strconv.ParseInt(game.Comments.TotalItems, 0, 64)
		if err != nil {
			log.Println(err)
			continue
		}

		log.Printf("%s has %d ratings", gameName, totalComments)
		for i := 2; int64(i*100) < totalComments; i++ {
			data, err = getGameRatingPage(id, i)
			if err != nil {
				log.Println(err)
				continue
			}
			log.Println(data)
		}
	}
}

type comment struct {
	XMLName  xml.Name `xml:"comment"`
	Username string   `xml:"username,attr"`
	Rating   string   `xml:"rating,attr"`
}

type comments struct {
	XMLName    xml.Name  `xml:"comments"`
	Comments   []comment `xml:"comment"`
	TotalItems string    `xml:"totalitems,attr"`
}

type name struct {
	XMLName xml.Name `xml:"name"`
	Name    string   `xml:"value,attr"`
	Type    string   `xml:"type,attr"`
}

type item struct {
	XMLName  xml.Name `xml:"item"`
	Names    []name   `xml:"name"`
	Comments comments `xml:"comments"`
}

type items struct {
	XMLName xml.Name `xml:"items"`
	Items   []item   `xml:"item"`
}
