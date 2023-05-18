package v1alpha1

import (
	rtv1 "github.com/krateoplatformops/provider-runtime/apis/common/v1"
)

// GetCondition of this Endpoint.
func (mg *Endpoint) GetCondition(ct rtv1.ConditionType) rtv1.Condition {
	return mg.Status.GetCondition(ct)
}

// GetDeletionPolicy of this Endpoint.
func (mg *Endpoint) GetDeletionPolicy() rtv1.DeletionPolicy {
	return mg.Spec.DeletionPolicy
}

// SetConditions of this Endpoint.
func (mg *Endpoint) SetConditions(c ...rtv1.Condition) {
	mg.Status.SetConditions(c...)
}

// SetDeletionPolicy of this Endpoint.
func (mg *Endpoint) SetDeletionPolicy(r rtv1.DeletionPolicy) {
	mg.Spec.DeletionPolicy = r
}
