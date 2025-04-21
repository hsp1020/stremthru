package seedr

import (
	"net/url"
)

type ListContentsParams struct {
	Ctx
	ContentId string
}

type ListContentsDataFolder struct {
	Id         int    `json:"id"`
	Name       string `json:"name"`
	Fullname   string `json:"fullname"`
	Size       int64  `json:"size"`
	PlayAudio  bool   `json:"play_audio"`
	PlayVideo  bool   `json:"play_video"`
	IsShared   bool   `json:"is_shared"`
	LastUpdate string `json:"last_update"`
}

type ListContentsData struct {
	ResponseContainer
	SpaceMax       int64                    `json:"space_max"`
	SpaceUsed      int64                    `json:"space_used"`
	SawWalkthrough int                      `json:"saw_walkthrough"` // 1
	T              []float64                `json:"t"`               // YYYY-MM-DD hh:mm:ss
	Timestamp      string                   `json:"timestamp"`
	FolderId       int                      `json:"folder_id"`
	Fullname       string                   `json:"fullname"`
	Type           string                   `json:"type"` // folder
	Name           string                   `json:"name"`
	Parent         int                      `json:"parent"` // -1
	Indexes        []any                    `json:"indexes"`
	Torrents       []any                    `json:"torrents"`
	Folders        []ListContentsDataFolder `json:"folders"`
	Files          []any                    `json:"files"`
}

func (c APIClient) ListContents(params *ListContentsParams) (APIResponse[ListContentsData], error) {
	response := ListContentsData{}

	err := c.withAccessToken(&params.Ctx)
	if err != nil {
		return newAPIResponse(nil, response), err
	}

	if params.ContentId == "" {
		params.ContentId = "0"
	}

	params.Form = &url.Values{}
	params.Form.Add("func", "list_contents")
	params.Form.Add("content_type", "folder")
	params.Form.Add("content_id", params.ContentId)

	res, err := c.Request("POST", "/api/resource", params, &response)
	return newAPIResponse(res, response), err
}

type AddFolderParams struct {
	Ctx
	Name string
}

type AddFolderData struct {
	ResponseContainer
	Code   int  `json:"code"`
	Result bool `json:"result"`
}

func (c APIClient) AddFolder(params *AddFolderParams) (APIResponse[AddFolderData], error) {
	response := AddFolderData{}

	err := c.withAccessToken(&params.Ctx)
	if err != nil {
		return newAPIResponse(nil, response), err
	}

	params.Form = &url.Values{}
	params.Form.Add("func", "add_folder")
	params.Form.Add("name", params.Name)

	res, err := c.Request("POST", "/api/resource", params, &response)
	return newAPIResponse(res, response), err
}
