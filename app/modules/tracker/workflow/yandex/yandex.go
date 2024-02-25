package yandex

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httputil"
	"strconv"

	"go.uber.org/zap"

	"ru/kovardin/getapp/app/modules/tracker/models"
	"ru/kovardin/getapp/pkg/database"
	"ru/kovardin/getapp/pkg/logger"
)

type Yandex struct {
	log         *logger.Logger
	conversions *database.Repository[models.Conversion]
	trackers    *database.Repository[models.Tracker]
	client      *http.Client
}

func New(
	log *logger.Logger,
	conversions *database.Repository[models.Conversion],
	trackers *database.Repository[models.Tracker],
) *Yandex {
	return &Yandex{
		log:         log,
		conversions: conversions,
		trackers:    trackers,
		client:      &http.Client{},
	}
}

func (u *Yandex) Execute(ctx context.Context, name string) (string, error) {
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

	return "uploaded yandex tracker data", nil
}

func (u *Yandex) process(tracker models.Tracker) {
	conversions, err := u.conversions.Find(database.Condition{
		In: map[string]any{
			"fire":       false,
			"partner":    models.PartnerYadirect,
			"tracker_id": tracker.ID,
		},
	})
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

func (u *Yandex) yadirect(items []models.Conversion, yaurl, yatoken string) error {
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

	_ = dump // debug here

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

	_ = dump // debug here

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

func (u *Yandex) fire(items []models.Conversion) {
	for i := range items {
		func(item models.Conversion) {
			item.Fire = true
			if err := u.conversions.Save(&item); err != nil {
				u.log.Error("error on fire conversion", zap.Error(err))
			}
		}(items[i])
	}
}
