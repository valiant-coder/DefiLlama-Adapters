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

func (s *UserInvitationService) DeleteInvitationLink(ctx context.Context, linkCode string) error {
	// 获取邀请链接
	userInviteLink, err := s.repo.GetUserInviteLink(ctx, linkCode)
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
