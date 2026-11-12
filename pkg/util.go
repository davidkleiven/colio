package pkg

func Must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func Assert(cond bool, msg string) {
	if !cond {
		panic(msg)
	}
}
