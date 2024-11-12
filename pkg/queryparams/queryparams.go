package queryparams

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

/*


false => field = false
1     => field = 1
%123% => field like %123%
%123 =>  field like %123
123% => filed like 123%

(,1)    => field < 1
(1,)    => filed > 1
[,1]    => filed <= 1
[1,]    => filed >= 1
[1,2] => filed >= 1 and field <= 2
(1,2) => file > 1 and filed < 2
t1629771113t  => filed = 2021-08-24 10:11:53

{1,2}   => filed in (1,2)
~1      => filed <> 1
~{1,2}  => filed not in (1,2)

e@is null@e      => field is null
e@is not null@e  => field is not null

filed1||field2 = 3 => filed1 = 3 or filed2 = 3

m[123]m => MATCH(field) AGAINST('123')
*/
func plainUnit(field string, v ...string) (condition string, values []interface{}) {
	conditions := make([]string, 0)
	fs := strings.Split(field, "||")
	for _, f := range fs {
		if len(v) > 1 {
			unitConditions := make([]string, 0)
			for _, vv := range v {
				c, vs := plainUnit(f, vv)
				if c != "" {
					unitConditions = append(unitConditions, c)
					values = append(values, vs...)
				}
			}
			unitCondition := strings.Join(unitConditions, " AND ")
			conditions = append(conditions, fmt.Sprintf("(%s)", unitCondition))
			continue
		}

		value := v[0]
		if value == "" {
			continue
		}
		if !strings.EqualFold(value, "1") && !strings.EqualFold(value, "0") {
			b, err := strconv.ParseBool(value)
			if err == nil {
				conditions = append(conditions, fmt.Sprintf("`%s` = ?", f))
				values = append(values, b)
				continue
			}
		}

		if strings.HasPrefix(value, "%") || strings.HasSuffix(value, "%") {
			conditions = append(conditions, fmt.Sprintf("`%s` LIKE ?", f))
			values = append(values, value)
			continue
		}

		if strings.HasPrefix(value, "(") && strings.HasSuffix(value, ")") {
			rang := strings.Split(value[1:len(value)-1], ",")
			if len(rang) != 2 {
				continue
			}
			if rang[0] == "" && rang[1] != "" {
				conditions = append(conditions, fmt.Sprintf("`%s` < ?", f))
				values = append(values, wrapValueType(rang[1]))
			}
			if rang[0] != "" && rang[1] == "" {
				conditions = append(conditions, fmt.Sprintf("`%s` > ?", f))
				values = append(values, wrapValueType(rang[0]))
			}

			if rang[0] != "" && rang[1] == "" {
				conditions = append(conditions, fmt.Sprintf("`%s` > ? AND `%s` < ?", f, f))
				values = append(values, wrapValueType(rang[0]), wrapValueType(rang[1]))
			}
			continue

		}

		if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
			rang := strings.Split(value[1:len(value)-1], ",")
			if len(rang) != 2 {
				continue
			}
			if rang[0] == "" && rang[1] != "" {
				conditions = append(conditions, fmt.Sprintf("`%s` <= ?", f))
				values = append(values, wrapValueType(rang[1]))
			}
			if rang[0] != "" && rang[1] == "" {
				conditions = append(conditions, fmt.Sprintf("`%s` >= ?", f))
				values = append(values, wrapValueType(rang[0]))
			}

			if rang[0] != "" && rang[1] == "" {
				conditions = append(conditions, fmt.Sprintf("`%s` >= ? AND `%s` <= ?", f, f))
				values = append(values, wrapValueType(rang[0]), wrapValueType(rang[1]))
			}
			continue

		}

		if strings.HasPrefix(value, "{") && strings.HasSuffix(value, "}") {
			conditions = append(conditions, fmt.Sprintf("`%s` in (?)", f))
			vs := strings.Split(value[1:len(value)-1], ",")
			values = append(values, vs)
			continue
		}

		if strings.HasPrefix(value, "~") {
			if strings.HasPrefix(value, "~{") && strings.HasSuffix(value, "}") {
				conditions = append(conditions, fmt.Sprintf("`%s` not in (?)", f))
				values = append(values, strings.Split(value[2:len(value)-1], ","))
			} else {
				conditions = append(conditions, fmt.Sprintf("`%s` <> ?", f))
				values = append(values, value[1:])
			}
			continue
		}

		if value == "e@is null@e" {
			conditions = append(conditions, fmt.Sprintf("`%s` is null", f))
			continue

		}
		if value == "e@is not null@e" {
			conditions = append(conditions, fmt.Sprintf("`%s` is not null", f))
			continue
		}

		if strings.HasPrefix(value, "m[") && strings.HasSuffix(value, "]m") {
			conditions = append(conditions, fmt.Sprintf("MATCH(`%s`) AGAINST(?)", f))
			values = append(values, value[2:len(value)-2])
			continue
		}

		conditions = append(conditions, fmt.Sprintf("`%s` = ?", f))
		values = append(values, wrapValueType(value))

	}

	condition = strings.Join(conditions, " OR ")
	if len(conditions) > 1 {
		condition = fmt.Sprintf("(%s)", condition)
	}
	return
}

