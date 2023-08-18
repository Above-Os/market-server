package models

type UpdateReqData struct {
	AppName    string `json:"name"`
	CurVersion string `json:"curVersion"`
}

type UpdateReq struct {
	Updates []UpdateReqData `json:"updates"`
}

type UpdateResData struct {
	AppName       string `json:"name"`
	CurVersion    string `json:"curVersion"`
	LatestVersion string `json:"latestVersion"`
	NeedUpdate    bool   `json:"needUpdate"`
}

type UpdateRes struct {
	Updates []UpdateResData `json:"updates"`
}
