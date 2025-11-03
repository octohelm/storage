package sqlpipe_test

import (
	"testing"

	testingx "github.com/octohelm/x/testing"

	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlbuilder/modelscoped"
	"github.com/octohelm/storage/pkg/sqlfrag"
	"github.com/octohelm/storage/pkg/sqlfrag/testutil"
	"github.com/octohelm/storage/pkg/sqlpipe"
	"github.com/octohelm/storage/testdata/model"
	modelaggregate "github.com/octohelm/storage/testdata/model/aggregate"
)

func TestAggregate(t *testing.T) {
	t.Run("aggregate", func(t *testing.T) {
		aggr := sqlpipe.Aggregate[model.User, modelaggregate.CountedUser](
			sqlpipe.FromAll[model.User](),

			modelaggregate.CountedUserT.Count.TypedComputedBy(sqlbuilder.Count()),
		)

		t.Run("exec", func(t *testing.T) {
			testingx.Expect[sqlfrag.Fragment](t, aggr, testutil.BeFragment(`
SELECT COUNT(1) AS f_count
FROM t_user
`))
		})
	})

	t.Run("aggregate from", func(t *testing.T) {
		src := sqlpipe.FromAll[model.User]().Pipe(
			sqlpipe.Limit[model.User](10),
		)

		aggr := sqlpipe.Aggregate[model.User, modelaggregate.CountedUser](
			src,
			modelaggregate.CountedUserT.Count.TypedComputedBy(sqlbuilder.Count()),
		)

		testingx.Expect[sqlfrag.Fragment](t, aggr, testutil.BeFragment(`
SELECT COUNT(1) AS f_count
FROM (
	SELECT *
	FROM t_user
	LIMIT 10
) AS t_user
`))
	})

	t.Run("group by", func(t *testing.T) {
		src := sqlpipe.FromAll[model.User]().Pipe(
			sqlpipe.Limit[model.User](10),
		)

		aggr := sqlpipe.AggregateGroupBy[model.User, modelaggregate.CountedUser](
			src,

			modelscoped.AllColumns(model.UserT.Age),

			modelaggregate.CountedUserT.Age,
			modelaggregate.CountedUserT.Count.ComputedBy(sqlbuilder.Count(modelaggregate.CountedUserT.Age)),
		)

		t.Run("exec", func(t *testing.T) {
			testingx.Expect[sqlfrag.Fragment](t, aggr, testutil.BeFragment(`
SELECT f_age, COUNT(f_age) AS f_count
FROM (
	SELECT *
	FROM t_user
	LIMIT 10
) AS t_user
GROUP BY f_age
`))
		})
	})

	t.Run("group by from", func(t *testing.T) {
		aggr := sqlpipe.AggregateGroupBy[model.User, modelaggregate.CountedUser](
			sqlpipe.FromAll[model.User](),

			modelscoped.AllColumns[model.User](model.UserT.Age),

			modelaggregate.CountedUserT.Age,
			modelaggregate.CountedUserT.Count.ComputedBy(sqlbuilder.Count(model.UserT.Age)),
		)

		t.Run("exec", func(t *testing.T) {
			testingx.Expect[sqlfrag.Fragment](t, aggr, testutil.BeFragment(`
SELECT f_age, COUNT(f_age) AS f_count
FROM t_user
GROUP BY f_age
`))
		})

		t.Run("then where", func(t *testing.T) {
			filtered := aggr.Pipe(
				sqlpipe.Where(modelaggregate.CountedUserT.Count, sqlbuilder.Gt(10)),
			)

			testingx.Expect[sqlfrag.Fragment](t, filtered, testutil.BeFragment(`
SELECT f_age, f_count
FROM (
	SELECT f_age, COUNT(f_age) AS f_count
	FROM t_user
	GROUP BY f_age
) AS t_counted_user
WHERE f_count > ?
`, 10))
		})
	})
}