func wrapValueType(v string) interface{} {
	if strings.HasPrefix(v, "t") && strings.HasSuffix(v, "t") {
		t, err := cast.ToInt64E(v[1 : len(v)-1])
		if err != nil {
			return v
		}
		return time.Unix(t, 0).Format(time.RFC3339)
	} else {
		return v
	}
}

type Query struct {
	url.Values
}

func (q Query) Plains(fields ...string) []interface{} {
	var conditions []string
	var values []interface{}
	for field, v := range q.Values {
		if !includeOr(field) {
			if !inSlice(field, fields) {
				continue
			}
		}
		c, vs := plainUnit(field, v...)
		if c != "" {
			conditions = append(conditions, c)
			values = append(values, vs...)
		}
	}
	condition := strings.Join(conditions, " AND ")
	return append([]interface{}{condition}, values...)
}

func inSlice(field string, fields []string) bool {
	for _, f := range fields {
		if f == field {
			return true
		}
	}
	return false
}

func includeOr(field string) bool {
	if strings.Contains(field, "||") {
		return true
	}
	return false
}

const (
	fieldOffset   = "offset"
	fieldLimit    = "limit"
	fieldOrder    = "order_by"
	fieldPreload  = "populate"
	defaultLimit  = 20
	defaultOffset = 0
	maxLimit      = 100
)

type QueryParams struct {
	Query       Query                    `json:"query"`
	Limit       int                      `json:"limit"`
	Offset      int                      `json:"offset"`
	Order       string                   `json:"order"`
	Group       string                   `json:"group"`
	Having      string                   `json:"having"`
	Joins       string                   `json:"joins"`
	CustomQuery map[string][]interface{} `json:"custom_query"`
	Select      string                   `json:"select"`
	TableName   string                   `json:"table_name"`
	Preload     []string                 `json:"preload"`
}

func NewQueryParams(c *gin.Context) *QueryParams {
	values := c.Request.URL.Query()
	params := &QueryParams{
		Limit:  defaultLimit,
		Offset: defaultOffset,
	}
	limit := cast.ToInt(c.Query(fieldLimit))
	if limit > 0 {
		if limit > maxLimit {
			params.Limit = maxLimit
		} else {
			params.Limit = limit
		}
	}
	values.Del(fieldLimit)

	offset := cast.ToInt(c.Query(fieldOffset))
	if offset > 0 {
		params.Offset = offset
	}
	values.Del(fieldOffset)
	orders := c.QueryArray(fieldOrder)
	sortBys := make([]string, 0)
	for _, order := range orders {
		if strings.HasPrefix(order, "-") {
			order = strings.TrimLeft(order, "-")
			sortBys = append(sortBys, fmt.Sprintf("%s %s", order, "desc"))
		} else {
			sortBys = append(sortBys, fmt.Sprintf("%s %s", order, "asc"))
		}
	}
	params.Order = strings.Join(sortBys, ",")

	values.Del(fieldOrder)

	preload := cast.ToString(c.Query(fieldPreload))
	if len(preload) > 0 {
		preload = strings.Trim(preload, "[]")
		preloads := strings.Split(preload, ",")
		for _, p := range preloads {
			p = strings.Trim(p, "\"")
			params.Preload = append(params.Preload, strings.Title(p))
		}
		values.Del(fieldPreload)
	}
	params.Query = Query{Values: values}
	params.CustomQuery = make(map[string][]interface{})

	return params
}

func NewCustomQueryParams() *QueryParams {
	qp := &QueryParams{
		Query:       Query{Values: make(url.Values)},
		Limit:       defaultLimit,
		CustomQuery: map[string][]interface{}{},
	}
	return qp
}

func (q *QueryParams) Add(key, value string) {
	q.Query.Add(key, value)
}

func (q *QueryParams) Del(key string) {
	q.Query.Del(key)
}

func (q *QueryParams) Get(key string) string {
	return q.Query.Get(key)
}

func (q *QueryParams) Set(key string, values []string) {
	q.Query.Values[key] = values
}

func (q *QueryParams) Replace(oldKey, newKey string) {
	value, ok := q.Query.Values[oldKey]
	if ok {
		q.Set(newKey, value)
		q.Del(oldKey)
	}
}

func (q *QueryParams) ReplaceHas2Compare(oldKey, newKey string) {
	if has := q.Get(oldKey); len(has) > 0 {
		if cast.ToBool(has) {
			q.Set(newKey, []string{"[0,]"})
		} else {
			q.Set(newKey, []string{"0"})
		}
		q.Del(oldKey)
	}
}

func (q *QueryParams) AddCustomQuery(queryStr string, value ...interface{}) {
	q.CustomQuery[queryStr] = value
}

func (q *QueryParams) DelCustomQuery(queryStr string) {
	delete(q.CustomQuery, queryStr)
}


func (q *QueryParams) AddSelect(selectStr string) {
	q.Select = selectStr
}
