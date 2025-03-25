package data

type UserInvitationListParam struct {
	ListParam

	UID        string `json:"uid" url:"uid"`
	Inviter    string `json:"inviter" url:"inviter"`
	InviteCode string `json:"invite_code" url:"invite_code"`
}

type UILinkListParam struct {
	ListParam

	UID  string `json:"uid" url:"uid"`
	Code string `json:"code" url:"code"`
}

type UILinkParam struct {
	Percent uint `json:"percent" url:"percent"`
}
