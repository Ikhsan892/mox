package core

type Queue struct {
	handler []QueueHandler
}

type QueueHandler interface {
}

func (q *Queue) Dispatch(handler QueueHandler) error {
	return nil
}
