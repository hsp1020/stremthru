package seedr

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/MunifTanjim/stremthru/core"
	"github.com/MunifTanjim/stremthru/internal/config"
	"github.com/MunifTanjim/stremthru/internal/request"
	"github.com/MunifTanjim/stremthru/store"
)

var DefaultHTTPClient = config.DefaultHTTPClient

type APIClientConfig struct {
	BaseURL    string // default: https://www.seedr.cc
	APIKey     string
	HTTPClient *http.Client
	UserAgent  string
}

type APIClient struct {
	BaseURL    *url.URL
	HTTPClient *http.Client
	apiKey     string
	agent      string

	reqQuery  func(query *url.Values, params request.Context)
	reqHeader func(query *http.Header, params request.Context)

	// authCache cache.Cache[cachedAuth]
}

func NewAPIClient(conf *APIClientConfig) *APIClient {
	if conf.UserAgent == "" {
		conf.UserAgent = ""
	}

	if conf.BaseURL == "" {
		conf.BaseURL = "https://www.seedr.cc"
	}

	if conf.HTTPClient == nil {
		conf.HTTPClient = DefaultHTTPClient
	}

	c := &APIClient{}

	baseUrl, err := url.Parse(conf.BaseURL)
	if err != nil {
		panic(err)
	}

	c.BaseURL = baseUrl
	c.HTTPClient = conf.HTTPClient
	c.apiKey = conf.APIKey
	c.agent = conf.UserAgent

	c.reqQuery = func(query *url.Values, params request.Context) {
	}

	c.reqHeader = func(header *http.Header, params request.Context) {
		header.Add("User-Agent", c.agent)
	}

	// c.authCache = cache.NewCache[cachedAuth](&cache.CacheConfig{
	// 	Name:     "store:offcloud:cookie",
	// 	Lifetime: 6 * time.Hour,
	// })

	return c
}

type Ctx struct {
	request.Ctx
	auth *AuthState `json:"-"`
}

type Context interface {
	PrepareSeedrHeader(header *http.Header)
}

func (ctx Ctx) PrepareSeedrHeader(header *http.Header) {
	if ctx.auth == nil {
		return
	}
	if ctx.auth.AccessToken != "" {
		header.Add("Authorization", "Bearer "+ctx.auth.AccessToken)
		log.Debug("H", "h", header)
	}
}

type SeedrUser struct {
	Username string
	Password string
}

func (ctx Ctx) GetUser() *SeedrUser {
	if username, password, ok := strings.Cut(ctx.APIKey, ":"); ok {
		return &SeedrUser{
			Username: strings.ToLower(username),
			Password: password,
		}
	}
	return &SeedrUser{}
}

func (c APIClient) doRequest(params request.Context, req *http.Request, v ResponseEnvelop) (*http.Response, error) {
	if ctx, ok := params.(Context); ok {
		ctx.PrepareSeedrHeader(&req.Header)
	}

	res, err := c.HTTPClient.Do(req)
	err = processResponseBody(res, err, v)
	if err != nil {
		err := UpstreamErrorWithCause(err)
		err.InjectReq(req)
		if res != nil {
			err.StatusCode = res.StatusCode
		}
		return res, err
	}
	return res, nil

}

func (c APIClient) Request(method, path string, params request.Context, v ResponseEnvelop) (*http.Response, error) {
	if params == nil {
		params = &Ctx{}
	}
	req, err := params.NewRequest(c.BaseURL, method, path, c.reqHeader, c.reqQuery)
	if err != nil {
		error := core.NewStoreError("failed to create request")
		error.StoreName = string(store.StoreNameSeedr)
		error.Cause = err
		return nil, error
	}
	return c.doRequest(params, req, v)
}
