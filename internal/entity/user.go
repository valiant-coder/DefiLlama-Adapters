package entity


type ReqUserLogin struct {
	// google,apple
	Method   string `json:"method"`
	IdToken  string `json:"id_token"`
}
