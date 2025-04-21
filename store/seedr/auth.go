package seedr

import (
	"net/url"
	"time"

	"github.com/MunifTanjim/stremthru/internal/kv"
)

var seedrAuthState = kv.NewKVStore[AuthState](&kv.KVStoreConfig{
	Type: "seedr:auth",
})

type AuthState struct {
	AccessToken  string `json:"atok"`
	RefreshToken string `json:"rtok"`
	ExpiresAt    int64  `json:"exp"`
}

func (as *AuthState) load(username string) error {
	return seedrAuthState.Get(username, as)
}

func (as AuthState) save(username string) error {
	return seedrAuthState.Set(username, as)
}

func (as AuthState) IsAuthed() bool {
	return as.RefreshToken != ""
}

func (as AuthState) IsExpired() bool {
	return as.ExpiresAt <= time.Now().Unix()
}

func (as AuthState) IsExpiring() bool {
	return as.ExpiresAt <= time.Now().Add(10*time.Minute).Unix()
}

func (c APIClient) withAccessToken(ctx *Ctx) error {
	user := ctx.GetUser()

	if ctx.auth == nil {
		ctx.auth = &AuthState{}
	}

	err := ctx.auth.load(user.Username)
	if err != nil {
		return err
	}

	if ctx.auth.IsAuthed() {
		if ctx.auth.IsExpiring() {
			if err := c.refreshToken(ctx); err != nil {
				log.Error("failed to refresh access token", "error", err)
			} else {
				log.Info("refreshed access token")
			}
		}
		return nil
	}

	res, err := c.generateToken(ctx)
	if err != nil {
		return err
	}

	ctx.auth.AccessToken = res.Data.AccessToken
	ctx.auth.RefreshToken = res.Data.RefreshToken
	ctx.auth.ExpiresAt = time.Now().Unix() + res.Data.ExpiresIn

	err = ctx.auth.save(user.Username)
	if err != nil {
		log.Error("failed to store auth state", "error", err)
		return err
	}
	return nil
}

type generateTokenData struct {
	ResponseContainer
	AccessToken  string `json:"access_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"` // Bearer
	Scope        string `json:"scope"`
	RefreshToken string `json:"refresh_token"`
}

func (c APIClient) generateToken(ctx *Ctx) (APIResponse[generateTokenData], error) {
	user := ctx.GetUser()
	ctx.Form = &url.Values{}
	ctx.Form.Add("grant_type", "password")
	ctx.Form.Add("client_id", "seedr_chrome")
	ctx.Form.Add("type", "login")
	ctx.Form.Add("username", user.Username)
	ctx.Form.Add("password", user.Password)
	response := generateTokenData{}
	res, err := c.Request("POST", "/api/token", ctx, &response)
	return newAPIResponse(res, response), err
}

type refreshTokenParams struct {
	Ctx
}

type refreshTokenData = generateTokenData

func (c APIClient) refreshToken(ctx *Ctx) error {
	ctx.Form = &url.Values{}
	ctx.Form.Add("grant_type", "refresh_token")
	ctx.Form.Add("refresh_token", ctx.auth.RefreshToken)
	ctx.Form.Add("client_id", "seedr_chrome")
	response := refreshTokenData{}
	_, err := c.Request("POST", "/oauth_test/token.php", ctx, &response)
	if err != nil {
		return err
	}
	ctx.auth.AccessToken = response.AccessToken
	ctx.auth.RefreshToken = response.RefreshToken
	ctx.auth.ExpiresAt = time.Now().Unix() + response.ExpiresIn
	err = ctx.auth.save(ctx.GetUser().Username)
	return err
}

type GetAccountSettingsParams struct {
	Ctx
}

type GetAccountSettingsData struct {
	ResponseContainer
	Code     int  `json:"code"`
	Result   bool `json:"result"`
	Settings struct {
		AllowRemoteAccess  bool   `json:"allow_remote_access"`
		SiteLanguage       string `json:"site_language"`      // en
		SubtitlesLanguage  string `json:"subtitles_language"` // eng
		EmailAnnouncements bool   `json:"email_announcements"`
		EmailNewsletter    bool   `json:"email_newsletter"`
	} `json:"settings"`
	Account struct {
		Username        string `json:"username"`
		UserId          int    `json:"user_id"`
		Premium         int    `json:"premium"`      // 0
		PackageId       int    `json:"package_id"`   // -1
		PackageName     string `json:"package_name"` // NON-PREMIUM
		SpaceUsed       int    `json:"space_used"`
		SpaceMax        int    `json:"space_max"`
		BandwidthUsed   int    `json:"bandwidth_used"`
		Email           string `json:"email"`
		Wishlist        []any  `json:"wishlist"`
		Invites         int    `json:"invites"`
		InvitesAccepted int    `json:"invites_accepted"`
		MaxInvites      int    `json:"max_invites"`
	} `json:"account"`
	Country string `json:"country"`
}

func (c APIClient) GetAccountSettings(params *GetAccountSettingsParams) (APIResponse[GetAccountSettingsData], error) {
	response := GetAccountSettingsData{}
	err := c.withAccessToken(&params.Ctx)
	if err != nil {
		return newAPIResponse(nil, response), err
	}
	params.Form = &url.Values{}
	params.Form.Add("func", "get_settings")
	res, err := c.Request("GET", "/api/resource", params, &response)
	return newAPIResponse(res, response), err
}
