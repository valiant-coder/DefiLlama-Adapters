package errno

type ErrorPair struct {
	Code uint32
	Msg  string
}

var (
	EpUserNotFound = ErrorPair{Code: 10001, Msg: "user not found"}

	EpFaucetClaimed = ErrorPair{Code: 10101, Msg: "Address already claimed."}
	EpFaucetNotEnabled = ErrorPair{Code: 10102, Msg: "Faucet is not enabled."}
	EpFaucetNotRegistered = ErrorPair{Code: 10103, Msg: "Log in at test.1dex.com to get your deposit address."}
)
