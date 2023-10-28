package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httputil"
)

func main() {
	status()
}

func status() {
	r, _ := http.NewRequest("GET", "https://api-metrika.yandex.net/management/v1/counter/94618529/offline_conversions/uploading/642053281.", nil)
	r.Header.Add("Authorization", "OAuth y0_AgAAAAACdkHGAApcaAAAAADqoW6oqVIfHMrfSn-zC0s6IzhJkEZ1lfc")

	client := &http.Client{}
	resp, _ := client.Do(r)

	dump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("response:\n%s\n", dump)
}

func upload() {
	// https://yandex.ru/support/direct/statistics/url-tags.html
	// https://api-metrika.yandex.net/management/v1/counter/94618529/offline_conversions/upload?client_id_type=CLIENT_ID
	// app_install
	// get latest conversions
	// prepare for upload
	// upload

	// create all body
	body := &bytes.Buffer{}

	// create writer for body
	writer := multipart.NewWriter(body)

	// create conversions file data
	data := &bytes.Buffer{}

	// create writer for conversions file data
	file := csv.NewWriter(data)

	if err := file.Write([]string{
		"UserId",
		"ClientId",
		"Yclid",
		"Target",
		"DateTime",
	}); err != nil {
		log.Print(err)
	}

	if err := file.Write([]string{
		"1",
		"1",
		"1",
		"1",
		"1",
	}); err != nil {
		log.Print(err)
	}

	file.Flush()

	part, _ := writer.CreateFormFile("file", "data.csv")
	io.Copy(part, data)
	writer.Close()

	r, _ := http.NewRequest("POST", "http://example.com", body)
	r.Header.Add("Content-Type", writer.FormDataContentType())

	dump, err := httputil.DumpRequest(r, true)
	if err != nil {
		return
	}

	fmt.Printf("%s", dump)

	client := &http.Client{}
	client.Do(r)
}
