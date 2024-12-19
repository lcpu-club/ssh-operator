package main

func PoolExecute[I any, T any](fn func(I) (T, error), in []I, threads int) ([]T, error) {
	if threads < 1 {
		threads = 1
	}
	results := make(chan T, threads)
	errors := make(chan error, threads)
	waitCh := make(chan struct{}, threads)
	for _, item := range in {
		go func() {
			waitCh <- struct{}{}
			r, err := fn(item)
			results <- r
			errors <- err
		}()
	}

	var ret []T
	for range in {
		<-waitCh
		if err := <-errors; err != nil {
			return nil, err
		}
		ret = append(ret, <-results)
	}
	return ret, nil
}
