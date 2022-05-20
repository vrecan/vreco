package broadcast

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
)

type BroadCast struct {
	InputChan chan string
	Listeners map[uuid.UUID]Listener
	lock      sync.Locker
}

//NewBroadcast is a simple wrapper to allow you to broadcast to many channels
func NewBroadcast() *BroadCast {
	return &BroadCast{
		Listeners: make(map[uuid.UUID]Listener, 0),
		lock:      &sync.Mutex{},
	}
}

type Listener struct {
	ID   uuid.UUID
	Chan chan string
}

func (b *BroadCast) AddListener() Listener {
	b.lock.Lock()
	defer b.lock.Unlock()
	id, err := uuid.NewUUID()
	if err != nil {
		panic("Failed to get a uuid")
	}

	list := Listener{
		ID:   id,
		Chan: make(chan string),
	}
	b.Listeners[id] = list
	fmt.Println("listeners ID: ", list.ID)
	return list
}

func (b *BroadCast) RemoveListener(list Listener) {
	b.lock.Lock()
	defer b.lock.Unlock()
	delete(b.Listeners, list.ID)
}

func (b *BroadCast) Send(msg string) {
	for _, l := range b.Listeners {

		select {
		case l.Chan <- msg:
		default:
			fmt.Println("failed to send message to chan for listner: ", l.ID)
		}
	}
}
