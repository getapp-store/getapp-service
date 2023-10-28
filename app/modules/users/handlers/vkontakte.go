package handlers

import (
	"encoding/json"
	"go.uber.org/zap"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"ru/kovardin/getapp/app/modules/users/models"
	"ru/kovardin/getapp/pkg/database"
	"ru/kovardin/getapp/pkg/logger"
)

type Vkontakte struct {
	log       *logger.Logger
	users     *database.Repository[models.User]
	templates map[string]*template.Template

	appId       string
	accessToken string
}

func NewVkontakte(log *logger.Logger, users *database.Repository[models.User]) *Vkontakte {
	return &Vkontakte{
		log:         log,
		users:       users,
		accessToken: "58a0982158a0982158a09821aa5bb4545f558a058a098213c0255c57a26b681d5d93b4c",
		templates: map[string]*template.Template{
			"login": template.Must(template.ParseFiles(
				"templates/users/vkontakte/login.gohtml",
			)),
		},
	}
}

type Payload struct {
	Type  string `json:"type"`
	Auth  int    `json:"auth"`
	Token string `json:"token"`
	Uuid  string `json:"uuid"`
	Hash  string `json:"hash"`
	Ttl   int    `json:"ttl"`
	User  struct {
		Id        int    `json:"id"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Avatar    string `json:"avatar"`
	}
}

type Data struct {
	Response struct {
		AccessToken       string `json:"access_token"`
		AccessTokenID     string `json:"access_token_id"`
		UserID            int    `json:"user_id"`
		Phone             string `json:"phone"`
		PhoneValidated    int    `json:"phone_validated"`
		IsPartial         bool   `json:"is_partial"`
		IsService         bool   `json:"is_service"`
		Email             string `json:"email"`
		Source            int    `json:"source"`
		SourceDescription string `json:"source_description"`
	} `json:"response"`
	Error struct {
		ErrorCode     int    `json:"error_code"`
		ErrorMsg      string `json:"error_msg"`
		RequestParams []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"request_params"`
	} `json:"error"`
}

func (u *Vkontakte) Auth(w http.ResponseWriter, r *http.Request) {
	productId := r.URL.Query().Get("state")
	raw := r.URL.Query().Get("payload")

	payload := Payload{}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		u.log.Error("error on unmarshal auth payload", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	u.log.Info("payload", zap.Any("payload", payload))

	// check is user real
	resp, err := http.Post(
		"https://api.vk.com/method/auth.exchangeSilentAuthToken",
		"application/x-www-form-urlencoded",
		strings.NewReader("v=5.131&token="+payload.Token+"&access_token="+u.accessToken+"&uuid="+payload.Uuid),
	)

	if err != nil {
		u.log.Error("error on check user token", zap.Error(err))
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	//dump, _ := httputil.DumpResponse(resp, true)

	//u.log.Info("response code", zap.Int("code", resp.StatusCode), zap.String("dump", string(dump)))

	data := Data{}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		u.log.Error("error on check unmarshal user", zap.Error(err))
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if data.Error.ErrorCode != 0 {
		u.log.Error("error on check user data", zap.Error(err))
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	u.log.Info("user", zap.Any("user", data))

	if data.Response.UserID != payload.User.Id {
		u.log.Error("error on check user id", zap.Error(err))
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// create internal user
	user, err := u.users.First(database.Condition{
		In: map[string]any{
			"external_id": data.Response.UserID,
		},
	})
	if err != nil {
		u.log.Error("error on get user by external id", zap.Error(err), zap.Int("external", data.Response.UserID))
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if user.ID == 0 {
		user.VkAccessToken = data.Response.AccessToken
		user.ExternalId = data.Response.UserID

		if err := u.users.Create(&user); err != nil {
			u.log.Error("error on create user", zap.Error(err))
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	}

	http.Redirect(w, r, "/v1/payments/purchase?user="+strconv.Itoa(int(user.ID))+"&product="+productId, http.StatusTemporaryRedirect)
}

func (u *Vkontakte) Login(w http.ResponseWriter, r *http.Request) {
	product := r.URL.Query().Get("product")

	err := u.templates["login"].ExecuteTemplate(w, "login", struct {
		Title   string
		Product string
	}{
		Title:   "Success",
		Product: product,
	})
	if err != nil {
		u.log.Error("auth error", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
	}
}
