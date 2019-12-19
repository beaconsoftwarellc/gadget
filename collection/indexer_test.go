package collection

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/beaconsoftwarellc/gadget/generator"
)

type TestIndexable struct {
	id string
	m  map[string]string
}

func (ti *TestIndexable) GetID() string {
	return ti.id
}

func (ti *TestIndexable) GetField(fieldName string) interface{} {
	value, ok := ti.m[fieldName]
	if !ok {
		value = ""
	}
	return value
}

func NewTestIndexable(id string, fname, fvalue string) *TestIndexable {
	return &TestIndexable{id: id, m: map[string]string{fname: fvalue}}
}

func Test_newIndex(t *testing.T) {
	assert := assert.New(t)
	expected := generator.String(20)
	indexFace := NewIndex(expected)
	index := indexFace.(*index)
	assert.Equal(index.name, expected)
	assert.NotNil(index.valuesToIDs)
}

func TestIndex_Add(t *testing.T) {
	assert := assert.New(t)
	id := generator.TestID()
	fieldName := generator.String(20)
	fieldValue := generator.String(20)
	indexable := NewTestIndexable(id, fieldName, fieldValue)

	indexFace := NewIndex(fieldName)
	index := indexFace.(*index)
	index.Add(indexable)

	set, ok := index.valuesToIDs[fieldValue]
	assert.True(ok)
	assert.True(set.Contains(id))

	id2 := generator.TestID()
	indexable2 := NewTestIndexable(id2, fieldName, fieldValue)
	index.Add(indexable2)

	set, ok = index.valuesToIDs[fieldValue]
	assert.True(ok)
	assert.True(set.Contains(id2))
}

func TestIndex_Update(t *testing.T) {
	assert := assert.New(t)
	id := generator.TestID()
	fieldName := generator.String(20)
	fieldValue := generator.String(20)
	indexable := NewTestIndexable(id, fieldName, fieldValue)

	indexFace := NewIndex(fieldName)
	index := indexFace.(*index)
	index.Add(indexable)

	fieldValue2 := generator.String(20)
	indexable.m[fieldName] = fieldValue2

	index.Update(indexable)
	set, ok := index.valuesToIDs[fieldValue]
	assert.True(ok)
	assert.False(set.Contains(id))
	set, ok = index.valuesToIDs[fieldValue2]
	assert.True(ok)
	assert.True(set.Contains(id))
}

func TestIndex_Remove(t *testing.T) {
	assert := assert.New(t)
	id := generator.TestID()
	fieldName := generator.String(20)
	fieldValue := generator.String(20)
	indexable := NewTestIndexable(id, fieldName, fieldValue)

	indexFace := NewIndex(fieldName)
	index := indexFace.(*index)
	index.Add(indexable)
	index.Remove(indexable)

	set, ok := index.valuesToIDs[fieldValue]
	assert.True(ok)
	assert.False(set.Contains(id))
}

func Test_stringify(t *testing.T) {
	type args struct {
		objs []interface{}
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			args: args{objs: []interface{}{}},
			want: []string{},
		},
		{
			args: args{objs: []interface{}{"a", "b"}},
			want: []string{"a", "b"},
		},
		{
			args: args{objs: []interface{}{"a", 1}},
			want: []string{"a", ""},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stringify(tt.args.objs); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stringify() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIndex_IdsForValue(t *testing.T) {
	assert := assert.New(t)
	fieldName := generator.String(20)
	fieldValue := generator.String(20)
	indexFace := NewIndex(fieldName)
	index := indexFace.(*index)
	expected := NewSet()
	actual := NewSet(anonymize(index.LookupValue(fieldValue)...)...)
	assert.Equal(expected, actual)

	for i := 0; i < 10; i++ {
		id := strconv.Itoa(i)
		indexable := NewTestIndexable(id, fieldName, fieldValue)
		index.Add(indexable)
		expected.Add(id)
	}
	actual = NewSet(anonymize(index.LookupValue(fieldValue)...)...)
	assert.Equal(expected, actual)
}

func TestIndexer_Index(t *testing.T) {
	assert := assert.New(t)
	indexer := NewIndexer()

	fieldName := generator.String(20)
	fieldValue := generator.String(20)
	indexer.Index(fieldName)

	fieldName1 := generator.String(20)
	fieldValue1 := generator.String(20)
	indexer.Index(fieldName1)

	indexable := NewTestIndexable(generator.TestID(), fieldName, fieldValue)
	indexable.m[fieldName1] = fieldValue1
	indexer.Add(indexable)

	actual, ok := indexer.Values(fieldName, fieldValue)
	if assert.True(ok) && assert.Equal(1, len(actual)) {
		assert.Equal(indexable, actual[0])
	}

	indexer.Remove(indexable)
	actual, ok = indexer.Values(fieldName, fieldValue)
	assert.True(ok)
	assert.Equal(0, len(actual))
}
