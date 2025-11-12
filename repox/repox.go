package repox

import (
	"context"
	"errors"
	"strings"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"

	"github.com/chi07/go-svc-kit/fieldx"
)

func IsDuplicateErr(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate key") || strings.Contains(msg, "unique constraint")
}

func IsNoRowsErr(err error) bool {
	return err != nil && errors.Is(err, pg.ErrNoRows)
}

func MapFieldToDB(alias string, aliasMap map[string]string) (string, bool) {
	col, ok := aliasMap[alias]
	return col, ok
}

func AllowedAliasSetLower(aliasMap map[string]string) map[string]struct{} {
	keys := make([]string, 0, len(aliasMap))
	for k := range aliasMap {
		keys = append(keys, k)
	}
	return fieldx.MakeSetLower(keys)
}

func SafeColumns(requested []string, aliasMap map[string]string) []string {
	return fieldx.FilterAllowedFieldsWithAlias(requested, aliasMap, AllowedAliasSetLower(aliasMap), nil)
}

// ---------- ORDER BY helpers (generic) ----------

func BuildOrderExpr[T any](
	items []T,
	fieldFn func(T) string,
	descFn func(T) bool,
	aliasMap map[string]string,
	defaultOrders ...string,
) []string {
	if len(items) == 0 {
		return append([]string{}, defaultOrders...)
	}
	out := make([]string, 0, len(items))
	for _, it := range items {
		alias := strings.TrimSpace(strings.ToLower(fieldFn(it)))
		if col, ok := MapFieldToDB(alias, aliasMap); ok && col != "" {
			if descFn(it) {
				out = append(out, col+" DESC")
			} else {
				out = append(out, col+" ASC")
			}
		}
	}
	if len(out) == 0 {
		return append([]string{}, defaultOrders...)
	}
	return out
}

func WithNotDeleted(q *orm.Query) *orm.Query {
	return q.Where("deleted_at IS NULL")
}

func ApplyExactFilters(q *orm.Query, extras map[string]any, aliasMap map[string]string) {
	for k, v := range extras {
		alias := strings.ToLower(strings.TrimSpace(k))
		if col, ok := MapFieldToDB(alias, aliasMap); ok && col != "" {
			q.Where(col+" = ?", v)
		}
	}
}

func ApplyLimitOffset(q *orm.Query, limit, offset int64) *orm.Query {
	if limit > 0 {
		q.Limit(int(limit))
	}
	if offset > 0 {
		q.Offset(int(offset))
	}
	return q
}

func ApplyLimitOffsetPlusOne(q *orm.Query, limit, offset int64) *orm.Query {
	if limit > 0 {
		q.Limit(int(limit + 1))
	}
	if offset > 0 {
		q.Offset(int(offset))
	}
	return q
}

func TrimHasNext[T any](rows []T, limit int64) ([]T, bool) {
	if limit <= 0 {
		return rows, false
	}
	if len(rows) > int(limit) {
		return rows[:limit], true
	}
	return rows, false
}

func Count(ctx context.Context, q *orm.Query) (int64, error) {
	n, err := q.Count()
	if err != nil {
		return 0, err
	}
	return int64(n), nil
}
