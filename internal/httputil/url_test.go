package httputil_test

import (
	"testing"

	"gihtub.com/krateoplatformops/azuredevops-provider/internal/httputil"
	"github.com/google/go-cmp/cmp"
)

func TestBuildURL(t *testing.T) {
	url, err := httputil.BuildURL("https://dev.azure.com", []string{
		"my-org",
		"_apis/projects",
		"1234",
	}, httputil.NewMultimap("api-version", "7.0"))
	if err != nil {
		t.Fatal(err)
	}

	want := "https://dev.azure.com/my-org/_apis/projects/1234?api-version=7.0"
	got := url.String()
	if !cmp.Equal(got, want) {
		t.Fatalf("expected: %v, got: %v", want, got)
	}

}
