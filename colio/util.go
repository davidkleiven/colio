package colio

import "slices"

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

type Deleter interface {
	Delete(int)
}

// DeleteMany the method sorts indices and removes duplicates
func DeleteMany(d Deleter, indices []int) {
	slices.Sort(indices)
	indices = slices.Compact(indices)
	for i := len(indices); i > 0; i-- {
		d.Delete(indices[i-1])
	}
}
