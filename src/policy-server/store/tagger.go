package store

import (
	"encoding/binary"
	"errors"
	"fmt"
	"policy-server/models"
	"sync"
)

type Tagger interface {
	GetTag(groupID string) (*models.PacketTag, error)
}

type memoryTagger struct {
	TagLength uint

	tags map[string]*models.PacketTag
	lock sync.Mutex
}

func NewMemoryTagger(tagLength int) (Tagger, error) {
	if tagLength < 1 || tagLength > 8 {
		return nil, errors.New("invalid tag length")
	}
	return &memoryTagger{
		TagLength: uint(tagLength),
		tags:      make(map[string]*models.PacketTag),
	}, nil
}

func (t *memoryTagger) intToPacketTag(x int) (*models.PacketTag, error) {
	buffer := make([]byte, 8)
	binary.LittleEndian.PutUint64(buffer, uint64(x))
	if x >= 1<<t.TagLength {
		return nil, fmt.Errorf("not enough bytes to represent %d", x)
	}
	pt := models.PacketTag(buffer[0:t.TagLength])
	return &pt, nil

}

func (t *memoryTagger) GetTag(groupID string) (*models.PacketTag, error) {
	t.lock.Lock()
	defer t.lock.Unlock()

	if tag, ok := t.tags[groupID]; ok {
		return tag, nil
	}
	newTag, err := t.intToPacketTag(len(t.tags) + 1)
	if err != nil {
		return nil, fmt.Errorf("form new packet tag: %s", err)
	}
	t.tags[groupID] = newTag
	return newTag, nil
}
