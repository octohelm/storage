package sqlpipe

import (
	"iter"
	"testing"

	. "github.com/octohelm/x/testing/v2"

	"github.com/octohelm/storage/testdata/model"
)

func TestValueConstructors(t *testing.T) {
	empty := Values([]*model.User(nil))
	Then(
		t, "空 Values 返回 noop source",
		Expect((*noop[model.User])(nil) == nil, Equal(true)),
		Expect(empty.IsNil(), Equal(true)),
	)

	src := Value(&model.User{Name: "alice"}, model.UserT.Name)
	srcOmit := ValueOmit(&model.User{Name: "alice"}, model.UserT.Age)
	srcOmitZero := ValueOmitZero(&model.User{Name: "alice"}, model.UserT.Name)
	srcSeq := ValueSeq(iter.Seq[*model.User](func(yield func(*model.User) bool) {
		yield(&model.User{Name: "alice"})
	}), model.UserT.Name)
	srcSeqOmit := ValueSeqOmit(iter.Seq[*model.User](func(yield func(*model.User) bool) {
		yield(&model.User{Name: "alice"})
	}), model.UserT.Age)

	Then(
		t, "单值和序列构造器都会生成可用 source",
		Expect(src.IsNil(), Equal(false)),
		Expect(srcOmit.IsNil(), Equal(false)),
		Expect(srcOmitZero.IsNil(), Equal(false)),
		Expect(srcSeq.IsNil(), Equal(false)),
		Expect(srcSeqOmit.IsNil(), Equal(false)),
	)
}
