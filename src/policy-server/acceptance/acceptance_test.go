package acceptance_test

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"policy-server/client"
	"policy-server/config"
	"policy-server/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"github.com/pivotal-golang/lager/lagertest"
)

var _ = Describe("Policy server", func() {
	var (
		session        *gexec.Session
		address        string
		configFilePath string
		outerClient    *client.OuterClient
		innerClient    *client.InnerClient

		logger *lagertest.TestLogger
	)

	BeforeEach(func() {
		address = fmt.Sprintf("127.0.0.1:%d", 4001+GinkgoParallelNode())

		logger = lagertest.NewTestLogger("test")
		configFilePath = WriteConfigFile(&config.ServerConfig{
			ListenAddress: address,
		})

		serverCmd := exec.Command(serverBinPath, "-configFile", configFilePath)
		var err error
		session, err = gexec.Start(serverCmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		outerClient = client.NewOuterClient("http://"+address, http.DefaultClient)
		innerClient = client.NewInnerClient("http://"+address, http.DefaultClient)
	})

	AfterEach(func() {
		session.Interrupt()
		Eventually(session, DEFAULT_TIMEOUT).Should(gexec.Exit(0))
		Expect(os.Remove(configFilePath)).To(Succeed())
	})

	var serverIsAvailable = func() error {
		return VerifyTCPConnection(address)
	}

	It("should boot and gracefully terminate", func() {
		Eventually(serverIsAvailable, DEFAULT_TIMEOUT).Should(Succeed())

		Consistently(session).ShouldNot(gexec.Exit())

		session.Interrupt()
		Eventually(session, DEFAULT_TIMEOUT).Should(gexec.Exit(0))
	})

	Describe("rule lifecycle", func() {
		It("should support list, add and delete on the set of rules", func() {
			Eventually(serverIsAvailable, DEFAULT_TIMEOUT).Should(Succeed())

			By("polling for whitelists from the inside")
			groupRules, err := innerClient.GetWhitelists([]string{"group1", "group2"})
			Expect(err).NotTo(HaveOccurred())
			Expect(groupRules).To(HaveLen(2))
			Expect(groupRules[0].Destination.ID).To(Equal("group1"))
			Expect(groupRules[0].Destination.Tag).To(BeNil())
			Expect(groupRules[0].AllowedSources).To(BeEmpty())
			Expect(groupRules[1].Destination.ID).To(Equal("group2"))
			Expect(groupRules[1].Destination.Tag).To(BeNil())
			Expect(groupRules[1].AllowedSources).To(BeEmpty())

			By("listing the rules from the outside")
			rules, err := outerClient.ListRules()
			Expect(err).NotTo(HaveOccurred())
			Expect(rules).To(BeEmpty())

			By("adding a new rule")
			Expect(outerClient.AddRule(models.Rule{
				Source:      "group1",
				Destination: "group2",
			})).To(Succeed())

			By("getting the packet tags for the two groups")
			groupRules, err = innerClient.GetWhitelists([]string{"group1", "group2", "group3"})
			Expect(err).NotTo(HaveOccurred())
			Expect(groupRules).To(HaveLen(3))
			Expect(groupRules[0].Destination.ID).To(Equal("group1"))
			Expect(groupRules[0].Destination.Tag).NotTo(BeNil())
			group1Tag := *groupRules[0].Destination.Tag
			Expect(groupRules[0].AllowedSources).To(BeEmpty())

			Expect(groupRules[1].Destination.ID).To(Equal("group2"))
			Expect(groupRules[1].Destination.Tag).NotTo(BeNil())
			group2Tag := *groupRules[1].Destination.Tag
			Expect(groupRules[1].AllowedSources).To(HaveLen(1))
			Expect(*groupRules[1].AllowedSources[0].Tag).To(Equal(group1Tag))

			Expect(groupRules[2].Destination.ID).To(Equal("group3"))
			Expect(groupRules[2].Destination.Tag).To(BeNil())
			Expect(groupRules[2].AllowedSources).To(BeEmpty())

			By("checking that the packet tags are unique")
			Expect(group1Tag).NotTo(Equal(group2Tag))

			By("adding a second rule")
			Expect(outerClient.AddRule(models.Rule{
				Source:      "group2",
				Destination: "group3",
			})).To(Succeed())

			By("getting the packet tags for the third group")
			groupRules, err = innerClient.GetWhitelists([]string{"group3"})
			Expect(err).NotTo(HaveOccurred())
			Expect(groupRules).To(HaveLen(1))
			Expect(groupRules[0].Destination.ID).To(Equal("group3"))
			Expect(groupRules[0].Destination.Tag).NotTo(BeNil())
			group3Tag := *groupRules[0].Destination.Tag
			Expect(groupRules[0].AllowedSources).To(HaveLen(1))
			Expect(groupRules[0].AllowedSources[0]).To(BeEquivalentTo(models.TaggedGroup{
				ID:  "group2",
				Tag: &group2Tag,
			}))

			Expect(group3Tag).NotTo(Equal(group2Tag))

			By("adding a third rule")
			Expect(outerClient.AddRule(models.Rule{
				Source:      "group2",
				Destination: "group2",
			})).To(Succeed())

			By("getting the packet tags for the second group")
			groupRules, err = innerClient.GetWhitelists([]string{"group2"})
			Expect(err).NotTo(HaveOccurred())
			Expect(groupRules).To(HaveLen(1))
			Expect(groupRules[0].Destination.ID).To(Equal("group2"))
			Expect(*groupRules[0].Destination.Tag).To(Equal(group2Tag))
			Expect(groupRules[0].AllowedSources).To(HaveLen(2))
			Expect(groupRules[0].AllowedSources).To(ContainElement(BeEquivalentTo(models.TaggedGroup{
				ID:  "group1",
				Tag: &group1Tag,
			})))
			Expect(groupRules[0].AllowedSources).To(ContainElement(BeEquivalentTo(models.TaggedGroup{
				ID:  "group2",
				Tag: &group2Tag,
			})))

			By("listing the rules from the outside")
			rules, err = outerClient.ListRules()
			Expect(err).NotTo(HaveOccurred())
			Expect(rules).To(HaveLen(3))
			Expect(rules).To(ConsistOf([]models.Rule{
				{Source: "group1", Destination: "group2"},
				{Source: "group2", Destination: "group3"},
				{Source: "group2", Destination: "group2"},
			}))

			By("removing the second rule")
			Expect(outerClient.DeleteRule(models.Rule{
				Source:      "group2",
				Destination: "group3",
			})).To(Succeed())

			By("listing the rules")
			rules, err = outerClient.ListRules()
			Expect(err).NotTo(HaveOccurred())
			Expect(rules).To(HaveLen(2))
			Expect(rules).To(ConsistOf([]models.Rule{
				{Source: "group1", Destination: "group2"},
				{Source: "group2", Destination: "group2"},
			}))

			By("getting the packet tags for the third group")
			groupRules, err = innerClient.GetWhitelists([]string{"group3"})
			Expect(err).NotTo(HaveOccurred())
			Expect(groupRules).To(HaveLen(1))
			Expect(groupRules[0].Destination.ID).To(Equal("group3"))
			Expect(*groupRules[0].Destination.Tag).To(Equal(group3Tag))
			Expect(groupRules[0].AllowedSources).To(BeEmpty())
		})
	})
})
