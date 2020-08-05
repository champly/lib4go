package tool

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("guid", func() {
	Describe("build guid", func() {
		Context("length must equal 32", func() {
			It("should be a novel", func() {
				Expect(GetGUID()).Should(&matcher{})
			})
		})
	})
})

type matcher struct{}

func (m *matcher) Match(actual interface{}) (success bool, err error) {
	return len(actual.(string)) == 32, nil
}

func (m *matcher) FailureMessage(actual interface{}) (message string) {
	return
}

func (m *matcher) NegatedFailureMessage(actual interface{}) (message string) {
	return
}

func TestGetGUID(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GetGUID")
}
