package seedr

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/MunifTanjim/stremthru/core"
	"github.com/MunifTanjim/stremthru/store"
)

type StoreClientConfig struct {
	HTTPClient *http.Client
	UserAgent  string
}

type StoreClient struct {
	Name   store.StoreName
	client *APIClient
}

func (s *StoreClient) getFolderIdByName(ctx Ctx, folderName string) (int, error) {
	params := &ListContentsParams{
		Ctx: ctx,
	}
	res, err := s.client.ListContents(params)
	if err != nil {
		return -1, err
	}

	for _, folder := range res.Data.Folders {
		if folder.Name == folderName {
			return folder.Id, nil
		}
	}

	return -1, nil
}

func (s *StoreClient) ensureFolder(ctx Ctx, name string) (int, error) {
	if folderId, err := s.getFolderIdByName(ctx, name); err != nil {
		return -1, err
	} else if folderId != -1 {
		return folderId, nil
	}

	af_params := &AddFolderParams{
		Ctx:  Ctx{Ctx: ctx.Ctx},
		Name: name,
	}
	_, err := s.client.AddFolder(af_params)
	if err != nil {
		return -1, err
	}

	if folderId, err := s.getFolderIdByName(ctx, name); err != nil {
		return -1, err
	} else if folderId != -1 {
		return folderId, nil
	}
	return -1, errors.New("failed to ensure folder")
}

func (s *StoreClient) AddMagnet(params *store.AddMagnetParams) (*store.AddMagnetData, error) {
	magnet, err := core.ParseMagnetLink(params.Magnet)
	if err != nil {
		return nil, err
	}
	folderId, err := s.ensureFolder(Ctx{Ctx: params.Ctx}, magnet.Hash)
	if err != nil {
		return nil, err
	}
	res, err := s.client.AddTorrent(&AddTorrentParams{
		Ctx:           Ctx{Ctx: params.Ctx},
		TorrentMagnet: magnet.Link,
		FolderId:      folderId,
	})
	log.Debug("AddMagnet", "res", res, "err", err)
	if err != nil {
		return nil, err
	}
	data := &store.AddMagnetData{}
	return data, nil
}

// CheckMagnet implements store.Store.
func (s *StoreClient) CheckMagnet(params *store.CheckMagnetParams) (*store.CheckMagnetData, error) {
	panic("unimplemented")
}

// GenerateLink implements store.Store.
func (s *StoreClient) GenerateLink(params *store.GenerateLinkParams) (*store.GenerateLinkData, error) {
	panic("unimplemented")
}

// GetMagnet implements store.Store.
func (s *StoreClient) GetMagnet(params *store.GetMagnetParams) (*store.GetMagnetData, error) {
	panic("unimplemented")
}

func (s *StoreClient) GetUser(params *store.GetUserParams) (*store.User, error) {
	res, err := s.client.GetAccountSettings(&GetAccountSettingsParams{
		Ctx: Ctx{Ctx: params.Ctx},
	})
	if err != nil {
		return nil, err
	}
	data := &store.User{
		Id:                 strconv.Itoa(res.Data.Account.UserId),
		Email:              res.Data.Account.Email,
		SubscriptionStatus: store.UserSubscriptionStatusTrial,
	}
	return data, nil
}

// ListMagnets implements store.Store.
func (s *StoreClient) ListMagnets(params *store.ListMagnetsParams) (*store.ListMagnetsData, error) {
	_, err := s.client.ListContents(&ListContentsParams{
		Ctx: Ctx{Ctx: params.Ctx},
	})
	if err != nil {
		return nil, err
	}
	data := &store.ListMagnetsData{
		Items: []store.ListMagnetsDataItem{},
	}
	return data, nil
}

// RemoveMagnet implements store.Store.
func (s *StoreClient) RemoveMagnet(params *store.RemoveMagnetParams) (*store.RemoveMagnetData, error) {
	panic("unimplemented")
}

func NewStoreClient(config *StoreClientConfig) *StoreClient {
	c := &StoreClient{}
	c.client = NewAPIClient(&APIClientConfig{
		HTTPClient: config.HTTPClient,
		UserAgent:  config.UserAgent,
	})
	c.Name = store.StoreNameOffcloud

	return c
}

func (s *StoreClient) GetName() store.StoreName {
	return s.Name
}
