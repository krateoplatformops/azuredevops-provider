package v1alpha2

import "github.com/krateoplatformops/provider-runtime/pkg/resource"

// GetItems of this PipelinePermissionList.
func (l *PipelinePermissionList) GetItems() []resource.Managed {
	items := make([]resource.Managed, len(l.Items))
	for i := range l.Items {
		items[i] = &l.Items[i]
	}
	return items
}
