package entity


type ReqUserLogin struct {
	// google,apple
	Method   string `json:"method"`
	IdToken  string `json:"id_token"`
}



type UserAsset struct {
	TokenID  string `json:"token_id"`
	Balance  string `json:"balance"`
	Locked   string `json:"locked"`
	Free     string `json:"free"`
}



