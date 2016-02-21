package content

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
    "github.com/railsagainstignorance/alignment/Godeps/_workspace/src/github.com/joho/godotenv"
    "time"
    "strconv"
)

type Article struct {
    SiteUrl string
    Uuid    string
    Title   string
    Author  string
    Excerpt string
    Body    string
    PubDateString string
    PubDate time.Time
}

type SearchResponse struct {
    SiteSearchUrl string
    NumItems      int
    NumMatches    int
    Articles      *[]*Article
    QueryString   string
    SearchRequest *SearchRequest
}

const longformPubDate     = "2015-05-22T18:06:49Z"
const baseUriCapi         = "http://api.ft.com/content/items/v1/"
const baseUriSapi         = "http://api.ft.com/content/search/v1"

const sapiKeyEnvParamName = "SAPI_KEY"

func getApiKey() string {
    key := os.Getenv(sapiKeyEnvParamName)

    if key == "" {
        fmt.Println("content: getApiKey: no such env params: ", sapiKeyEnvParamName)
    }

    return key
}

var apiKey = getApiKey()

func getCapiArticleJsonBody(uuid string) (*[]byte) {
    url := baseUriCapi + uuid + "?apiKey=" + apiKey
    fmt.Println("content: getCapiArticleJsonBody: url=", url)

    req, err := http.NewRequest("GET", url, nil)
    req.Header.Set("Content-Type", "application/json")
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()
    fmt.Println("content: getCapiArticleJsonBody: response Status:", resp.Status)
    jsonBody, _ := ioutil.ReadAll(resp.Body)

    return &jsonBody
}

func parseCapiArticleJsonBody(jsonBody *[]byte) (*Article) {

    var data interface{}
    json.Unmarshal(*jsonBody, &data)

    aSiteUrl := ""
    aUuid    := ""
    aTitle   := ""
    aAuthor  := ""
    aBody    := ""
    aPubDateString := ""
    var aPubDate time.Time

    if item, ok := data.(map[string]interface{})[`item`].(map[string]interface{}); ok {
        if uuid, ok := item["id"].(string); ok {
            aUuid    = uuid
            aSiteUrl = "http://www.ft.com/cms/s/2/" + uuid + ".html"

            if titleOuter, ok := item["title"].(map[string]interface{}); ok {
                if title, ok := titleOuter["title"].(string); ok {
                    aTitle = title
                }
            }

            if bodyOuter, ok := item["body"].(map[string]interface{}); ok {
                if body, ok := bodyOuter["body"].(string); ok {
                    aBody = body
                }
            }

            if lifecycle, ok := item["lifecycle"].(map[string]interface{}); ok {
                if lastPublishDateTime, ok := lifecycle["lastPublishDateTime"].(string); ok {
                    aPubDateString = lastPublishDateTime
                }
            }
        }
    }

    if aPubDateString != "" {
        if pd,err := time.Parse(longformPubDate, aPubDateString); err == nil {
            aPubDate = pd
        } else {
            fmt.Println("WARNING: content.Article: could not parse pubdate string: ", aPubDateString, ", for UUID=", aUuid)
        }
    }

    article := Article{
        SiteUrl: aSiteUrl,
        Uuid:    aUuid,
        Title:   aTitle,
        Author:  aAuthor,
        Body:    aBody,
        PubDateString: aPubDateString,
        PubDate: aPubDate,
    }

    return &article
}

var uuidJsonBodyCache = map[string]*[]byte{}

func GetArticle(uuid string) (*Article) {
    var jsonBody *[]byte

    if _, ok := uuidJsonBodyCache[uuid]; ok {
        fmt.Println("capi.GetArticle: cache hit: uuid=", uuid)
        jsonBody = uuidJsonBodyCache[uuid]
    } else {
        fmt.Println("capi.GetArticle: cache miss: uuid=", uuid)
        jsonBody = getCapiArticleJsonBody(uuid)
        uuidJsonBodyCache[uuid] = jsonBody
    }

    article := parseCapiArticleJsonBody( jsonBody )

    return article
}

// now same for SAPI stuff
type SearchRequest struct {
    QueryType         string // e.g "keyword", "title", "topicXYZ", etc
    QueryText         string // e.g. "tail spin" or "\"tail spin\""
    MaxResponses      int
    MaxDurationMillis int
    SearchOnly        bool // i.e. don't bother looking up articles
    QueryStringValue  string
}

func constructQueryString(sr SearchRequest ) string {

    var queryString string

    // does not yet handle "page"
    switch sr.QueryType {
    case "keyword", "":
        queryString = sr.QueryText
    default:
        queryString = sr.QueryType + `:\"` + sr.QueryText + `\"`
    }

    return queryString
}

func convertStringsToQuotedCSV( sList []string ) string {
    sListQuoted := []string{}

    for _,s := range sList {
        sListQuoted = append( sListQuoted, `"` + s + `"`)
    }

    sCsv := strings.Join( sListQuoted, ", " )
    return sCsv
}

func getSapiResponseJsonBody(queryString string) ([]byte) {
    url := "http://api.ft.com/content/search/v1?apiKey=" + apiKey

    fmt.Println("sapi: getSapiResponseJsonBody: queryString:", queryString)
    curationsString := convertStringsToQuotedCSV( []string{ "ARTICLES", "BLOGS" } )
    aspectsString   := convertStringsToQuotedCSV( []string{ "title", "location", "summary", "lifecycle", "metadata" } )
    maxResults      := 100

    jsonStr         := []byte(
        `{` +
            `"queryString" : "` + queryString + `",` +
            `"queryContext" : {"curations" : [ ` + curationsString + ` ]},` +
            `"resultContext" : {` + 
                `"maxResults" : "` + strconv.Itoa(maxResults) + `",` + 
                `"offset" : "0",` + 
                `"aspects" : [ ` + aspectsString + `],` + 
                `"sortOrder": "DESC",` + 
                `"sortField": "lastPublishDateTime"` +
            `}` + 
        `}` )

    req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
    req.Header.Set("Content-Type", "application/json")
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    if resp.Status != "200" {
        fmt.Println("WARNING: content: getSapiResponseJsonBody: response Status:", resp.Status)
    }

    jsonBody, _ := ioutil.ReadAll(resp.Body)

    return jsonBody
}

func main() {
    godotenv.Load()
    uuid := "b57fee24-cb3c-11e5-be0b-b7ece4e953a0"

    article := GetArticle( uuid )
    fmt.Println("main: article.Title=", article.Title)
}
