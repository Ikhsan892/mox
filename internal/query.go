package core

import "context"

type ErrQuery struct {
	Message string
	Type    string
	Meta    []interface{}
}

// Query interface
type IQuery[Result any] interface {
	Validate() error
	Handle(ctx context.Context) (Result, error)
}

type withQuery[Result any] struct {
	handler         IQuery[Result]
	err             error
	res             Result
	beforeQueryFunc func() error
	afterQueryFunc  func() error
}

// abstraction function
func WithQuery[Result any](handler IQuery[Result]) *withQuery[Result] {
	return &withQuery[Result]{
		handler: handler,
	}
}

func (withQuery *withQuery[Result]) BeforeQueryFunc(handler func() error) *withQuery[Result] {
	if err := handler(); err != nil {
		withQuery.err = err
	}

	withQuery.beforeQueryFunc = handler

	return withQuery
}

func (withQuery *withQuery[Result]) AfterQueryFunc(handler func() error) *withQuery[Result] {
	if err := handler(); err != nil {
		withQuery.err = err
	}

	withQuery.afterQueryFunc = handler

	return withQuery
}

func (withQuery *withQuery[Result]) IsError() bool {
	return withQuery.err != nil
}

func (withQuery *withQuery[Result]) Error() error {
	return withQuery.err
}

func (withQuery *withQuery[Result]) GetResult() Result {

	return withQuery.res
}

func (withQuery *withQuery[Result]) Validate() *withQuery[Result] {
	if err := withQuery.handler.Validate(); err != nil {
		withQuery.err = err
	}

	return withQuery
}

func (withQuery *withQuery[Result]) Execute(ctx context.Context) *withQuery[Result] {
	if withQuery.err != nil {
		return withQuery
	}

	if withQuery.beforeQueryFunc != nil {
		withQuery.beforeQueryFunc()
	}

	result, err := withQuery.handler.Handle(ctx)
	withQuery.err = err
	withQuery.res = result

	if withQuery.afterQueryFunc != nil {
		withQuery.afterQueryFunc()
	}

	return withQuery
}
