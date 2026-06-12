package api

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	db "github.com/MonitorAllen/nostalgia/db/sqlc"
	"github.com/gin-gonic/gin"
)

type sitemapURLSet struct {
	XMLName xml.Name     `xml:"urlset"`
	XMLNS   string       `xml:"xmlns,attr"`
	URLs    []sitemapURL `xml:"url"`
}

type sitemapURL struct {
	Loc        string `xml:"loc"`
	LastMod    string `xml:"lastmod,omitempty"`
	ChangeFreq string `xml:"changefreq,omitempty"`
	Priority   string `xml:"priority,omitempty"`
}

func (server *Server) robotsTxt(ctx *gin.Context) {
	origin := server.publicSiteOrigin(ctx)
	body := strings.Join([]string{
		"User-agent: *",
		"Disallow: /backend",
		"Disallow: /api",
		"Disallow: /v1",
		"Disallow: /setup",
		"Disallow: /login",
		"Disallow: /register",
		fmt.Sprintf("Sitemap: %s/sitemap.xml", origin),
		"",
	}, "\n")

	ctx.Data(http.StatusOK, "text/plain; charset=utf-8", []byte(body))
}

func (server *Server) sitemapXML(ctx *gin.Context) {
	categoryRows, err := server.store.ListPublishedCategorySitemapItems(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	articleRows, err := server.store.ListPublishedArticleSitemapItems(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	origin := server.publicSiteOrigin(ctx)
	urlSet := sitemapURLSet{
		XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs: []sitemapURL{
			{
				Loc:        origin + "/",
				ChangeFreq: "daily",
				Priority:   "1.0",
			},
		},
	}

	for _, category := range categoryRows {
		urlSet.URLs = append(urlSet.URLs, sitemapURL{
			Loc:        fmt.Sprintf("%s/category/%d", origin, category.ID),
			LastMod:    sitemapDate(category.UpdatedAt),
			ChangeFreq: "weekly",
			Priority:   "0.7",
		})
	}

	for _, article := range articleRows {
		urlSet.URLs = append(urlSet.URLs, sitemapURL{
			Loc:        fmt.Sprintf("%s%s", origin, sitemapArticlePath(article)),
			LastMod:    sitemapDate(articleLastModified(article.CreatedAt, article.UpdatedAt)),
			ChangeFreq: "monthly",
			Priority:   "0.8",
		})
	}

	var buffer bytes.Buffer
	buffer.WriteString(xml.Header)
	encoder := xml.NewEncoder(&buffer)
	encoder.Indent("", "  ")
	if err := encoder.Encode(urlSet); err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.Data(http.StatusOK, "application/xml; charset=utf-8", buffer.Bytes())
}

func (server *Server) publicSiteOrigin(ctx *gin.Context) string {
	if origin := normalizeOrigin(server.config.Domain); origin != "" {
		return origin
	}

	scheme := ctx.GetHeader("X-Forwarded-Proto")
	if scheme == "" {
		scheme = "http"
	}
	host := ctx.GetHeader("X-Forwarded-Host")
	if host == "" {
		host = ctx.Request.Host
	}
	if host == "" {
		host = "localhost"
	}

	return normalizeOrigin(fmt.Sprintf("%s://%s", scheme, host))
}

func normalizeOrigin(value string) string {
	source := strings.TrimSpace(value)
	if source == "" {
		return ""
	}

	parsed, err := url.Parse(source)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return strings.TrimRight(source, "/")
	}

	return fmt.Sprintf("%s://%s", parsed.Scheme, parsed.Host)
}

func sitemapArticlePath(article db.ListPublishedArticleSitemapItemsRow) string {
	if article.Slug.Valid && strings.TrimSpace(article.Slug.String) != "" {
		return "/article/" + strings.TrimSpace(article.Slug.String)
	}
	return "/article/" + article.ID.String()
}

func articleLastModified(createdAt time.Time, updatedAt time.Time) time.Time {
	if updatedAt.IsZero() {
		return createdAt
	}
	return updatedAt
}

func sitemapDate(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format("2006-01-02")
}
