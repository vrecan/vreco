package broadcast

import (
	"fmt"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
)

func TestBroadcastInit(t *testing.T) {
	b := NewBroadcast()
	assert.NotNil(t, b)
}

func TestBroadcastAddElem(t *testing.T) {
	b := NewBroadcast()
	lid := b.AddListener()
	b.RemoveListener(lid)
	assert.Equal(t, 0, len(b.Listeners))
	spew.Println(b)
}

func TestBroadcastAddManyElem(t *testing.T) {
	size := 100
	b := NewBroadcast()
	lids := make([]Listener, size)
	for i := 0; i < size; i++ {
		list := b.AddListener()
		spew.Println(list)
		lids = append(lids, list)
	}
	for _, v := range lids {
		fmt.Println("ID:", v.ID)
		b.RemoveListener(v)
	}
	assert.Equal(t, 0, len(b.Listeners))
	spew.Println(b)
}
