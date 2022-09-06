package mmap

func ch() <-chan struct{} {
	ch := make(chan struct{})
	close(ch)

	return ch
}
