package marketplace

import (
	"context"
	"errors"
	"exapp-go/config"
	"exapp-go/data"
	"exapp-go/internal/db/ckhdb"
	"exapp-go/internal/db/db"
	"exapp-go/pkg/log"
	"exapp-go/pkg/nsqutil"
)

type UserInvitationService struct {
	repo    *db.Repo
	ckhRepo *ckhdb.ClickHouseRepo
	nsqPub  *nsqutil.Publisher
}

func NewUserInvitationService() *UserInvitationService {
	nsqConf := config.Conf().Nsq
	return &UserInvitationService{
		repo:    db.New(),
		ckhRepo: ckhdb.New(),
		nsqPub:  nsqutil.NewPublisher(nsqConf.Nsqds),
	}
}

func (s *UserInvitationService) GetUserInvitation(ctx context.Context, uid string) (*db.UserInvitation, error) {
	userInvitation, err := s.repo.GetUserInvitation(ctx, uid)
	if err != nil {
		return nil, err
	}
	
	return userInvitation, nil
}

func (s *UserInvitationService) GetUserInviteUsers(ctx context.Context, uid string) (*db.UserInvitation, error) {
	userInvitation, err := s.repo.GetUserInvitation(ctx, uid)
	if err != nil {
		return nil, err
	}
	
	return userInvitation, nil
}

func (s *UserInvitationService) GetInviteUsers(ctx context.Context, params data.UserInvitationListParam) (*db.UserInvitationListRes, error) {
	
	userInvitations, err := s.repo.ListUserInvitation(params)
	if err != nil {
		
		return nil, err
	}
	
	return &userInvitations, nil
}

func (s *UserInvitationService) GetUserInvitationLinks(ctx context.Context, params data.UILinkListParam) (*db.UILinkListRes, error) {
	userInvitationLink, err := s.repo.ListUserInviteLink(ctx, params)
	if err != nil {
		return nil, err
	}
	
	return userInvitationLink, nil
}

func (s *UserInvitationService) CreateUILink(ctx context.Context, uid string, params *data.UILinkParam) error {
	userInvitation, err := s.repo.GetUserInvitation(ctx, uid)
	if err != nil {
		return err
	}
	
	invitationConf, err := s.repo.GetUserPointsConf(ctx)
	if err != nil {
		
		return err
	}
	
	// 检查百分比
	if params.Percent > userInvitation.MaxPercent {
		log.Logger().Errorf("invite percent is max, %d > %d", params.Percent, userInvitation.MaxPercent)
		return errors.New("invite percent is max")
	}
	
	// 检查邀请链接数量是否超过限制
	if userInvitation.LinkCount+1 > invitationConf.MaxInviteLinkCount {
		log.Logger().Errorf("invite link count is max, %d > %d", userInvitation.LinkCount+1, invitationConf.MaxInviteLinkCount)
		return errors.New("invite link count is max")
	}
	
	err = s.repo.CreateUserInviteLink(ctx, userInvitation, params)
	if err != nil {
		
		log.Logger().Error("create user invite link error ->", err)
		return err
	}
	
	return nil
}

func (s *UserInvitationService) DeleteInvitationLink(ctx context.Context, code string) error {
	// 获取邀请链接
	userInviteLink, err := s.repo.GetUserInviteLink(ctx, code)
	if err != nil {
		
		log.Logger().Error("get user invite link error ->", err)
		return err
	}
	
	err = s.repo.DeleteUserInviteLink(ctx, userInviteLink)
	if err != nil {
		
		log.Logger().Error("delete user invite link error ->", err)
		return err
	}
	return nil
}

func (s *UserInvitationService) GetInvitationLinkByCode(ctx context.Context, code string) (*db.UserInviteLink, error) {
	link, err := s.repo.GetUserInviteLink(ctx, code)
	if err != nil {
		log.Logger().Error(code, "get user invite link error ->", err)
		return nil, err
	}
	return link, nil
}

func (s *UserInvitationService) BindInvitationLink(ctx context.Context, uid, code string) error {
	link, err := s.repo.GetUserInviteLink(ctx, code)
	if err != nil {
		
		log.Logger().Error(code, "get user invite link error ->", err)
		return err
	}
	
	// 查看当前用户是否已经绑定邀请链接
	invitation, _ := s.repo.GetUserInvitation(ctx, uid)
	if invitation != nil {
		
		log.Logger().Error(uid, "already bind invitation link ->", err)
		return err
	}
	
	inviter, err := s.repo.GetUser(ctx, link.UID)
	if err != nil {
		
		log.Logger().Error(link.UID, "get user error ->", err)
		return err
	}
	
	// check owner
	if inviter.UID == uid {
		
		log.Logger().Error(uid, "cannot bind your own invitation link ->", err)
		return err
	}
	
	conf, err := s.repo.GetUserPointsConf(ctx)
	if err != nil {
		
		log.Logger().Error("get user points conf error ->", err)
		return err
	}
	
	// TODO: 判断用户状态
	
	err = s.repo.Transaction(ctx, func(repo *db.Repo) error {
		
		// 创建邀请链接
		invitation = &db.UserInvitation{
			UID:           uid,
			MaxPercent:    conf.InvitePercent,
			Inviter:       inviter.UID,
			InviteCode:    link.Code,
			InvitePercent: link.Percent,
		}
		
		e := repo.Insert(ctx, invitation)
		if e != nil {
			
			log.Logger().Error(uid, "create user invitation error ->", e)
			return e
		}
		
		// 更新邀请人邀请信息
		e = repo.UpdateUIInviteCount(ctx, inviter.UID)
		if e != nil {
			
			log.Logger().Error(inviter.UID, "update user invitation invite count error ->", e)
		}
		return e
	})
	
	return err
}
