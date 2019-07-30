package rxgo

import (
	"context"

	"github.com/pkg/errors"
)

type functor struct {
	value interface{}
	error error
}

var emptyFunctor = functor{}

func from(value interface{}) functor {
	switch v := value.(type) {
	default:
		return functor{
			value: v,
		}
	case error:
		return functor{
			error: v,
		}
	}
}

// Iterator allows to iterates on an given structure
type Iterator interface {
	Next(ctx context.Context) (functor, error)
}

type iteratorFromChannel struct {
	ch         chan interface{}
	ctx        context.Context
	cancelFunc context.CancelFunc
}

type iteratorFromRange struct {
	current int
	end     int // Included
}

type iteratorFromSlice struct {
	index int
	s     []interface{}
}

func (f *functor) isValue() bool {
	return f.error == nil
}

func (f *functor) isError() bool {
	return f.error != nil
}

func (it *iteratorFromChannel) Next(ctx context.Context) (functor, error) {
	select {
	case <-ctx.Done():
		return emptyFunctor, &CancelledSubscriptionError{}
	case <-it.ctx.Done():
		return emptyFunctor, &CancelledIteratorError{}
	case next, ok := <-it.ch:
		if ok {
			return from(next), nil
		}
		return emptyFunctor, &NoSuchElementError{}
	}
}

func (it *iteratorFromRange) Next(ctx context.Context) (functor, error) {
	it.current++
	if it.current <= it.end {
		return from(it.current), nil
	}
	return emptyFunctor, errors.Wrap(&NoSuchElementError{}, "range does not contain anymore elements")
}

func (it *iteratorFromSlice) Next(ctx context.Context) (functor, error) {
	it.index++
	if it.index < len(it.s) {
		return from(it.s[it.index]), nil
	}
	return emptyFunctor, errors.Wrap(&NoSuchElementError{}, "slice does not contain anymore elements")
}

func newIteratorFromChannel(ch chan interface{}) Iterator {
	ctx, cancel := context.WithCancel(context.Background())
	return &iteratorFromChannel{
		ch:         ch,
		ctx:        ctx,
		cancelFunc: cancel,
	}
}

func newIteratorFromRange(start, end int) Iterator {
	return &iteratorFromRange{
		current: start,
		end:     end,
	}
}

func newIteratorFromSlice(s []interface{}) Iterator {
	return &iteratorFromSlice{
		index: -1,
		s:     s,
	}
}
