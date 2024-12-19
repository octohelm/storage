package sqlfrag

import (
	testingx "github.com/octohelm/x/testing"
	"testing"
)

func TestSafeProjected(t *testing.T) {
	testingx.Expect(t,
		SafeProjected("t_vendor_vendor_artifact_version_annotation", "f_id"),
		testingx.Be("t_vendor_vendor_artifact_version_annotation__f_id"),
	)

	testingx.Expect(t,
		SafeProjected("t_vendor_vendor_artifact_version_annotation", "f_artifact_version_id"),
		testingx.Be("t_vendor_vendor_49fecef6__f_artifact_version_id"),
	)

	testingx.Expect(t,
		SafeProjected("t_vendor_vendor_artifact_version_annotation_0", "f_artifact_version_id"),
		testingx.Be("t_vendor_vendor_6c234282__f_artifact_version_id"),
	)
}
