package store_test

import (
	"policy-server/fakes"
	"policy-server/models"
	"policy-server/store"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager/lagertest"
)

var _ = Describe("Store", func() {
	var (
		memStore     *store.MemoryStore
		logger       *lagertest.TestLogger
		tagger       *fakes.Tagger
		tagCallCount int
	)
	BeforeEach(func() {
		tagger = &fakes.Tagger{}
		tagger.GetTagStub = func(groupID string) (*models.PacketTag, error) {
			defer func() { tagCallCount++ }()
			return models.PT(groupID + "-tag"), nil
		}
		memStore = store.NewMemoryStore(tagger)
		logger = lagertest.NewTestLogger("test")
	})

	Describe("tagging", func() {
		It("gets a tag for each rule", func() {
			Expect(memStore.Add(logger, models.Rule{
				Source:      "group0",
				Destination: "group1",
			})).To(Succeed())
			Expect(tagCallCount).To(Equal(2))
			Expect(memStore.Add(logger, models.Rule{
				Source:      "group0",
				Destination: "group1",
			})).To(Succeed())
			Expect(tagCallCount).To(Equal(4))
		})
	})

	Describe("GetWhitelists", func() {
		var whitelists []models.IngressWhitelist

		BeforeEach(func() {
			Expect(memStore.Add(logger, models.Rule{
				Source:      "group0",
				Destination: "group1",
			})).To(Succeed())

			var err error
			whitelists, err = memStore.GetWhitelists(logger, []string{"group0", "group1"})
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when the list of groups is nil", func() {
			It("returns all whitelist rules, sorted by group id", func() {
				var err error
				whitelists, err = memStore.GetWhitelists(logger, nil)
				Expect(err).NotTo(HaveOccurred())

				Expect(whitelists).To(HaveLen(2))

				Expect(whitelists[0]).To(Equal(models.IngressWhitelist{
					Destination:    models.TaggedGroup{ID: "group0", Tag: models.PT("group0-tag")},
					AllowedSources: nil,
				}))
				Expect(whitelists[1]).To(Equal(models.IngressWhitelist{
					Destination:    models.TaggedGroup{ID: "group1", Tag: models.PT("group1-tag")},
					AllowedSources: []models.TaggedGroup{{ID: "group0", Tag: models.PT("group0-tag")}},
				}))
			})
		})

		It("returns the whitelist for each requested group", func() {
			Expect(whitelists).To(HaveLen(2))

			Expect(whitelists[0].Destination.ID).To(Equal("group0"))
			Expect(whitelists[1].Destination.ID).To(Equal("group1"))
		})

		It("returns the packet tag for each group", func() {
			Expect(*whitelists[0].Destination.Tag).To(BeEquivalentTo([]byte("group0-tag")))
			Expect(*whitelists[1].Destination.Tag).To(BeEquivalentTo([]byte("group1-tag")))
		})

		It("returns the AllowedSourceTags for each destination", func() {
			Expect(whitelists[0].AllowedSources).To(HaveLen(0))
			Expect(whitelists[1].AllowedSources).To(HaveLen(1))
			Expect(whitelists[1].AllowedSources).To(ContainElement(BeEquivalentTo(models.TaggedGroup{
				ID:  "group0",
				Tag: models.PT("group0-tag"),
			})))
		})

		Context("when a rule is deleted", func() {
			It("no longer reports stale tags", func() {
				Expect(memStore.Delete(logger, models.Rule{
					Source:      "group0",
					Destination: "group1",
				})).To(Succeed())

				Expect(whitelists).NotTo(BeEmpty())

				var err error
				whitelists, err = memStore.GetWhitelists(logger, nil)
				Expect(err).NotTo(HaveOccurred())

				Expect(whitelists).To(BeEmpty())
			})
		})

		Context("when the group is unknown", func() {
			BeforeEach(func() {
				var err error
				whitelists, err = memStore.GetWhitelists(logger, []string{"group1", "some-other-group", "group0"})
				Expect(err).NotTo(HaveOccurred())
			})

			It("should compute a tag and not error", func() {
				Expect(whitelists[0].Destination.ID).To(Equal("group1"))
				Expect(*whitelists[0].Destination.Tag).To(BeEquivalentTo([]byte("group1-tag")))

				Expect(whitelists[1].Destination.ID).To(Equal("some-other-group"))
				Expect(whitelists[1].Destination.Tag).To(Equal(models.PT("some-other-group-tag")))

				Expect(whitelists[2].Destination.ID).To(Equal("group0"))
				Expect(*whitelists[2].Destination.Tag).To(BeEquivalentTo([]byte("group0-tag")))
			})
		})
	})
})
