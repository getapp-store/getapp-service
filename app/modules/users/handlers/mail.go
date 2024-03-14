package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"ru/kovardin/getapp/app/modules/users/models"
	"ru/kovardin/getapp/pkg/database"
	"ru/kovardin/getapp/pkg/logger"
	"ru/kovardin/getapp/pkg/mail"
	"ru/kovardin/getapp/pkg/utils"
)

const (
	textTemplate = `
Your authorization code

PIN: %s
`
	htmlTemplate = `
<p>Your authorization code</p>
<h2>PIN: %s</h2>
`
)

type Mail struct {
	log       *logger.Logger
	mailer    *mail.Mailer
	users     *database.Repository[models.User]
	pincodes  *database.Repository[models.Pincode]
	templates map[string]*template.Template
}

func NewMail(log *logger.Logger, mailer *mail.Mailer, users *database.Repository[models.User], pins *database.Repository[models.Pincode]) *Mail {
	return &Mail{
		log:      log,
		mailer:   mailer,
		users:    users,
		pincodes: pins,
		templates: map[string]*template.Template{
			"login": template.Must(template.ParseFiles(
				"templates/users/mail/login.gohtml",
			)),
			"send": template.Must(template.ParseFiles(
				"templates/users/mail/send.gohtml",
			)),
			"success": template.Must(template.ParseFiles(
				"templates/users/mail/success.gohtml",
			)),
		},
	}
}

func (m *Mail) Login(w http.ResponseWriter, r *http.Request) {
	application, err := strconv.Atoi(chi.URLParam(r, "application"))
	if err != nil {
		m.log.Error("error on get application", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := m.templates["login"].ExecuteTemplate(w, "login", struct {
		Title       string
		Application int
	}{
		Title:       "Login",
		Application: application,
	}); err != nil {
		m.log.Error("auth error", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (m *Mail) Send(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		m.log.Error("error on parse mail form", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")

	if email == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	application, err := strconv.Atoi(chi.URLParam(r, "application"))
	if err != nil {
		m.log.Error("error on get application", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err := m.users.First(database.Condition{
		In: map[string]any{
			"email":          email,
			"application_id": application,
		},
	})

	if err != nil {
		m.log.Error("error on search user", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if user.ID == 0 {
		// create new user
		user.Email = email
		user.ApplicationID = uint(application)
		if err := m.users.Save(&user); err != nil {
			m.log.Error("error on save new user", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	// generate code
	code := utils.GeneratePincodeString(6)

	pin := models.Pincode{
		UserID: user.ID,
		Code:   code,
	}

	// save code
	if err := m.pincodes.Save(&pin); err != nil {
		m.log.Error("error on saving pin", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// send code to mail
	if err := m.mailer.Send(mail.Message{
		From:    "robot@getapp.store",
		Name:    "getapp",
		To:      email,
		Subject: "Pincode",
		Text:    fmt.Sprintf(textTemplate, pin.Code),
		Html:    fmt.Sprintf(htmlTemplate, pin.Code),
	}); err != nil {
		m.log.Error("error on send mail", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		// todo error template
		return
	}

	if err := m.templates["send"].ExecuteTemplate(w, "send", struct {
		Title       string
		Application int
	}{
		Title:       "Pincode",
		Application: application,
	}); err != nil {
		m.log.Error("auth error", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (m *Mail) Auth(w http.ResponseWriter, r *http.Request) {
	application := chi.URLParam(r, "application")

	if err := r.ParseForm(); err != nil {
		m.log.Error("error on parse pincode form", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	code := r.FormValue("pincode")

	if code == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	pin, err := m.pincodes.First(database.Condition{
		In: map[string]any{
			"code": code,
		},
	})

	if err != nil {
		m.log.Error("error on search pincode", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if pin.ID == 0 {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	user, err := m.users.First(database.Condition{
		In: map[string]any{
			"id": pin.UserID,
		},
	})

	if err != nil {
		m.log.Error("error on search user", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	user.ApiToken = utils.GenerateTokenString(32)

	if err := m.users.Save(&user); err != nil {
		m.log.Error("error on save user", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// redirect to callback with token
	http.Redirect(w, r, "/v1/users/"+application+"/mail/success?token="+user.ApiToken, http.StatusFound)
}

func (m *Mail) Success(w http.ResponseWriter, r *http.Request) {
	if err := m.templates["success"].ExecuteTemplate(w, "success", struct {
		Title string
	}{
		Title: "Success",
	}); err != nil {
		m.log.Error("auth error", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
	}
}
