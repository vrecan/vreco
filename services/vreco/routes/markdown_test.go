package routes

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadBlogFolder(t *testing.T) {
	blog, err := readBlogFolder("../posts/concurrent_shutdown_with_death")
	assert.Nil(t, err)
	assert.NotEmpty(t, blog.Meta.Categories)
	assert.NotEmpty(t, blog.Meta.Date)
	assert.NotEmpty(t, blog.Meta.Description)
	assert.NotEmpty(t, blog.Meta.Tags)
	assert.NotEmpty(t, blog.Meta.Title)
	assert.NotEmpty(t, blog.Contents)
}

func TestGenerateBlogHtml(t *testing.T) {
	blogs, err := GenerateBlogHtml("../posts/")
	assert.Nil(t, err)
	assert.NotEmpty(t, blogs)

	for _, b := range blogs {
		fmt.Println(b.Meta.Date)
	}

}
