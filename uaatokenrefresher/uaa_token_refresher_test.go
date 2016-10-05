package uaatokenrefresher_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-community/firehose_exporter/uaatokenrefresher"
	"github.com/cloudfoundry-community/firehose_exporter/uaatokenrefresher/fakes"
)

var _ = Describe("UAATokenRefresher", func() {
	var (
		err       error
		fakeToken string

		fakeUAA            *fakes.FakeUAA
		authTokenRefresher *uaatokenrefresher.UAATokenRefresher
	)

	BeforeEach(func() {
		fakeUAA = fakes.NewFakeUAA("bearer", "123456789")
		fakeToken = fakeUAA.AuthToken()
		fakeUAA.Start()

		authTokenRefresher, err = uaatokenrefresher.New(
			fakeUAA.URL(), "client-id", "client-secret", true,
		)
	})

	It("fetches a token from the UAA", func() {
		authToken, err := authTokenRefresher.RefreshAuthToken()
		Expect(fakeUAA.Requested()).To(BeTrue())
		Expect(authToken).To(Equal(fakeToken))
		Expect(err).ToNot(HaveOccurred())
	})
})
