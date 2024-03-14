package handlers

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"go.uber.org/zap"
	"net/http"
	"net/http/httputil"

	settings "ru/kovardin/getapp/app/modules/admin/models"
	"ru/kovardin/getapp/app/modules/billing/models"
	"ru/kovardin/getapp/pkg/database"
	"ru/kovardin/getapp/pkg/logger"
)

type Payload struct {
	Payment uint `json:"payment"`
}

type Confirm struct {
	secret   string
	log      *logger.Logger
	payments *database.Repository[models.Payment]
	settings *database.Repository[settings.Setting]
}

func NewConfirm(log *logger.Logger, payments *database.Repository[models.Payment], settings *database.Repository[settings.Setting]) *Confirm {
	return &Confirm{
		log:      log,
		payments: payments,
		settings: settings,
	}
}

func (c *Confirm) Hook(w http.ResponseWriter, r *http.Request) {
	data, err := httputil.DumpRequest(r, true)
	if err != nil {
		c.log.Error("error on dump http response", zap.Error(err))
	}

	c.log.Info("confirm http response", zap.ByteString("resp", data))

	key, err := c.settings.First(database.Condition{
		In: map[string]any{
			"key": "yoomoney_callback_secret",
		},
	})
	if err != nil {
		c.log.Error("error on get yoomoney secret from setting", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if key.Value == "" {
		c.log.Error("yoomoney secret from setting is empty")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		c.log.Error("error parse request", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	c.log.Info("parsed from", zap.Any("form", r.Form))

	check := r.Form.Get("notification_type") +
		"&" + r.Form.Get("operation_id") +
		"&" + r.Form.Get("amount") +
		"&" + r.Form.Get("currency") +
		"&" + r.Form.Get("datetime") +
		"&" + r.Form.Get("sender") +
		"&" + r.Form.Get("codepro") +
		"&" + key.Value +
		"&" + r.Form.Get("label")

	h := sha1.New()
	h.Write([]byte(check))
	hash := hex.EncodeToString(h.Sum(nil))

	if hash != r.Form.Get("sha1_hash") {
		c.log.Error("cant check hash",
			zap.String("income_hash", r.Form.Get("sha1_hash")),
			zap.String("counted_hash", hash))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	label, err := base64.StdEncoding.DecodeString(r.Form.Get("label"))
	if err != nil {
		c.log.Error("error on decode label", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	payload := Payload{}
	if err := json.Unmarshal(label, &payload); err != nil {
		c.log.Error("error on unmarshal payload", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := c.payments.Update(&models.Payment{}, "status", models.PaymentStatusConfirm, payload.Payment); err != nil {
		c.log.Error("error on update payment", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
