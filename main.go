package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cgi"
	"net/url"
	"os"
	"text/template"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	cgi.Serve(http.HandlerFunc(handler))
	//http.ListenAndServe(":8081", http.HandlerFunc(handler))
}

type Data struct {
	Git  string
	Name string
}

func handler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.New("todos").Parse("<!DOCTYPE html><html><head><meta http-equiv=\"refresh\" content=\"5; url={{ .Git }}\"><meta name=\"go-import\" content=\"{{ .Name }} git {{ .Git }}\"></head><body>Redirecting...</body></html>")
	var tpl bytes.Buffer
	t.Execute(&tpl, Data{Git: requestAPI(r.URL.Path), Name: r.Host + r.URL.Path})
	w.Write(tpl.Bytes())
}

type AirtableFields struct {
	Name string `json:"name"`
	Git  string `json:"git"`
}
type AirtableRecord struct {
	Id          string         `json:"id"`
	CreatedTime string         `json:"createdTime"`
	Fields      AirtableFields `json:"fields"`
}
type AirtableResp struct {
	Records []AirtableRecord `json:"records"`
}

func requestAPI(path string) string {
	client := http.Client{
		Timeout: time.Second * 2,
	}
	url := fmt.Sprintf("https://api.airtable.com/v0/%s/%s?maxRecords=3&view=%s&filterByFormula=%s", os.Getenv("BASE_ID"), os.Getenv("TABLE_NAME"), url.PathEscape("Grid view"), url.PathEscape(fmt.Sprintf("{name} = '%s'", path)))
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ""
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("API_KEY")))
	res, err := client.Do(req)
	if err != nil {
		return ""
	}
	if res.Body != nil {
		defer res.Body.Close()
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return ""
	}

	respData := AirtableResp{}
	err = json.Unmarshal(body, &respData)
	if err != nil {
		return ""
	}

	if len(respData.Records) > 0 {
		return respData.Records[0].Fields.Git
	}

	return ""
}
