package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/upthebuzzard/alignment/Godeps/_workspace/src/github.com/joho/godotenv"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	// "reflect"
	"sort"
	"strings"
)

func alignFormHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("align.html")
	t.Execute(w, nil)
}

type PhraseBits struct {
	Before      string
	Common      string
	After       string
	Excerpt     string
	Title       string
	LocationUri string
}

type ByBeforeBit []PhraseBits

func (s ByBeforeBit) Len() int {
	return len(s)
}
func (s ByBeforeBit) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ByBeforeBit) Less(i, j int) bool {
	return len(s[i].Before) > len(s[j].Before)
}

type AlignParams struct {
	Text      string
	Source    string
	MaxIndent int
	Phrases   []PhraseBits
}

func getSapiResponseJsonBody(text string, titleOnly bool) []byte {
	sapiKey := os.Getenv("SAPI_KEY")
	url := "http://api.ft.com/content/search/v1?apiKey=" + sapiKey
	queryString := `\"` + text + `\"`
	if titleOnly {
		queryString = "title:" + queryString
	}

	var jsonStr = []byte(`{"queryString": "` + queryString + `","queryContext" : {"curations" : [ "ARTICLES", "BLOGS" ]},  "resultContext" : {"maxResults" : "100", "offset" : "0", "aspects" : [ "title", "location", "summary", "lifecycle", "metadata"], "sortOrder": "DESC", "sortField": "lastPublishDateTime" } } }`)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	jsonBody, _ := ioutil.ReadAll(resp.Body)

	return jsonBody
}

func alignHandler(w http.ResponseWriter, r *http.Request) {
    defaultText := os.Getenv("DEFAULT_TEXT")
    if defaultText=="" {
        defaultText = "has its own"
    }
	text := r.FormValue("text")
    if text=="" {
        text = defaultText
    }
	source := r.FormValue("source")
    if source=="" {
        source="any"
    }
	titleOnly := source == "title-only"

	jsonBody := getSapiResponseJsonBody(text, titleOnly)

	// locate results
	var data interface{}
	json.Unmarshal(jsonBody, &data)
	results := data.(map[string]interface{})[`results`].([]interface{})[0].(map[string]interface{})[`results`].([]interface{})

	// loop over results to pick out relevant fields
	textLength := len(text)
	phrases := []PhraseBits{}
	maxIndent := 0

	for _, r := range results {
		excerpt := r.(map[string]interface{})["summary"].(map[string]interface{})["excerpt"].(string)
		title := r.(map[string]interface{})["title"].(map[string]interface{})["title"].(string)
		locationUri := r.(map[string]interface{})["location"].(map[string]interface{})["uri"].(string)

		phrase := excerpt
		if titleOnly {
			phrase = title
		}

		if indent := strings.Index(phrase, text); indent > -1 {
			bits := &PhraseBits{
				Before:      phrase[0:indent],
				Common:      text,
				After:       phrase[indent+textLength : len(phrase)],
				Excerpt:     excerpt,
				Title:       title,
				LocationUri: locationUri,
			}
			phrases = append(phrases, *bits)

			if maxIndent < indent {
				maxIndent = indent
			}
		}
	}

	// because it looks better this way
	sort.Sort(ByBeforeBit(phrases))

	p := &AlignParams{Text: text, Source: source, MaxIndent: maxIndent, Phrases: phrases}
	t, _ := template.ParseFiles("aligned.html")
	t.Execute(w, p)
}

func main() {
	godotenv.Load()
	port := os.Getenv("PORT")

	http.HandleFunc("/", alignFormHandler)
	http.HandleFunc("/align", alignHandler)
	http.ListenAndServe(":"+string(port), nil)
}
