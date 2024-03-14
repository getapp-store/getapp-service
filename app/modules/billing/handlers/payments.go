package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"go.uber.org/zap"

	"ru/kovardin/getapp/app/auth"
	settings "ru/kovardin/getapp/app/modules/admin/models"
	applications "ru/kovardin/getapp/app/modules/applications/models"
	"ru/kovardin/getapp/app/modules/billing/models"
	users "ru/kovardin/getapp/app/modules/users/models"
	"ru/kovardin/getapp/pkg/database"
	"ru/kovardin/getapp/pkg/logger"
)

type Payments struct {
	log          *logger.Logger
	payments     *database.Repository[models.Payment]
	products     *database.Repository[models.Product]
	applications *database.Repository[applications.Application]
	users        *database.Repository[users.User]
	settings     *database.Repository[settings.Setting]
	templates    map[string]*template.Template
}

func NewPayments(
	log *logger.Logger,
	payments *database.Repository[models.Payment],
	products *database.Repository[models.Product],
	users *database.Repository[users.User],
	applications *database.Repository[applications.Application],
	setting *database.Repository[settings.Setting],
) *Payments {
	return &Payments{
		log:          log,
		payments:     payments,
		products:     products,
		users:        users,
		applications: applications,
		settings:     setting,
		templates: map[string]*template.Template{
			"purchase": template.Must(template.ParseFiles(
				"templates/billing/base.gohtml",
				"templates/billing/purchase.gohtml",
			)),
			"success": template.Must(template.ParseFiles(
				"templates/billing/base.gohtml",
				"templates/billing/success.gohtml",
			)),
		},
	}
}

type Payment struct {
	Id      uint   `json:"id"`
	Amount  int    `json:"amount"`
	Product uint   `json:"product"`
	Name    string `json:"name"`
	Title   string `json:"title"`
	Status  string `json:"status"`
}

type RestoreResponse struct {
	Items []Payment `json:"items"`
}

func (p *Payments) Restore(w http.ResponseWriter, r *http.Request) {
	appid, err := strconv.Atoi(chi.URLParam(r, "application"))
	if err != nil {
		p.log.Error("error on get application", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userId, ok := r.Context().Value(auth.UserIdKey).(uint)
	if !ok {
		p.log.Error("error on get user id")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err := p.users.First(database.Condition{
		In: map[string]any{"id": userId},
	})
	if err != nil {
		p.log.Error("error on get user from db", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if user.ID == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	application, err := p.applications.First(database.Condition{
		In: map[string]any{"id": appid},
	})
	if err != nil {
		p.log.Error("error on get application from db", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if application.ID == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	payments, err := p.payments.Find(database.Condition{
		In: map[string]any{
			"user_id":        user.ID,
			"application_id": application.ID,
		},
		Preload: []string{
			"Product",
		},
	})
	if err != nil {
		p.log.Error("error on search payments", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp := RestoreResponse{
		Items: []Payment{},
	}

	for _, item := range payments {
		resp.Items = append(resp.Items, Payment{
			Id:      item.ID,
			Amount:  item.Amount,
			Product: item.ProductID,
			Name:    item.Product.Name,
			Title:   item.Product.Title,
			Status:  item.Status,
		})
	}

	render.JSON(w, r, resp)
}

func (p *Payments) Purchase(w http.ResponseWriter, r *http.Request) {
	appid, err := strconv.Atoi(chi.URLParam(r, "application"))
	if err != nil {
		p.log.Error("error on get application", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userId, ok := r.Context().Value(auth.UserIdKey).(uint)
	if !ok {
		p.log.Error("error on get user id")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err := p.users.First(database.Condition{
		In: map[string]any{"id": userId},
	})
	if err != nil {
		p.log.Error("error on get user from db", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if user.ID == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	application, err := p.applications.First(database.Condition{
		In: map[string]any{"id": appid},
	})
	if err != nil {
		p.log.Error("error on get application from db", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if application.ID == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	productId, err := strconv.Atoi(r.URL.Query().Get("product"))
	if err != nil {
		p.log.Error("error on get product id", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// find product in db
	product, err := p.products.First(database.Condition{
		In: map[string]any{
			"id":             productId,
			"application_id": application.ID,
		},
	})
	if err != nil {
		p.log.Error("error on get product from db", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if product.ID == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	wallet, err := p.settings.First(database.Condition{
		In: map[string]any{
			"key": "yoomoney_wallet",
		},
	})
	if err != nil {
		p.log.Error("error on get yoomoney wallet", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// save payment
	payment := models.Payment{
		ProductID:     product.ID,
		UserID:        user.ID,
		ApplicationID: application.ID,
		Amount:        product.Amount,
		Status:        models.PaymentStatusCreated,
	}

	if err := p.payments.Create(&payment); err != nil {
		p.log.Error("error on create payment", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	payload, err := json.Marshal(Payload{
		Payment: payment.ID,
	})
	if err != nil {
		p.log.Error("error on marshal payload", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	label := base64.StdEncoding.EncodeToString(payload)

	if err := p.templates["purchase"].ExecuteTemplate(w, "base", struct {
		Title   string
		Label   string
		Payment uint
		Status  string
		Product uint
		Amount  string
		Wallet  string
	}{
		Title:   "Purchase",
		Label:   label,
		Payment: payment.ID,
		Status:  payment.Status,
		Product: product.ID,
		// https://yourbasic.org/golang/round-float-2-decimal-places/
		Amount: fmt.Sprintf("%.2f", float64(product.Amount)/100),
		Wallet: wallet.Value,
	}); err != nil {
		p.log.Error("pay error", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (p *Payments) Success(w http.ResponseWriter, r *http.Request) {
	paymentId, err := strconv.Atoi(r.URL.Query().Get("payment"))
	if err != nil {
		p.log.Error("error on get payment id", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	payment, err := p.payments.First(database.Condition{
		In: map[string]any{
			"id": paymentId,
		},
	})
	if err != nil {
		p.log.Error("error on get payment from db", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if payment.ID == 0 {
		p.log.Error("error on get payment from db", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := p.payments.Switch(&payment, "status", models.PaymentStatusCreated, models.PaymentStatusSuccess); err != nil {
		p.log.Error("cant update status", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if p.templates["success"].ExecuteTemplate(w, "base", struct {
		Title string
	}{
		Title: "Success",
	}); err != nil {
		p.log.Error("pay error", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
	}
}

type PaymentResponse struct {
	Id      uint   `json:"id"`
	Amount  int    `json:"amount"`
	Product uint   `json:"product"`
	Name    string `json:"name"`
	Title   string `json:"title"`
	Status  string `json:"status"`
}

func (p *Payments) Payment(w http.ResponseWriter, r *http.Request) {
	paymentId, err := strconv.Atoi(chi.URLParam(r, "payment"))
	if err != nil {
		p.log.Error("error on get payment id", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	payment, err := p.payments.First(database.Condition{
		In: map[string]any{
			"id": paymentId,
		},
		Preload: []string{
			"Product",
		},
	})
	if err != nil {
		p.log.Error("error on get payment from db", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if payment.ID == 0 {
		p.log.Error("error on get payment from db", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	render.JSON(w, r, PaymentResponse{
		Id:      payment.ID,
		Product: payment.ProductID,
		Status:  payment.Status,
		Name:    payment.Product.Name,
		Title:   payment.Product.Title,
		Amount:  payment.Amount,
	})
}
