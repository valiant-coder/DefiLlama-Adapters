package db

import (
	"context"
	"exapp-go/config"
	"exapp-go/data"
	"exapp-go/internal/tools"
	
	"gorm.io/gorm"
)

func init() {
	
	addMigrateFunc(func(r *Repo) error {
		
		return r.AutoMigrate(&UserInviteLink{})
	})
}

type UserInviteLink struct {
	gorm.Model
	UID     string `gorm:"column:uid;type:varchar(255);not null;index:idx_uid;"`
	Code    string `gorm:"column:code;type:varchar(255);not null;;uniqueIndex:idx_code"`
	Link    string `gorm:"column:link;type:varchar(255);"`
	Count   uint   `gorm:"column:invite_count;type:int(11);not null;default:0"`   // 邀请人数
	Percent uint   `gorm:"column:invite_percent;type:int(11);not null;default:0"` // 返佣比例
}

func UILRedisKey(code string) string {
	
	return "user-invite-link::detail:code:" + code
}

func (uil *UserInviteLink) RedisKey() string {
	
	return UILRedisKey(uil.Code)
}

func InviteCodeRedisKey() string {
	
	return "user-invitation:invite_code"
}

func (r *Repo) GenerateInviteCode() string {
	
	code := tools.GenerateRandomString(8)
	
	// check redis hash
	if r.redis.HExists(context.TODO(), InviteCodeRedisKey(), code).Val() {
		return r.GenerateInviteCode()
	}
	
	r.redis.HSet(context.TODO(), InviteCodeRedisKey(), code, 1)
	return code
}

func (r *Repo) GetUserInviteLink(ctx context.Context, code string) (*UserInviteLink, error) {
	
	if res := GetCache[UserInviteLink](UILRedisKey(code)); res != nil {
		
		return res, nil
	}
	
	var uil UserInviteLink
	if err := r.WithContext(ctx).DB.Where("code = ?", code).First(&uil).Error; err != nil {
		
		return nil, err
	}
	
	r.SaveCache(&uil)
	return &uil, nil
}

type UILinkListRes ListResult[UserInviteLink]

func (r *Repo) ListUserInviteLink(ctx context.Context, params data.UILinkListParam) (*UILinkListRes, error) {
	
	if params.Order == "" {
		params.Order = "created_at desc"
	}
	
	res, err := List[data.UILinkListParam, UserInviteLink](params)
	if err != nil {
		return nil, err
	}
	
	result := UILinkListRes(res)
	return &result, nil
}

func (r *Repo) UpdateLinkInviteCount(ctx context.Context, link *UserInviteLink) (err error) {
	
	return r.WithContext(ctx).DB.Model(link).Update("invite_count", gorm.Expr("invite_count + 1")).Error
}

func (r *Repo) CreateUserInviteLink(ctx context.Context, invitation *UserInvitation, params *data.UILinkParam) (err error) {
	
	code := r.GenerateInviteCode()
	inviteLink := config.Conf().Invitation.Host + "/" + code
	
	link := &UserInviteLink{
		UID:     invitation.UID,
		Code:    code,
		Link:    inviteLink,
		Percent: params.Percent,
	}
	
	err = r.Transaction(ctx, func(r *Repo) error {
		
		// 创建邀请链接
		if err := r.Create(link).Error; err != nil {
			return err
		}
		// 更新Invitation 邀请数量
		if err := r.UpdateUILinkCount(ctx, invitation.UID, false); err != nil {
			return err
		}
		
		return nil
	})
	
	return
}

func (r *Repo) DeleteUserInviteLink(ctx context.Context, link *UserInviteLink) (err error) {
	
	err = r.Transaction(ctx, func(rep *Repo) error {
		
		if e := rep.Delete(ctx, link); e != nil {
			
			return e
		}
		
		if e := rep.UpdateUILinkCount(ctx, link.UID, true); e != nil {
			
			return e
		}
		
		return nil
	})
	
	return
}
