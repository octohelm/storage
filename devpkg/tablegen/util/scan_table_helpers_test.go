package util

import (
	"testing"
)

func TestToDefaultTableName(t *testing.T) {
	cases := map[string]struct {
		name       string
		tableGroup string
		want       string
	}{
		"无分组时直接转下划线": {
			name: "UserProfile",
			want: "t_user_profile",
		},
		"分组与名称完全一致时不重复前缀": {
			name:       "Org",
			tableGroup: "org",
			want:       "t_org",
		},
		"名称包含分组前缀时复用后缀": {
			name:       "OrgUser",
			tableGroup: "org",
			want:       "t_org_user",
		},
		"名称不含分组前缀时补全分组": {
			name:       "Member",
			tableGroup: "org",
			want:       "t_org_member",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := toDefaultTableName(tc.name, tc.tableGroup); got != tc.want {
				t.Fatalf("toDefaultTableName(%q, %q) = %q, want %q", tc.name, tc.tableGroup, got, tc.want)
			}
		})
	}
}

func TestCommentAndDesc(t *testing.T) {
	comment, desc := commentAndDesc([]string{"", "首行说明", "补充一", "补充二"})

	if comment != "首行说明" {
		t.Fatalf("commentAndDesc comment = %q, want %q", comment, "首行说明")
	}

	if len(desc) != 2 || desc[0] != "补充一" || desc[1] != "补充二" {
		t.Fatalf("commentAndDesc desc = %#v, want %#v", desc, []string{"补充一", "补充二"})
	}
}
