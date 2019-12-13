// Package pagination inspired by Spring RESTful API
// since 2019-12-13
package pagination

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type key int

const (
	pageKey key = iota
	sortKey
)

// Page ...
type Page struct {
	Page int    `json:"page"`
	Size int    `json:"size"`
	Sort []Sort `json:"sort"`
}

// WrapMySQL wrap MySQL dialect limit and order-by
func (p *Page) WrapMySQL(stmt string) string {
	parts := []string{"SELECT * FROM (", stmt, ") AS pagination_table_wrapper"}
	orders := []string{}
	if len(p.Sort) > 0 {
		for _, sort := range p.Sort {
			orders = append(orders, fmt.Sprintf("%s %s", sort.Sort, sort.Dir))
		}
		parts = append(parts, "ORDER BY "+strings.Join(orders, ","))
	}
	if p.Size > 0 && p.Page > 0 {
		parts = append(parts, fmt.Sprintf("limit %d,%d", (p.Page-1)*p.Size, p.Size))
	}
	return strings.Join(parts, " ")
}

// Sort ...
type Sort struct {
	Sort string `json:"sort"`
	Dir  string `json:"dir"`
}

// WithPage put page into context
func WithPage(ctx context.Context, page *Page) context.Context {
	return context.WithValue(ctx, pageKey, page)
}

// PageFrom get page from context
func PageFrom(ctx context.Context) (*Page, bool) {
	p, ok := ctx.Value(pageKey).(*Page)
	return p, ok
}

// Options ...
type Options struct {
	PageSize int // default page size
	Page     int // default page start from
}

var (
	// DefaultOptions ...
	DefaultOptions = Options{
		PageSize: 20,
		Page:     1,
	}
)

// Middleware assemble page & sort from request and shared them with context
//
// examples:
// 1.
// ?page=1&size=10
// 2.
// ?page=2&sort=name
// 3.
// ?page=2&sort=name,desc
func Middleware(opts *Options) func(http.Handler) http.Handler {
	if opts == nil {
		opts = &Options{}
		*opts = DefaultOptions
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var (
				err error
				q   = r.URL.Query()
				p   = Page{
					Page: opts.Page,
					Size: opts.PageSize,
					Sort: []Sort{},
				}
			)
			p.Page, err = strconv.Atoi(q.Get("page"))

			// page incorrect
			if err != nil || p.Page < 0 {
				p.Page = opts.Page
			}

			// size incorrect
			p.Size, err = strconv.Atoi(q.Get("size"))
			if err != nil {
				p.Size = opts.PageSize
			}

			// multi sort
			if sorts, ok := q["sort"]; ok {
				for _, sort := range sorts {
					if sort == "" {
						continue
					}
					s := Sort{}
					clips := strings.Split(sort, ",")
					switch len(clips) {
					case 1:
						s.Sort = clips[0]
						s.Dir = "ASC"
						p.Sort = append(p.Sort, s)
					case 2:
						s.Sort = clips[0]
						switch strings.ToUpper(clips[1]) {
						case "DESC":
							s.Dir = "DESC"
						default:
							s.Dir = "ASC"
						}
						p.Sort = append(p.Sort, s)
					default:
						// ignore
					}
				}
			}

			next.ServeHTTP(w, r.WithContext(WithPage(r.Context(), &p)))
		})
	}
}
