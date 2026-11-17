// Generated code by coliosoa. DO NOT EDIT

package parquet

import (
	"github.com/davidkleiven/colio/colio"
	"iter"
	"time"
)

type PersonSoa struct {
	Name        []string    `parquet:"name"`
	Age         []int       `parquet:"age"`
	NumChildren []int32     `parquet:"numChildren"`
	Height      []float64   `parquet:"height"`
	Weight      []float32   `parquet:"weight"`
	CreatedAt   []time.Time `parquet:"createdAt"`
}

func (v *PersonSoa) Allocate(n int) {
	v.Name = make([]string, n)
	v.Age = make([]int, n)
	v.NumChildren = make([]int32, n)
	v.Height = make([]float64, n)
	v.Weight = make([]float32, n)
	v.CreatedAt = make([]time.Time, n)
}

func (v *PersonSoa) Columns() []colio.Column {
	return []colio.Column{
		{Name: "Name", Tags: `parquet:"name"`, Ptr: &v.Name},
		{Name: "Age", Tags: `parquet:"age"`, Ptr: &v.Age},
		{Name: "NumChildren", Tags: `parquet:"numChildren"`, Ptr: &v.NumChildren},
		{Name: "Height", Tags: `parquet:"height"`, Ptr: &v.Height},
		{Name: "Weight", Tags: `parquet:"weight"`, Ptr: &v.Weight},
		{Name: "CreatedAt", Tags: `parquet:"createdAt"`, Ptr: &v.CreatedAt},
	}
}

func (v *PersonSoa) Cap() int {
	return min(
		len(v.Name),
		len(v.Age),
		len(v.NumChildren),
		len(v.Height),
		len(v.Weight),
		len(v.CreatedAt),
	)
}

func (v *PersonSoa) MaxCap() int {
	return max(
		len(v.Name),
		len(v.Age),
		len(v.NumChildren),
		len(v.Height),
		len(v.Weight),
		len(v.CreatedAt),
	)
}

func (v *PersonSoa) FillEmpty() *PersonSoa {
	n := v.MaxCap()
	if len(v.Name) == 0 {
		v.Name = make([]string, n)
	}
	if len(v.Age) == 0 {
		v.Age = make([]int, n)
	}
	if len(v.NumChildren) == 0 {
		v.NumChildren = make([]int32, n)
	}
	if len(v.Height) == 0 {
		v.Height = make([]float64, n)
	}
	if len(v.Weight) == 0 {
		v.Weight = make([]float32, n)
	}
	if len(v.CreatedAt) == 0 {
		v.CreatedAt = make([]time.Time, n)
	}
	return v
}

func (v *PersonSoa) ToAos() iter.Seq[Person] {
	return func(yield func(Person) bool) {
		for i := range v.Cap() {
			item := Person{
				Name:        v.Name[i],
				Age:         v.Age[i],
				NumChildren: v.NumChildren[i],
				Height:      v.Height[i],
				Weight:      v.Weight[i],
				CreatedAt:   v.CreatedAt[i],
			}
			if !yield(item) {
				return
			}
		}
	}
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
	v.CreatedAt[i] = v.CreatedAt[len(v.CreatedAt)-1]
	v.CreatedAt = v.CreatedAt[:len(v.CreatedAt)-1]
}

func (v *PersonSoa) FromAos(aos []Person) *PersonSoa {
	v.Allocate(len(aos))
	for i, item := range aos {
		v.Name[i] = item.Name
		v.Age[i] = item.Age
		v.NumChildren[i] = item.NumChildren
		v.Height[i] = item.Height
		v.Weight[i] = item.Weight
		v.CreatedAt[i] = item.CreatedAt
	}
	return v
}

func (v *PersonSoa) At(i int) Person {
	return Person{
		Name:        v.Name[i],
		Age:         v.Age[i],
		NumChildren: v.NumChildren[i],
		Height:      v.Height[i],
		Weight:      v.Weight[i],
		CreatedAt:   v.CreatedAt[i],
	}
}

type EventSoa struct {
	Name  []string      `parquet:"name"`
	Tags  [][]int       `parquet:"tags,list"`
	Steps [][]float64   `parquet:"steps"`
	Times [][]time.Time `parquet:"times,list"`
}

func (v *EventSoa) Allocate(n int) {
	v.Name = make([]string, n)
	v.Tags = make([][]int, n)
	v.Steps = make([][]float64, n)
	v.Times = make([][]time.Time, n)
}

func (v *EventSoa) Columns() []colio.Column {
	return []colio.Column{
		{Name: "Name", Tags: `parquet:"name"`, Ptr: &v.Name},
		{Name: "Tags", Tags: `parquet:"tags,list"`, Ptr: &v.Tags},
		{Name: "Steps", Tags: `parquet:"steps"`, Ptr: &v.Steps},
		{Name: "Times", Tags: `parquet:"times,list"`, Ptr: &v.Times},
	}
}

func (v *EventSoa) Cap() int {
	return min(
		len(v.Name),
		len(v.Tags),
		len(v.Steps),
		len(v.Times),
	)
}

func (v *EventSoa) MaxCap() int {
	return max(
		len(v.Name),
		len(v.Tags),
		len(v.Steps),
		len(v.Times),
	)
}

func (v *EventSoa) FillEmpty() *EventSoa {
	n := v.MaxCap()
	if len(v.Name) == 0 {
		v.Name = make([]string, n)
	}
	if len(v.Tags) == 0 {
		v.Tags = make([][]int, n)
	}
	if len(v.Steps) == 0 {
		v.Steps = make([][]float64, n)
	}
	if len(v.Times) == 0 {
		v.Times = make([][]time.Time, n)
	}
	return v
}

