package broadcast

import (
	"testing"

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
}

func TestBroadcastAddManyElem(t *testing.T) {
	size := 100
	b := NewBroadcast()
	lids := make([]Listener, size)
	for i := 0; i < size; i++ {
		list := b.AddListener()
		lids = append(lids, list)
	}
	for _, v := range lids {
		b.RemoveListener(v)
	}
	assert.Equal(t, 0, len(b.Listeners))

}

func TestBroadcastSendToMany(t *testing.T) {
	size := 5
	b := NewBroadcast()
	lids := make([]Listener, size)
	for i := 0; i < size; i++ {
		list := b.AddListener()
		lids = append(lids, list)
	}
	for _, v := range lids {

		b.RemoveListener(v)
	}
	assert.Equal(t, 0, len(b.Listeners))

	testMsg := "this is a test message for all"

	go func() {
		for _, listener := range lids {
			msg := <-listener.Chan
			assert.Equal(t, testMsg, msg)

		}
	}()
	errors := b.Send(testMsg)
	assert.Empty(t, errors)

	errors = b.Send(testMsg)
	assert.Empty(t, errors)
}
