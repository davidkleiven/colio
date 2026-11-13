// Generated code by coliosoa. DO NOT EDIT

package parquet

import "github.com/davidkleiven/colio/pkg"

type PersonSoa struct {
	Name        []string  `parquet:"name"`
	Age         []int     `parquet:"age"`
	NumChildren []int32   `parquet:"numChildren"`
	Height      []float64 `parquet:"height"`
	Weight      []float32 `parquet:"weight"`
}

func (v *PersonSoa) Allocate(n int) {
	v.Name = make([]string, n)
	v.Age = make([]int, n)
	v.NumChildren = make([]int32, n)
	v.Height = make([]float64, n)
	v.Weight = make([]float32, n)
}

func (v *PersonSoa) Columns() []pkg.Column {
	return []pkg.Column{
		{Name: "Name", Tags: `parquet:"name"`, Ptr: &v.Name},
		{Name: "Age", Tags: `parquet:"age"`, Ptr: &v.Age},
		{Name: "NumChildren", Tags: `parquet:"numChildren"`, Ptr: &v.NumChildren},
		{Name: "Height", Tags: `parquet:"height"`, Ptr: &v.Height},
		{Name: "Weight", Tags: `parquet:"weight"`, Ptr: &v.Weight},
	}
}

func (v *PersonSoa) Cap() int {
	return min(
		len(v.Name),
		len(v.Age),
		len(v.NumChildren),
		len(v.Height),
		len(v.Weight),
	)
}

func (v *PersonSoa) ToAos() []Person {
	result := make([]Person, 0, v.Cap())
	for i := range v.Cap() {
		item := Person{
			Name:        v.Name[i],
			Age:         v.Age[i],
			NumChildren: v.NumChildren[i],
			Height:      v.Height[i],
			Weight:      v.Weight[i],
		}
		result = append(result, item)
	}
	return result
}

func (v *PersonSoa) Delete(i int) {
	v.Name[i] = v.Name[len(v.Name)-1]
	v.Name = v.Name[:len(v.Name)-1]
	v.Age[i] = v.Age[len(v.Age)-1]
	v.Age = v.Age[:len(v.Age)-1]
	v.NumChildren[i] = v.NumChildren[len(v.NumChildren)-1]
	v.NumChildren = v.NumChildren[:len(v.NumChildren)-1]
	v.Height[i] = v.Height[len(v.Height)-1]
	v.Height = v.Height[:len(v.Height)-1]
	v.Weight[i] = v.Weight[len(v.Weight)-1]
	v.Weight = v.Weight[:len(v.Weight)-1]
}
