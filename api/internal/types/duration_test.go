package types_test

import (
	"encoding/json"
	"time"

	"github.com/pivotal-cf/azure-blobstore-resource/api/internal/types"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type TestMessage struct {
	MyDuration types.MarshalableDuration `json:"my_duration"`
}

var _ = Describe("MarshalableDuration", func() {
	It("marshals into json", func() {
		contents, err := json.Marshal(TestMessage{
			MyDuration: types.MarshalableDuration(time.Minute * 23),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(contents).To(MatchJSON(`{
			"my_duration": "23m0s"
		}`))
	})

	It("unmarshals a well-formed duration string into a time duration type", func() {
		var testMessage TestMessage
		err := json.Unmarshal([]byte(`{
			"my_duration": "32h"
		}`), &testMessage)
		Expect(err).NotTo(HaveOccurred())
		Expect(testMessage).To(Equal(TestMessage{
			MyDuration: types.MarshalableDuration(time.Hour * 32),
		}))
	})

	It("unmarshals an integer into a time duration type", func() {
		var testMessage TestMessage
		err := json.Unmarshal([]byte(`{
			"my_duration": 1000000
		}`), &testMessage)
		Expect(err).NotTo(HaveOccurred())
		Expect(testMessage).To(Equal(TestMessage{
			MyDuration: types.MarshalableDuration(time.Millisecond * 1),
		}))
	})

	It("returns an error when an invalid type is provided", func() {
		var testMessage TestMessage
		err := json.Unmarshal([]byte(`{
			"my_duration": {}
		}`), &testMessage)
		Expect(err).To(MatchError("duration must be a string or integer"))
	})
})
