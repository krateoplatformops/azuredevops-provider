package endpoints

import (
	"reflect"
	"strings"

	"github.com/krateoplatformops/provider-runtime/pkg/helpers"
)

func Equal(a *ServiceEndpoint, b *ServiceEndpoint) bool {
	if helpers.String(a.Id) != helpers.String(b.Id) {
		return false
	}

	if helpers.String(a.Url) != helpers.String(b.Url) {
		return false
	}
	if helpers.String(a.Description) != helpers.String(b.Description) {
		return false
	}

	if helpers.String(a.Type) != helpers.String(b.Type) {
		return false
	}

	if !strings.EqualFold(helpers.String(a.Owner), helpers.String(b.Owner)) {
		return false
	}

	if !reflect.DeepEqual(a.Data, b.Data) {
		return false
	}
	found := 0
	for _, ref := range a.ServiceEndpointProjectReferences {
		for _, refb := range b.ServiceEndpointProjectReferences {
			if reflect.DeepEqual(ref, refb) {
				found++
			}
		}
	}

	return found == len(a.ServiceEndpointProjectReferences)
}
