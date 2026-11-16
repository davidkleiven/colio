package colio

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type deleter struct {
	data []int
}

func (d *deleter) Delete(i int) {
	d.data[i] = d.data[len(d.data)-1]
	d.data = d.data[:len(d.data)-1]
}

func TestDeleteMany(t *testing.T) {

	data := deleter{
		data: []int{1, 2, 3, 4, 5},
	}

	DeleteMany(&data, []int{2, 4})
	require.Equal(t, []int{1, 2, 4}, data.data)

	DeleteMany(&data, []int{1, 1, 1})
	require.Equal(t, []int{1, 4}, data.data)

}
