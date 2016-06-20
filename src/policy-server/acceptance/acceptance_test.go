package acceptance_test

import (
	"fmt"
	"io/ioutil"
	"lib/testsupport"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"policy-server/config"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Acceptance", func() {
	var (
		session       *gexec.Session
		conf          config.Config
		address       string
		mockUAAServer *httptest.Server
		testDatabase  *testsupport.TestDatabase
	)

	var serverIsAvailable = func() error {
		return VerifyTCPConnection(address)
	}

	BeforeEach(func() {
		mockUAAServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/check_token" {
				if r.Header["Authorization"][0] == "Basic dGVzdDp0ZXN0" {
					bodyBytes, _ := ioutil.ReadAll(r.Body)
					token := strings.Split(string(bodyBytes), "=")[1]
					Expect(token).NotTo(BeEmpty())

					if string(token) == "valid-token" {
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(`{"scope":["network.admin"], "user_name":"some-user"}`))
					} else {
						w.WriteHeader(http.StatusBadRequest)
						w.Write([]byte(`{"error_description":"Some requested scopes are missing: network.admin"}`))
					}
				} else {
					w.WriteHeader(http.StatusUnauthorized)
				}
				return
			}
			w.WriteHeader(http.StatusNotFound)
		}))

		dbName := fmt.Sprintf("test_netman_database_%x", rand.Int())
		dbConnectionInfo := testsupport.GetDBConnectionInfo()
		testDatabase = dbConnectionInfo.CreateDatabase(dbName)

		conf = config.Config{
			ListenHost:      "127.0.0.1",
			ListenPort:      9001 + GinkgoParallelNode(),
			UAAClient:       "test",
			UAAClientSecret: "test",
			UAAURL:          mockUAAServer.URL,
			Database:        testDatabase.DBConfig(),
		}
		configFilePath := WriteConfigFile(conf)

		policyServerCmd := exec.Command(policyServerPath, "-config-file", configFilePath)
		var err error
		session, err = gexec.Start(policyServerCmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		address = fmt.Sprintf("%s:%d", conf.ListenHost, conf.ListenPort)

		Eventually(serverIsAvailable, DEFAULT_TIMEOUT).Should(Succeed())
	})

	AfterEach(func() {
		session.Interrupt()
		Eventually(session, DEFAULT_TIMEOUT).Should(gexec.Exit())

		if testDatabase != nil {
			testDatabase.Destroy()
		}
	})

	It("should boot and gracefully terminate", func() {
		Consistently(session).ShouldNot(gexec.Exit())

		session.Interrupt()
		Eventually(session, DEFAULT_TIMEOUT).Should(gexec.Exit())
	})

	It("has a whoami endpoint", func() {
		client := &http.Client{}
		tokenString := "valid-token"
		req, err := http.NewRequest("GET", fmt.Sprintf("http://%s:%d/networking/v0/external/whoami", conf.ListenHost, conf.ListenPort), nil)
		Expect(err).NotTo(HaveOccurred())
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenString))

		resp, err := client.Do(req)
		Expect(err).NotTo(HaveOccurred())

		Expect(resp.StatusCode).To(Equal(http.StatusOK))
		responseString, err := ioutil.ReadAll(resp.Body)
		Expect(responseString).To(ContainSubstring("some-user"))
	})

	Describe("adding policies", func() {
		Context("when the request is missing an Authorization header", func() {
			It("responds with 401", func() {
				client := &http.Client{}
				body := strings.NewReader(`{ "policies": [ {"source": { "id": "some-app-guid" }, "destination": { "id": "some-other-app-guid", "protocol": "tcp", "port": 8090 } } ] }`)
				req, err := http.NewRequest("POST", fmt.Sprintf("http://%s:%d/networking/v0/external/policies", conf.ListenHost, conf.ListenPort), body)
				Expect(err).NotTo(HaveOccurred())

				resp, err := client.Do(req)
				Expect(err).NotTo(HaveOccurred())

				Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
				responseString, err := ioutil.ReadAll(resp.Body)
				Expect(responseString).To(MatchJSON(`{ "error": "missing authorization header"}`))
			})
		})

		Context("when the authorization token is invalid", func() {
			It("responds with 403", func() {
				client := &http.Client{}
				body := strings.NewReader(`{ "policies": [ {"source": { "id": "some-app-guid" }, "destination": { "id": "some-other-app-guid", "protocol": "tcp", "port": 8090 } } ] }`)
				req, err := http.NewRequest("POST", fmt.Sprintf("http://%s:%d/networking/v0/external/policies", conf.ListenHost, conf.ListenPort), body)
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set("Authorization", "Bearer bad-token")

				resp, err := client.Do(req)
				Expect(err).NotTo(HaveOccurred())

				Expect(resp.StatusCode).To(Equal(http.StatusForbidden))
				responseString, err := ioutil.ReadAll(resp.Body)
				Expect(responseString).To(MatchJSON(`{ "error": "failed to verify token with uaa" }`))
			})
		})

		Context("when the user is authorized", func() {
			It("responds with 200 and a body of {} and we can see it in the list", func() {
				client := &http.Client{}
				body := strings.NewReader(`{ "policies": [ {"source": { "id": "some-app-guid" }, "destination": { "id": "some-other-app-guid", "protocol": "tcp", "port": 8090 } } ] }`)
				req, err := http.NewRequest("POST", fmt.Sprintf("http://%s:%d/networking/v0/external/policies", conf.ListenHost, conf.ListenPort), body)
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set("Authorization", "Bearer valid-token")

				resp, err := client.Do(req)
				Expect(err).NotTo(HaveOccurred())

				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				responseString, err := ioutil.ReadAll(resp.Body)
				Expect(responseString).To(MatchJSON("{}"))

				req, err = http.NewRequest("GET", fmt.Sprintf("http://%s:%d/networking/v0/external/policies", conf.ListenHost, conf.ListenPort), nil)
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set("Authorization", "Bearer valid-token")
				resp, err = client.Do(req)
				Expect(err).NotTo(HaveOccurred())

				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				responseString, err = ioutil.ReadAll(resp.Body)
				Expect(responseString).To(MatchJSON(`{ "policies": [ {"source": { "id": "some-app-guid" }, "destination": { "id": "some-other-app-guid", "protocol": "tcp", "port": 8090 } } ] }`))
			})
		})

	})
	Describe("listing policies", func() {
		Context("when the request is missing an Authorization header", func() {
			It("responds with 401", func() {
				client := &http.Client{}
				req, err := http.NewRequest("GET", fmt.Sprintf("http://%s:%d/networking/v0/external/policies", conf.ListenHost, conf.ListenPort), nil)
				Expect(err).NotTo(HaveOccurred())

				resp, err := client.Do(req)
				Expect(err).NotTo(HaveOccurred())

				Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
				responseString, err := ioutil.ReadAll(resp.Body)
				Expect(responseString).To(MatchJSON(`{ "error": "missing authorization header"}`))
			})
		})

		Context("when the authorization token is invalid", func() {
			It("responds with 403", func() {
				client := &http.Client{}
				req, err := http.NewRequest("GET", fmt.Sprintf("http://%s:%d/networking/v0/external/policies", conf.ListenHost, conf.ListenPort), nil)
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set("Authorization", "Bearer bad-token")

				resp, err := client.Do(req)
				Expect(err).NotTo(HaveOccurred())

				Expect(resp.StatusCode).To(Equal(http.StatusForbidden))
				responseString, err := ioutil.ReadAll(resp.Body)
				Expect(responseString).To(MatchJSON(`{ "error": "failed to verify token with uaa" }`))
			})
		})
	})
})
