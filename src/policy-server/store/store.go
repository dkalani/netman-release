package store

import (
	"errors"
	"fmt"
	"policy-server/models"
	"sort"
	"sync"

	"github.com/pivotal-golang/lager"
)

type MemoryStore struct {
	Tagger Tagger
	tags   map[string]*models.PacketTag
	rules  []models.Rule
	lock   sync.Mutex
}

func NewMemoryStore(tagger Tagger) *MemoryStore {
	return &MemoryStore{
		Tagger: tagger,
		tags:   make(map[string]*models.PacketTag),
	}
}

func (s *MemoryStore) GetWhitelists(logger lager.Logger, groups []string) ([]models.IngressWhitelist, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if groups == nil {
		for key := range s.tags {
			groups = append(groups, key)
		}
		sort.Strings(groups)
	}

	all := make([]models.IngressWhitelist, len(groups))

	for i, destGroup := range groups {
		all[i].Destination.ID = destGroup
		var found bool
		all[i].Destination.Tag, found = s.tags[destGroup]
		if !found {
			logger.Info("no-tag-found", lager.Data{"destination": destGroup})
			continue
		}
		for _, rule := range s.rules {
			if rule.Destination != destGroup {
				continue
			}
			all[i].AllowedSources = append(all[i].AllowedSources, models.TaggedGroup{
				ID:  rule.Source,
				Tag: s.tags[rule.Source],
			})
		}
	}
	logger.Info("built-whitelist", lager.Data{"whitelist": all})
	return all, nil
}

func (s *MemoryStore) Add(logger lager.Logger, rule models.Rule) error {
	logger = logger.Session("memory-store-add")
	logger.Info("start")
	defer logger.Info("done")

	sourceTag, err := s.Tagger.GetTag(rule.Source)
	if err != nil {
		logger.Error("get-tag", err, lager.Data{"source": rule.Source})
		return fmt.Errorf("get tag: %s", err)
	}

	destinationTag, err := s.Tagger.GetTag(rule.Destination)
	if err != nil {
		logger.Error("get-tag", err, lager.Data{"destination": rule.Destination})
		return fmt.Errorf("get tag: %s", err)
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	s.rules = append(s.rules, rule)
	s.tags[rule.Source] = sourceTag
	s.tags[rule.Destination] = destinationTag
	logger.Info("added", lager.Data{"rule": rule, "source-tag": sourceTag, "destination-tag": destinationTag})

	return nil
}

func (s *MemoryStore) Delete(logger lager.Logger, rule models.Rule) error {
	logger = logger.Session("memory-store-delete")
	logger.Info("start")
	defer logger.Info("done")

	s.lock.Lock()
	defer s.lock.Unlock()

	newRules := []models.Rule{}

	for _, r := range s.rules {
		if !rule.Equals(r) {
			newRules = append(newRules, r)
		}
	}

	if len(newRules) == len(s.rules) {
		return errors.New("not found")
	}

	s.rules = newRules

	logger.Info("deleted", lager.Data{"rule": rule})
	return nil
}

func (s *MemoryStore) List(logger lager.Logger) ([]models.Rule, error) {
	logger = logger.Session("memory-store-list")
	logger.Info("start")
	defer logger.Info("done")

	s.lock.Lock()
	defer s.lock.Unlock()

	toReturn := make([]models.Rule, len(s.rules))
	copy(toReturn, s.rules)

	return toReturn, nil
}
