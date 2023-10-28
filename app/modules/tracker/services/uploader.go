package services

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httputil"
	"strconv"
	"time"

	"go.uber.org/zap"

	"ru/kovardin/getapp/app/modules/tracker/models"
	"ru/kovardin/getapp/pkg/database"
)

type Uploader struct {
	log         *zap.Logger
	conversions *database.Repository[models.Conversion]
	trackers    *database.Repository[models.Tracker]
	client      *http.Client
	period      time.Duration
}

func NewUploader(log *zap.Logger, conversions *database.Repository[models.Conversion], trackers *database.Repository[models.Tracker]) *Uploader {
	return &Uploader{
		log:         log,
		conversions: conversions,
		trackers:    trackers,
		client:      &http.Client{},
		period:      time.Hour * 1,
	}
}

func (u *Uploader) Start() {
	ticker := time.NewTicker(u.period)
	go func() {
		for ; true; <-ticker.C {
			tt, err := u.trackers.Find(database.Condition{
				In: map[string]any{
					"active": true,
				},
			})
			if err != nil {
				u.log.Error("error on getting trackers", zap.Error(err))
			}

			for _, t := range tt {
				u.process(t)
			}
		}

		//for range ticker.C {
		//	u.process()
		//}
	}()

}

func (u *Uploader) Stop() {

}

func (u *Uploader) process(tracker models.Tracker) {

	func() {
		conversions, err := u.search(models.PartnerYadirect, tracker.ID)
		if err != nil {
			u.log.Error("error on get yadirect conversions", zap.Error(err))
			return
		}

		if len(conversions) == 0 {
			return
		}

		if err := u.yadirect(conversions, tracker.YandexMetricaTracker, tracker.YandexToken); err != nil {
			u.log.Error("error on upload conversions to yadirect", zap.Error(err))
			return
		}

		u.fire(conversions)
	}()

	func() {
		conversions, err := u.search(models.PartnerVkads, tracker.ID)
		if err != nil {
			u.log.Error("error on get vkads conversions", zap.Error(err))
			return
		}

		if len(conversions) == 0 {
			return
		}

		if err := u.vkads(conversions, tracker.VkTracker); err != nil {
			u.log.Error("error on upload conversions to vkads", zap.Error(err))
			return
		}

		u.fire(conversions)
	}()
}

func (u *Uploader) search(partner string, tracker uint) ([]models.Conversion, error) {
	cc, err := u.conversions.Find(database.Condition{
		In: map[string]any{
			"fire":       false,
			"partner":    partner,
			"tracker_id": tracker,
		},
	})

	return cc, err
}

type UploadError struct {
	ErrorType string `json:"error_type"`
	Message   string `json:"message"`
}
type UploadResponse struct {
	Errors  []UploadError `json:"errors"`
	Code    int           `json:"code"`
	Message string        `json:"message"`
}

func (u *Uploader) yadirect(items []models.Conversion, yaurl, yatoken string) error {
	// create all body
	body := &bytes.Buffer{}

	// create writer for body
	writer := multipart.NewWriter(body)

	// create conversions file data
	data := &bytes.Buffer{}

	// create writer for conversions file data
	file := csv.NewWriter(data)

	if err := file.Write([]string{
		//"ClientId",
		"Yclid",
		"Target",
		"DateTime",
	}); err != nil {
		u.log.Error("error on write csv title", zap.Error(err))
		return err
	}

	for _, item := range items {
		if err := file.Write([]string{
			//item.ClientId,
			item.Yclid,
			"app_install",
			strconv.Itoa(item.InstallTimestamp),
		}); err != nil {
			u.log.Error("error on write csv row", zap.Error(err))
			continue
		}
	}

	file.Flush()

	part, _ := writer.CreateFormFile("file", "file.csv")
	if _, err := io.Copy(part, data); err != nil {
		u.log.Error("error on copy file to part", zap.Error(err))
		return err
	}

	if err := writer.Close(); err != nil {
		u.log.Error("error on close file writer", zap.Error(err))
		return err
	}

	request, _ := http.NewRequest("POST", yaurl, body)
	request.Header.Add("Authorization", "OAuth "+yatoken)
	request.Header.Add("Content-Type", writer.FormDataContentType())

	dump, err := httputil.DumpRequest(request, true)
	if err != nil {
		u.log.Error("error on dump request", zap.Error(err))
		return err
	}
	fmt.Printf("reques:\n%s\n", dump)
	//fmt.Printf("%q", dump)

	resp, err := u.client.Do(request)
	if err != nil {
		u.log.Error("error on do request", zap.Error(err))
		return err
	}

	dump, err = httputil.DumpResponse(resp, true)
	if err != nil {
		u.log.Error("error on dump response", zap.Error(err))
		return err
	}
	fmt.Printf("response:\n%s\n", dump)

	result := UploadResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		u.log.Error("error on parse response", zap.Error(err))
		return err
	}

	if len(result.Errors) != 0 {
		u.log.Error("error on upload file", zap.String("error", result.Message))
		return fmt.Errorf("error on upload file: %s", result.Message)
	}

	return nil
}

func (u *Uploader) vkads(items []models.Conversion, vkurl string) error {
	for _, item := range items {
		link := vkurl + item.RbClickid

		u.log.Warn("vkads link for check", zap.String("link", link))

		if _, err := u.client.Get(link); err != nil {
			u.log.Error("error on send vk pixel", zap.Error(err))
			continue
		}

	}
	return nil
}

func (u *Uploader) fire(items []models.Conversion) {
	for i := range items {
		func(item models.Conversion) {
			item.Fire = true
			if err := u.conversions.Save(&item); err != nil {
				u.log.Error("error on fire conversion", zap.Error(err))
			}
		}(items[i])
	}
}
