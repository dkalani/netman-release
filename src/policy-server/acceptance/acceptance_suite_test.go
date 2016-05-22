package acceptance_test

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net"
	"policy-server/config"

	. "github.com/onsi/ginkgo"
	gconfig "github.com/onsi/ginkgo/config"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"testing"
)

const DEFAULT_TIMEOUT = "5s"

var serverBinPath string

func TestAcceptance(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Acceptance Suite")
}

type beforeSuiteData struct {
	ServerBinPath string
}

var _ = SynchronizedBeforeSuite(func() []byte {
	// only run on node 1
	serverBinPath, err := gexec.Build("policy-server")
	Expect(err).NotTo(HaveOccurred())

	bytesToMarshal, err := json.Marshal(beforeSuiteData{
		ServerBinPath: serverBinPath,
	})
	Expect(err).NotTo(HaveOccurred())

	return bytesToMarshal
}, func(marshaledBytes []byte) {
	// run on all nodes
	var data beforeSuiteData
	Expect(json.Unmarshal(marshaledBytes, &data)).To(Succeed())
	serverBinPath = data.ServerBinPath

	rand.Seed(gconfig.GinkgoConfig.RandomSeed + int64(GinkgoParallelNode()))
})

var _ = SynchronizedAfterSuite(func() {
	// run on all nodes
}, func() {
	// run only on node 1
	gexec.CleanupBuildArtifacts()
})

func VerifyTCPConnection(address string) error {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}

func WriteConfigFile(serverConfig *config.ServerConfig) string {
	configFile, err := ioutil.TempFile("", "test-config")
	Expect(err).NotTo(HaveOccurred())

	serverConfig.Marshal(configFile)
	Expect(configFile.Close()).To(Succeed())

	return configFile.Name()
}
