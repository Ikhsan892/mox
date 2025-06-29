package exceptions

type BaseException struct {
	Code    int
	Status  Status
	Message string
}

func NewBaseException(code int, message string) *BaseException {
	return &BaseException{
		Code:    code,
		Message: message,
	}
}

func (b *BaseException) WithStatus(status Status) *BaseException {
	b.Status = status

	return b
}

func (b BaseException) Error() string {
	return b.Message
}
