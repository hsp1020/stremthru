package seedr

import (
	"net/url"
	"strconv"
)

type AddTorrentParams struct {
	Ctx
	TorrentMagnet string
	FolderId      int
}

type AddTorrentData struct {
	ResponseContainer
}

func (c APIClient) AddTorrent(params *AddTorrentParams) (APIResponse[AddTorrentData], error) {
	response := AddTorrentData{}

	err := c.withAccessToken(&params.Ctx)
	if err != nil {
		return newAPIResponse(nil, response), err
	}

	params.Form = &url.Values{}
	params.Form.Add("func", "add_torrent")
	params.Form.Add("torrent_magnet", params.TorrentMagnet)
	params.Form.Add("folder_id", strconv.Itoa(params.FolderId))

	res, err := c.Request("POST", "/api/resource", params, &response)
	return newAPIResponse(res, response), err
}
