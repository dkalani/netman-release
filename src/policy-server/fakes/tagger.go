package fakes

import "policy-server/models"

type Tagger struct {
	GetTagStub func(groupID string) (*models.PacketTag, error)
}

func (t *Tagger) GetTag(groupID string) (*models.PacketTag, error) {
	return t.GetTagStub(groupID)
}
