package util

import "sync"

func Parallelize(functions ...func() error) (errors []error) {
	var waitGroup sync.WaitGroup
	mux := &sync.Mutex{}
	waitGroup.Add(len(functions))

	defer waitGroup.Wait()

	for _, function := range functions {
		// We can't do this with a transaction, but it should be safe with a standard read
		//go func(copy func() error) {
		func(copy func() error) {
			defer waitGroup.Done()
			err := copy()
			if err != nil {
				mux.Lock()
				errors = append(errors, err)
				mux.Unlock()
			}
		}(function)
	}
	return
}
