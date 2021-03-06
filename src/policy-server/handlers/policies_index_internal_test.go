package handlers_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"policy-server/fakes"
	"policy-server/handlers"
	"policy-server/models"

	lfakes "lib/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/pivotal-golang/lager/lagertest"
)

var _ = Describe("PoliciesIndexInternal", func() {
	var (
		// allPolicies []models.Policy
		handler   *handlers.PoliciesIndexInternal
		resp      *httptest.ResponseRecorder
		fakeStore *fakes.Store
		logger    *lagertest.TestLogger
		marshaler *lfakes.Marshaler
	)

	BeforeEach(func() {
		allPolicies := []models.Policy{{
			Source: models.Source{ID: "some-app-guid"},
			Destination: models.Destination{
				ID:       "some-other-app-guid",
				Protocol: "tcp",
				Port:     8080,
			},
		}, {
			Source: models.Source{ID: "another-app-guid"},
			Destination: models.Destination{
				ID:       "some-other-app-guid",
				Protocol: "udp",
				Port:     1234,
			},
		}}

		marshaler = &lfakes.Marshaler{}
		marshaler.MarshalStub = json.Marshal
		fakeStore = &fakes.Store{}
		fakeStore.AllReturns(allPolicies, nil)
		logger = lagertest.NewTestLogger("test")
		handler = &handlers.PoliciesIndexInternal{
			Logger:    logger,
			Store:     fakeStore,
			Marshaler: marshaler,
		}
		resp = httptest.NewRecorder()
	})

	Context("when there are no policies", func() {
		It("returns an empty set", func() {
			fakeStore.AllReturns([]models.Policy{}, nil)
			request, err := http.NewRequest("GET", "/networking/v0/internal/policies", nil)
			Expect(err).NotTo(HaveOccurred())
			handler.ServeHTTP(resp, request)

			Expect(resp.Body).To(MatchJSON(`{ "policies": [] }`))
		})
	})

	Context("when there are policies and no filter is passed", func() {
		It("it returns all of them", func() {
			expectedResponseJSON := `{"policies": [
				{
					"source": {
						"id": "some-app-guid"
					},
					"destination": {
						"id": "some-other-app-guid",
						"protocol": "tcp",
						"port": 8080
					}
				},
				{
					"source": {
						"id": "another-app-guid"
					},
					"destination": {
						"id": "some-other-app-guid",
						"protocol": "udp",
						"port": 1234
					}
				}
			]}`
			request, err := http.NewRequest("GET", "/networking/v0/internal/policies", nil)
			Expect(err).NotTo(HaveOccurred())
			handler.ServeHTTP(resp, request)

			Expect(fakeStore.AllCallCount()).To(Equal(1))
			Expect(resp.Code).To(Equal(http.StatusOK))
			Expect(resp.Body).To(MatchJSON(expectedResponseJSON))

		})
	})

	Context("when there are policies and a filter is passed", func() {
		It("it returns only policies that match the filter", func() {
			expectedResponseJSON := `{"policies": [
				{
					"source": {
						"id": "some-app-guid"
					},
					"destination": {
						"id": "some-other-app-guid",
						"protocol": "tcp",
						"port": 8080
					}
				}
			]}`
			request, err := http.NewRequest("GET", "/networking/v0/internal/policies?id=some-app-guid", nil)
			Expect(err).NotTo(HaveOccurred())
			handler.ServeHTTP(resp, request)

			Expect(fakeStore.AllCallCount()).To(Equal(1))
			Expect(resp.Code).To(Equal(http.StatusOK))
			Expect(resp.Body).To(MatchJSON(expectedResponseJSON))
		})
	})

	Context("when the store throws an error", func() {
		var request *http.Request

		BeforeEach(func() {
			var err error
			request, err = http.NewRequest("GET", "/networking/v0/internal/policies", nil)
			Expect(err).NotTo(HaveOccurred())
			fakeStore.AllReturns(nil, errors.New("banana"))
		})
		It("responds with 500", func() {
			var err error
			request, err = http.NewRequest("GET", "/networking/v0/internal/policies", nil)
			Expect(err).NotTo(HaveOccurred())
			handler.ServeHTTP(resp, request)

			Expect(resp.Code).To(Equal(http.StatusInternalServerError))
			Expect(resp.Body.String()).To(MatchJSON(`{"error": "database read failed"}`))
		})

		It("logs the full error", func() {
			handler.ServeHTTP(resp, request)
			Expect(logger).To(gbytes.Say("store-list-policies-failed.*banana"))
		})
	})

	Context("when the policy cannot be marshaled", func() {
		var request *http.Request

		BeforeEach(func() {
			marshaler.MarshalStub = func(interface{}) ([]byte, error) {
				return nil, errors.New("grapes")
			}

			var err error
			request, err = http.NewRequest("get", "/networking/v0/internal/policies", nil)
			Expect(err).NotTo(HaveOccurred())
		})

		It("responds with 500 and returns a descriptive error", func() {
			handler.ServeHTTP(resp, request)

			Expect(resp.Code).To(Equal(http.StatusInternalServerError))
			Expect(resp.Body.String()).To(MatchJSON(`{"error": "database marshaling failed"}`))
		})

		It("logs the full error", func() {
			handler.ServeHTTP(resp, request)
			Expect(logger).To(gbytes.Say("marshal-failed.*grapes"))
		})
	})
})
