package sqltype_test

import (
	"database/sql/driver"
	"testing"

	. "github.com/octohelm/x/testing/v2"

	sqltype "github.com/octohelm/storage/pkg/sqltype"
	"github.com/octohelm/storage/testdata/model"
)

type deletedAtGetter struct{}

func (deletedAtGetter) GetDeletedAt() driver.Value { return int64(1) }

func TestInterfaces(t *testing.T) {
	Then(
		t, "HasSoftDelete 识别实现 WithSoftDelete 的模型",
		Expect(sqltype.HasSoftDelete[model.User](), Equal(true)),
		Expect(sqltype.HasSoftDelete[model.OrgUser](), Equal(false)),
	)

	Then(
		t, "SoftDeleteValueGetter 可被实现",
		Expect(deletedAtGetter{}.GetDeletedAt(), Equal(driver.Value(int64(1)))),
	)
}
