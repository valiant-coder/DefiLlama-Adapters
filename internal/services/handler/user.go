package handler

func (s *Service) updateUserTokenBalance(account string) error {
	go s.publisher.PublishBalanceUpdate(account)

	return nil
}
