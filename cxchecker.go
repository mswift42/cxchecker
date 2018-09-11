package cxcecker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"
)

func init() {
	http.HandleFunc("/querycx", getResults)
}

type Result struct {
	Response struct {
		Data struct {
			Boxes []struct {
				BoxName   string  `json:"boxName"`
				BoxId     string  `json:"boxId"`
				SellPrice float64 `json:"sellPrice"`
				ImageUrls struct {
					Medium string `json:"large"`
				} `json:"imageUrls"`
			} `json:"boxes"`
		} `json:"data"`
	} `json:"response"`
}

func parseResults(r *http.Response) ([]*QueryResult, error) {
	var results []*QueryResult
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	var result Result

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	boxes := result.Response.Data.Boxes
	for _, i := range boxes {
		results = append(results, newQueryResult(i.BoxName,
			i.ImageUrls.Medium, i.SellPrice, "", i.BoxId))
	}
	return results, nil
}

func productUrl(id string) string {
	return "https://uk.webuy.com/product-detail?id=" + id
}

func newQueryResult(title, thumbnail string, price float64, description, url string) *QueryResult {
	produrl := productUrl(url)
	thumb := strings.Replace(thumbnail, " ", "", -1)
	return &QueryResult{title, thumb, price, "", produrl}
}

type QueryResult struct {
	Title       string  `json:"title"`
	Thumbnail   string  `json:"thumbnail"`
	Price       float64 `json:"price"`
	Description string  `json:"description"`
	URL         string  `json:"url"`
}

func getResults(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	location := r.FormValue("location")
	query := r.FormValue("query")
	url := "https://wss2.cex.uk.webuy.io/v3/boxes?q=" + query + "&storeIds=[" + location + "]" + "&firstRecord=1&count=50&sortBy=sellprice&sortOrder=desc"
	ctx := appengine.NewContext(r)
	client := urlfetch.Client(ctx)
	resp, err := client.Get(url)
	if err != nil {
		fmt.Fprintf(w, err.Error(), http.StatusInternalServerError)
	}
	res, err := parseResults(resp)
	if err != nil {
		fmt.Fprintf(w, err.Error(), http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(res)
}
