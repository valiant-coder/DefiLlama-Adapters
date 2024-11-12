package errno

type ErrorPair struct {
	Code uint32
	Msg  string
}

var (
	EpUserNotFound = ErrorPair{Code: 10001, Msg: "user not found"}
)
