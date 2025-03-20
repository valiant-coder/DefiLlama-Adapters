package entity_admin

type ReqLogin struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RespLogin struct {
	Key          string `json:"key"`
	IsFirstLogin bool   `json:"is_first_login"`
}

type ReqAuth struct {
	GoogleVerifyCode string `json:"google_verify_code"`
	Key              string `json:"key" binding:"required"`
}

type RespAuth struct {
	Token     string       `json:"token"`
	ExpiresAt int64        `json:"expires_at"`
	Roles     []*AdminRole `json:"roles"`
}

type ReqResetPassword struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password" binding:"required,password"`
	Name        string `json:"name" binding:"required,lte=30"`
}

type RespGoogleAuth struct {
	Secret string `json:"secret"`
	QRData string `json:"qr_data"`
}