func (v *EventSoa) ToAos() iter.Seq[Event] {
	return func(yield func(Event) bool) {
		for i := range v.Cap() {
			item := Event{
				Name:  v.Name[i],
				Tags:  v.Tags[i],
				Steps: v.Steps[i],
				Times: v.Times[i],
			}
			if !yield(item) {
				return
			}
		}
	}
}

func (v *EventSoa) Delete(i int) {
	v.Name[i] = v.Name[len(v.Name)-1]
	v.Name = v.Name[:len(v.Name)-1]
	v.Tags[i] = v.Tags[len(v.Tags)-1]
	v.Tags = v.Tags[:len(v.Tags)-1]
	v.Steps[i] = v.Steps[len(v.Steps)-1]
	v.Steps = v.Steps[:len(v.Steps)-1]
	v.Times[i] = v.Times[len(v.Times)-1]
	v.Times = v.Times[:len(v.Times)-1]
}

func (v *EventSoa) FromAos(aos []Event) *EventSoa {
	v.Allocate(len(aos))
	for i, item := range aos {
		v.Name[i] = item.Name
		v.Tags[i] = item.Tags
		v.Steps[i] = item.Steps
		v.Times[i] = item.Times
	}
	return v
}

func (v *EventSoa) At(i int) Event {
	return Event{
		Name:  v.Name[i],
		Tags:  v.Tags[i],
		Steps: v.Steps[i],
		Times: v.Times[i],
	}
}

type EmbeddedParentSoa struct {
	Id   []string
	Name []string
}

func (v *EmbeddedParentSoa) Allocate(n int) {
	v.Id = make([]string, n)
	v.Name = make([]string, n)
}

func (v *EmbeddedParentSoa) Columns() []colio.Column {
	return []colio.Column{
		{Name: "Id", Tags: "", Ptr: &v.Id},
		{Name: "Name", Tags: "", Ptr: &v.Name},
	}
}

func (v *EmbeddedParentSoa) Cap() int {
	return min(
		len(v.Id),
		len(v.Name),
	)
}

func (v *EmbeddedParentSoa) MaxCap() int {
	return max(
		len(v.Id),
		len(v.Name),
	)
}

func (v *EmbeddedParentSoa) FillEmpty() *EmbeddedParentSoa {
	n := v.MaxCap()
	if len(v.Id) == 0 {
		v.Id = make([]string, n)
	}
	if len(v.Name) == 0 {
		v.Name = make([]string, n)
	}
	return v
}

func (v *EmbeddedParentSoa) ToAos() iter.Seq[EmbeddedParent] {
	return func(yield func(EmbeddedParent) bool) {
		for i := range v.Cap() {
			item := EmbeddedParent{
				Id:   v.Id[i],
				Name: v.Name[i],
			}
			if !yield(item) {
				return
			}
		}
	}
}

func (v *EmbeddedParentSoa) Delete(i int) {
	v.Id[i] = v.Id[len(v.Id)-1]
	v.Id = v.Id[:len(v.Id)-1]
	v.Name[i] = v.Name[len(v.Name)-1]
	v.Name = v.Name[:len(v.Name)-1]
}

func (v *EmbeddedParentSoa) FromAos(aos []EmbeddedParent) *EmbeddedParentSoa {
	v.Allocate(len(aos))
	for i, item := range aos {
		v.Id[i] = item.Id
		v.Name[i] = item.Name
	}
	return v
}

func (v *EmbeddedParentSoa) At(i int) EmbeddedParent {
	return EmbeddedParent{
		Id:   v.Id[i],
		Name: v.Name[i],
	}
}

type EmbeddedChildSoa struct {
	Name []string
}

func (v *EmbeddedChildSoa) Allocate(n int) {
	v.Name = make([]string, n)
}

func (v *EmbeddedChildSoa) Columns() []colio.Column {
	return []colio.Column{
		{Name: "Name", Tags: "", Ptr: &v.Name},
	}
}

func (v *EmbeddedChildSoa) Cap() int {
	return min(
		len(v.Name),
	)
}

func (v *EmbeddedChildSoa) MaxCap() int {
	return max(
		len(v.Name),
	)
}

func (v *EmbeddedChildSoa) FillEmpty() *EmbeddedChildSoa {
	n := v.MaxCap()
	if len(v.Name) == 0 {
		v.Name = make([]string, n)
	}
	return v
}

func (v *EmbeddedChildSoa) ToAos() iter.Seq[EmbeddedChild] {
	return func(yield func(EmbeddedChild) bool) {
		for i := range v.Cap() {
			item := EmbeddedChild{
				Name: v.Name[i],
			}
			if !yield(item) {
				return
			}
		}
	}
}

func (v *EmbeddedChildSoa) Delete(i int) {
	v.Name[i] = v.Name[len(v.Name)-1]
	v.Name = v.Name[:len(v.Name)-1]
}

func (v *EmbeddedChildSoa) FromAos(aos []EmbeddedChild) *EmbeddedChildSoa {
	v.Allocate(len(aos))
	for i, item := range aos {
		v.Name[i] = item.Name
	}
	return v
}

func (v *EmbeddedChildSoa) At(i int) EmbeddedChild {
	return EmbeddedChild{
		Name: v.Name[i],
	}
}
