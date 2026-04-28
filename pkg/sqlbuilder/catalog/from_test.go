package catalog

import (
	"testing"

	. "github.com/octohelm/x/testing/v2"

	"github.com/octohelm/storage/testdata/model"
)

func TestFrom(t *testing.T) {
	c := From(&model.User{}, &model.Org{})

	Then(t, "From 收集模型表",
		Expect(c.Table("t_user").TableName(), Equal("t_user")),
		Expect(c.Table("t_org").TableName(), Equal("t_org")),
	)
}
