package v1alpha2

import (
	rtv1 "github.com/krateoplatformops/provider-runtime/apis/common/v1"
)

func (mg *PipelinePermission) GetCondition(ct rtv1.ConditionType) rtv1.Condition {
	return mg.Status.GetCondition(ct)
}

func (mg *PipelinePermission) GetDeletionPolicy() rtv1.DeletionPolicy {
	return mg.Spec.DeletionPolicy
}

func (mg *PipelinePermission) SetConditions(c ...rtv1.Condition) {
	mg.Status.SetConditions(c...)
}

func (mg *PipelinePermission) SetDeletionPolicy(r rtv1.DeletionPolicy) {
	mg.Spec.DeletionPolicy = r
}
