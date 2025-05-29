package time

import (
	"testing"
	"time"

	"github.com/octohelm/x/testing/bdd"
)

func TestTimestamp(t *testing.T) {
	b := bdd.FromT(t)

	b.Given("zero time", func(b bdd.T) {
		dt := TimestampZero

		b.Then("output empty",
			bdd.Equal("", dt.String()),
		)

		b.Then("output unix sec",
			bdd.Equal(-62135596800, dt.Unix()),
		)

		b.When("marshal text", func(b bdd.T) {
			data := bdd.Must(dt.MarshalText())

			b.Then("output empty string",
				bdd.Equal("", string(data)),
			)

			b.When("unmarshal text", func(b bdd.T) {
				var dt2 Timestamp

				b.Then("success",
					bdd.NoError(dt2.UnmarshalText(data)),
					bdd.Equal(dt, dt2),
				)
			})
		})
	})

	b.Given("time string", func(b bdd.T) {
		t0, _ := time.Parse(time.RFC3339, "2017-03-27T23:58:59+08:00")
		dt := Timestamp(t0)

		b.Then("output RFC3339 string",
			bdd.Equal("2017-03-27T23:58:59+08:00", dt.String()),
		)

		b.Then("output unix sec",
			bdd.Equal(1490630339, dt.Unix()),
		)

		b.When("marshal text", func(b bdd.T) {
			data := bdd.Must(dt.MarshalText())

			b.Then("output as string with RFC3339",
				bdd.Equal("2017-03-27T23:58:59+08:00", string(data)),
			)

			b.When("unmarshal text", func(b bdd.T) {
				var dt2 Timestamp

				b.Then("success",
					bdd.NoError(dt2.UnmarshalText(data)),
					bdd.Equal(dt, dt2),
				)
			})
		})
	})

	b.Given("time string for custom output layout", func(b bdd.T) {
		t0, _ := time.Parse(time.RFC3339, "2017-03-27T23:58:59+08:00")
		dt := Timestamp(t0)

		b.When("marshal text", func(b bdd.T) {
			SetOutput("2006-01-02 15:04:05", nil)

			data := bdd.Must(dt.MarshalText())

			b.Then("output as string with custom output layout",
				bdd.Equal("2017-03-27 23:58:59", string(data)),
			)

			b.When("unmarshal text", func(b bdd.T) {
				var dt2 Timestamp

				b.Then("success",
					bdd.NoError(dt2.UnmarshalText(data)),
					bdd.Equal(dt.Unix(), dt2.Unix()),
				)
			})
		})
	})
}
