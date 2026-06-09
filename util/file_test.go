package util

import (
	"reflect"
	"testing"
)

func TestExtractFileNamesFindsArticleResourceNames(t *testing.T) {
	content := `
		<figure class="image">
		<img src="/resources/articles/6d5f/relative-image.png" alt="relative">
		<img src="resources/articles/6d5f/no-leading-slash.webp" alt="relative">
		</figure>
		<img src="http://localhost:8080/resources/articles/6d5f/absolute-image.jpg?t=123">
		<a href="https://example.com/resources/articles/6d5f/document.pdf#page=1">file</a>
		<img src="/resources/articles/6d5f/relative-image.png">
		<a href="https://example.com/not-an-upload">external link</a>
	`

	got := ExtractFileNames(content)
	want := []string{"relative-image.png", "no-leading-slash.webp", "absolute-image.jpg", "document.pdf"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ExtractFileNames() = %#v, want %#v", got, want)
	}
}
