package data

type UserInvitationListParam struct {
	ListParam

	UID        string `json:"uid" form:"uid"`
	Inviter    string `json:"inviter" form:"inviter"`
	InviteCode string `json:"invite_code" form:"invite_code"`
}

type UILinkListParam struct {
	ListParam

	UID  string `json:"uid" form:"uid"`
	Code string `json:"code" form:"code"`
}

type UILinkParam struct {
	Percent uint `json:"percent" form:"percent"`
}
