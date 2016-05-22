package store_test

import (
	"policy-server/store"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Tagger", func() {
	var (
		tagger store.Tagger
	)

	BeforeEach(func() {
		var err error
		tagger, err = store.NewMemoryTagger(4)
		Expect(err).NotTo(HaveOccurred())
	})

	It("is consistent: the same input always yields the same output", func() {
		tag1, err := tagger.GetTag("input1")
		Expect(err).NotTo(HaveOccurred())

		tag1again, err := tagger.GetTag("input1")
		Expect(err).NotTo(HaveOccurred())
		Expect(tag1).To(Equal(tag1again))
	})

	It("is injective: distinct inputs yield distinct outputs", func() {
		tag1, err := tagger.GetTag("input1")
		Expect(err).NotTo(HaveOccurred())

		tag2, err := tagger.GetTag("input2")
		Expect(err).NotTo(HaveOccurred())

		Expect(tag1).NotTo(Equal(tag2))
	})
})
