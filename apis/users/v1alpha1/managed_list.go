package v1alpha1

import "github.com/krateoplatformops/provider-runtime/pkg/resource"

// GetItems of this UsersList.
func (l *UsersList) GetItems() []resource.Managed {
	items := make([]resource.Managed, len(l.Items))
	for i := range l.Items {
		items[i] = &l.Items[i]
	}
	return items
}
