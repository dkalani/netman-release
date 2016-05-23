package handlers_test

import (
	"encoding/json"
	"lib/marshal"
	"net/http"
	"net/http/httptest"
	"policy-server/fakes"
	"policy-server/handlers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/pivotal-golang/lager/lagertest"
)

var _ = Describe("Whitelists", func() {
	Context("when no groups are specified", func() {
		It("calls GetWhitelists with a nil argument", func() {
			store := &fakes.Store{}
			handler := handlers.Whitelists{
				Marshaler: marshal.MarshalFunc(json.Marshal),
				Logger:    lagertest.NewTestLogger("test"),
				Store:     store,
			}

			response := httptest.NewRecorder()
			request, _ := http.NewRequest("GET", "/whitelists", nil)
			handler.ServeHTTP(response, request)

			Expect(store.GetWhitelistsCallCount()).To(Equal(1))
			_, groups := store.GetWhitelistsArgsForCall(0)
			Expect(groups).To(BeEmpty())
		})
	})

})
