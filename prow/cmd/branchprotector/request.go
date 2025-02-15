/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	branchprotection "k8s.io/test-infra/prow/config"
	"k8s.io/test-infra/prow/github"

	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/sets"
)

// makeRequest renders a branch protection policy into the corresponding GitHub api request.
func makeRequest(policy branchprotection.Policy, enableAppsRestrictions bool) github.BranchProtectionRequest {
	return github.BranchProtectionRequest{
		EnforceAdmins:              makeAdmins(policy.Admins),
		RequiredPullRequestReviews: makeReviews(policy.RequiredPullRequestReviews),
		RequiredStatusChecks:       makeChecks(policy.RequiredStatusChecks),
		Restrictions:               makeRestrictions(policy.Restrictions, enableAppsRestrictions),
		RequiredLinearHistory:      makeBool(policy.RequiredLinearHistory),
		AllowForcePushes:           makeBool(policy.AllowForcePushes),
		AllowDeletions:             makeBool(policy.AllowDeletions),
	}

}

// makeAdmins returns true iff *val == true, else false
// TODO(skuznets): the API documentation tells us to pass
//    `nil` to unset, but that is broken so we need to pass
//    false. Change back when it's fixed
func makeAdmins(val *bool) *bool {
	if val != nil {
		return val
	}
	no := false
	return &no
}

// makeBool returns true iff *val == true
func makeBool(val *bool) bool {
	return val != nil && *val
}

// makeChecks renders a ContextPolicy into the corresponding GitHub api object.
//
// Returns nil when input policy is nil.
// Otherwise returns non-nil Contexts (empty if unset) and Strict if Strict is true
func makeChecks(cp *branchprotection.ContextPolicy) *github.RequiredStatusChecks {
	if cp == nil {
		return nil
	}
	return &github.RequiredStatusChecks{
		Contexts: append([]string{}, sets.NewString(cp.Contexts...).List()...),
		Strict:   makeBool(cp.Strict),
	}
}

// makeDismissalRestrictions renders restrictions into the corresponding GitHub api object.
//
// Returns nil when input restrictions is nil.
// Otherwise Teams and Users are both non-nil (empty list if unset).
func makeDismissalRestrictions(rp *branchprotection.DismissalRestrictions) *github.DismissalRestrictionsRequest {
	if rp == nil {
		return nil
	}
	teams := append([]string{}, sets.NewString(rp.Teams...).List()...)
	users := append([]string{}, sets.NewString(rp.Users...).List()...)
	return &github.DismissalRestrictionsRequest{
		Teams: &teams,
		Users: &users,
	}
}

// makeRestrictions renders restrictions into the corresponding GitHub api object.
//
// Returns nil when input restrictions is nil.
// Otherwise Teams and Users are non-nil (empty list if unset).
// If enableAppsRestrictions is set Apps behave like Teams and Users, otherwise Apps are nil
func makeRestrictions(rp *branchprotection.Restrictions, enableAppsRestrictions bool) *github.RestrictionsRequest {
	if rp == nil {
		return nil
	}
	// Only set restriction request for apps if feature flag is true
	// TODO: consider removing feature flag in the future
	var apps *[]string
	if enableAppsRestrictions {
		a := append([]string{}, sets.NewString(rp.Apps...).List()...)
		apps = &a
	}
	teams := append([]string{}, sets.NewString(rp.Teams...).List()...)
	users := append([]string{}, sets.NewString(rp.Users...).List()...)
	return &github.RestrictionsRequest{
		Apps:  apps,
		Teams: &teams,
		Users: &users,
	}
}

// makeReviews renders review policy into the corresponding GitHub api object.
//
// Returns nil if the policy is nil, or approvals is nil or 0.
func makeReviews(rp *branchprotection.ReviewPolicy) *github.RequiredPullRequestReviewsRequest {
	switch {
	case rp == nil:
		return nil
	case rp.Approvals == nil:
		logrus.Warn("WARNING: required_pull_request_reviews policy does not specify required_approving_review_count, disabling")
		return nil
	case *rp.Approvals == 0:
		return nil
	}
	rprr := github.RequiredPullRequestReviewsRequest{
		DismissStaleReviews:          makeBool(rp.DismissStale),
		RequireCodeOwnerReviews:      makeBool(rp.RequireOwners),
		RequiredApprovingReviewCount: *rp.Approvals,
	}
	if rp.DismissalRestrictions != nil {
		rprr.DismissalRestrictions = *makeDismissalRestrictions(rp.DismissalRestrictions)
	}
	return &rprr
}
