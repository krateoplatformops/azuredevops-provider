package v1alpha1

import (
	rtv1 "github.com/krateoplatformops/provider-runtime/apis/common/v1"
)

func (mg *CheckConfiguration) GetCondition(ct rtv1.ConditionType) rtv1.Condition {
	return mg.Status.GetCondition(ct)
}

func (mg *CheckConfiguration) GetDeletionPolicy() rtv1.DeletionPolicy {
	return mg.Spec.DeletionPolicy
}

func (mg *CheckConfiguration) SetConditions(c ...rtv1.Condition) {
	mg.Status.SetConditions(c...)
}

func (mg *CheckConfiguration) SetDeletionPolicy(r rtv1.DeletionPolicy) {
	mg.Spec.DeletionPolicy = r
}
