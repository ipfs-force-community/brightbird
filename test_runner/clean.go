package main

import (
	"github.com/hashicorp/go-multierror"
	"sync"
)

type Cleaner struct {
	fnLk sync.Mutex
	fns  []func() error
}

func (cleaner *Cleaner) AddFunc(fn func() error) {
	cleaner.fnLk.Lock()
	defer cleaner.fnLk.Unlock()

	cleaner.fns = append(cleaner.fns, fn)
}

func (cleaner *Cleaner) DoClean() error {
	cleaner.fnLk.Lock()
	defer cleaner.fnLk.Unlock()

	var mErr error
	for _, fn := range cleaner.fns {
		if err := fn(); err != nil {
			mErr = multierror.Append(mErr, err)
		}
	}
	return mErr
}
