package checkedexamples

import (
	"net/http"
	"testing"

	"github.com/verity-bdd/verity-bdd/verity_abilities/api"
)

func TestRequestBuilderExample(t *testing.T) {
	builder := api.RequestFor(http.MethodPost, "https://api.example.org/posts").
		WithHeader("Accept", "application/json")
	if err := builder.WithJSONBody(map[string]string{"title": "checked"}); err != nil {
		t.Fatal(err)
	}
	request, err := builder.Build()
	if err != nil {
		t.Fatal(err)
	}
	_ = api.LastResponseHeader("content-type")
	_ = api.LastResponseBodyAtJSONPath("data.user.name")
	_ = request
}
