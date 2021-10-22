package protocol

type LoginToGameServerRequest struct {
	Name    string `json:"name"`
	Uid     int64  `json:"uid"`
	HeadUrl string `json:"headUrl"`
}

type LoginToGameServerResponse struct {
	Uid      int64  `json:"acId"`
	Nickname string `json:"nickname"`
	HeadUrl  string `json:"headURL"`
}
