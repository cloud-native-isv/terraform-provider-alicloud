package alicloud

import (
	"fmt"
	"strings"
)

type flinkWorkspaceProtocolContext int

const (
	flinkWorkspaceProtocolExisting flinkWorkspaceProtocolContext = iota
	flinkWorkspaceProtocolFreshCreate
	flinkWorkspaceProtocolPreparedCreate
)

type flinkWorkspaceProtocolTuple struct {
	purchase    string
	capacity    string
	visibility  string
	token       string
	fingerprint string
}

type flinkWorkspaceProtocolGetter interface {
	Get(string) interface{}
}

func flinkWorkspaceProtocolTupleFromGetter(getter flinkWorkspaceProtocolGetter) flinkWorkspaceProtocolTuple {
	return flinkWorkspaceProtocolTuple{
		purchase:    flinkWorkspaceProtocolString(getter.Get("purchase_options_state")),
		capacity:    flinkWorkspaceProtocolString(getter.Get("capacity_intent_mode")),
		visibility:  flinkWorkspaceProtocolString(getter.Get("identity_visibility_state")),
		token:       flinkWorkspaceProtocolString(getter.Get("terraform_create_token")),
		fingerprint: flinkWorkspaceProtocolString(getter.Get("create_intent_fingerprint")),
	}
}

func flinkWorkspaceProtocolString(value interface{}) string {
	text, _ := value.(string)
	return text
}

func validateFlinkWorkspaceProtocol(id string, getter flinkWorkspaceProtocolGetter, context flinkWorkspaceProtocolContext) error {
	return validateFlinkWorkspaceProtocolTuple(id, flinkWorkspaceProtocolTupleFromGetter(getter), context)
}

func validateFlinkWorkspaceProtocolTuple(id string, tuple flinkWorkspaceProtocolTuple, context flinkWorkspaceProtocolContext) error {
	if id == "" {
		if tuple == (flinkWorkspaceProtocolTuple{}) && context == flinkWorkspaceProtocolFreshCreate {
			return nil
		}
		if context == flinkWorkspaceProtocolPreparedCreate &&
			tuple.purchase == flinkWorkspacePurchaseManaged &&
			(tuple.capacity == flinkWorkspaceCapacityLegacy || tuple.capacity == flinkWorkspaceCapacityInitial) &&
			tuple.visibility == flinkWorkspaceIdentityAwaitingFirstRead &&
			flinkWorkspaceRecoveryTokenPattern.MatchString(tuple.token) &&
			flinkWorkspaceRecoveryTokenPattern.MatchString(tuple.fingerprint) {
			return nil
		}
		return flinkWorkspaceProtocolError("resource has no ID but protocol tuple is not a valid fresh or prepared Create state")
	}
	if context != flinkWorkspaceProtocolExisting {
		return flinkWorkspaceProtocolError("existing resource ID %q was validated in invalid context", id)
	}
	for name, value := range map[string]string{
		"purchase_options_state":    tuple.purchase,
		"capacity_intent_mode":      tuple.capacity,
		"identity_visibility_state": tuple.visibility,
		"terraform_create_token":    tuple.token,
		"create_intent_fingerprint": tuple.fingerprint,
	} {
		if value == "" {
			return flinkWorkspaceProtocolError("existing resource %q has missing or empty %s", id, name)
		}
	}
	if tuple.visibility != flinkWorkspaceIdentityAwaitingFirstRead && tuple.visibility != flinkWorkspaceIdentityMigratedFirstRead && tuple.visibility != flinkWorkspaceIdentityStable {
		return flinkWorkspaceProtocolError("existing resource %q has invalid identity_visibility_state %q", id, tuple.visibility)
	}
	if !flinkWorkspaceProtocolTokenValueValid(tuple.token) {
		return flinkWorkspaceProtocolError("existing resource %q has invalid terraform_create_token", id)
	}
	if !flinkWorkspaceProtocolTokenValueValid(tuple.fingerprint) {
		return flinkWorkspaceProtocolError("existing resource %q has invalid create_intent_fingerprint", id)
	}

	pendingToken, pendingID := flinkWorkspaceCreateTokenFromPendingID(id)
	if pendingID {
		if tuple.purchase != flinkWorkspacePurchaseManaged ||
			(tuple.capacity != flinkWorkspaceCapacityLegacy && tuple.capacity != flinkWorkspaceCapacityInitial) ||
			tuple.visibility != flinkWorkspaceIdentityAwaitingFirstRead ||
			!flinkWorkspaceRecoveryTokenPattern.MatchString(tuple.token) || tuple.token != pendingToken ||
			(tuple.fingerprint != flinkWorkspaceProtocolUnavailable && !flinkWorkspaceRecoveryTokenPattern.MatchString(tuple.fingerprint)) {
			return flinkWorkspaceProtocolError("pending-create resource %q has contradictory protocol tuple", id)
		}
		return nil
	}

	switch tuple.purchase {
	case flinkWorkspacePurchaseImportedUnknown:
		if tuple.capacity != flinkWorkspaceCapacityLegacy && tuple.capacity != flinkWorkspaceCapacityInitialAdoptionPending && tuple.capacity != flinkWorkspaceCapacityInitial {
			return flinkWorkspaceProtocolError("imported-unknown resource %q has contradictory capacity_intent_mode %q", id, tuple.capacity)
		}
		if tuple.visibility != flinkWorkspaceIdentityStable || tuple.token != flinkWorkspaceProtocolUnavailable || tuple.fingerprint != flinkWorkspaceProtocolUnavailable {
			return flinkWorkspaceProtocolError("imported-unknown resource %q claims paid identity visibility or provider create identity", id)
		}
	case flinkWorkspacePurchaseLegacyUnclassified:
		if tuple.capacity != flinkWorkspaceCapacityLegacy || tuple.visibility != flinkWorkspaceIdentityStable || tuple.token != flinkWorkspaceProtocolUnavailable || tuple.fingerprint != flinkWorkspaceProtocolUnavailable {
			return flinkWorkspaceProtocolError("legacy-unclassified resource %q has contradictory protocol tuple", id)
		}
	case flinkWorkspacePurchaseRecoveryPending:
		if tuple.capacity != flinkWorkspaceCapacityLegacy && tuple.capacity != flinkWorkspaceCapacityInitialAdoptionPending {
			return flinkWorkspaceProtocolError("recovery-pending resource %q has contradictory capacity_intent_mode %q", id, tuple.capacity)
		}
		if !flinkWorkspaceRecoveryTokenPattern.MatchString(tuple.token) {
			return flinkWorkspaceProtocolError("recovery-pending resource %q has no valid create token", id)
		}
		if tuple.visibility == flinkWorkspaceIdentityStable && !flinkWorkspaceRecoveryTokenPattern.MatchString(tuple.fingerprint) {
			return flinkWorkspaceProtocolError("stable recovery-pending resource %q has no authoritative create intent fingerprint", id)
		}
		if tuple.visibility == flinkWorkspaceIdentityAwaitingFirstRead && tuple.fingerprint != flinkWorkspaceProtocolUnavailable && !flinkWorkspaceRecoveryTokenPattern.MatchString(tuple.fingerprint) {
			return flinkWorkspaceProtocolError("recovery-pending resource %q has invalid pre-read create intent fingerprint", id)
		}
		if tuple.visibility == flinkWorkspaceIdentityMigratedFirstRead {
			return flinkWorkspaceProtocolError("recovery-pending resource %q cannot claim schema-v0 migration visibility", id)
		}
	case flinkWorkspacePurchaseMigratedInitial:
		if tuple.capacity != flinkWorkspaceCapacityInitial ||
			tuple.visibility != flinkWorkspaceIdentityMigratedFirstRead ||
			!flinkWorkspaceRecoveryTokenPattern.MatchString(tuple.token) ||
			tuple.fingerprint != flinkWorkspaceProtocolUnavailable {
			return flinkWorkspaceProtocolError("schema-v0 initial migration resource %q has contradictory protocol tuple", id)
		}
	case flinkWorkspacePurchaseManaged:
		if tuple.capacity != flinkWorkspaceCapacityLegacy && tuple.capacity != flinkWorkspaceCapacityInitial {
			return flinkWorkspaceProtocolError("managed resource %q has contradictory capacity_intent_mode %q", id, tuple.capacity)
		}
		validIdentity := (tuple.visibility == flinkWorkspaceIdentityAwaitingFirstRead || tuple.visibility == flinkWorkspaceIdentityStable) &&
			flinkWorkspaceRecoveryTokenPattern.MatchString(tuple.token) && flinkWorkspaceRecoveryTokenPattern.MatchString(tuple.fingerprint)
		legacyIdentity := tuple.visibility == flinkWorkspaceIdentityStable &&
			tuple.token == flinkWorkspaceProtocolUnavailable && tuple.fingerprint == flinkWorkspaceProtocolUnavailable
		if !validIdentity && !legacyIdentity {
			return flinkWorkspaceProtocolError("managed resource %q has contradictory identity token/fingerprint tuple", id)
		}
	default:
		return flinkWorkspaceProtocolError("existing resource %q has invalid purchase_options_state %q", id, tuple.purchase)
	}
	return nil
}

func flinkWorkspaceProtocolTokenValueValid(value string) bool {
	return value == flinkWorkspaceProtocolUnavailable || flinkWorkspaceRecoveryTokenPattern.MatchString(value)
}

func flinkWorkspaceProtocolError(format string, args ...interface{}) error {
	return fmt.Errorf("invalid Flink workspace lifecycle protocol: "+format, args...)
}

func flinkWorkspaceProtocolTupleFromResourceDataChange(d interface {
	Get(string) interface{}
	GetChange(string) (interface{}, interface{})
}, old bool) flinkWorkspaceProtocolTuple {
	value := func(key string) string {
		before, after := d.GetChange(key)
		if old {
			return flinkWorkspaceProtocolString(before)
		}
		if after == nil {
			return flinkWorkspaceProtocolString(d.Get(key))
		}
		return flinkWorkspaceProtocolString(after)
	}
	return flinkWorkspaceProtocolTuple{
		purchase:    value("purchase_options_state"),
		capacity:    value("capacity_intent_mode"),
		visibility:  value("identity_visibility_state"),
		token:       value("terraform_create_token"),
		fingerprint: value("create_intent_fingerprint"),
	}
}

func validateFlinkWorkspaceProtocolTransition(id string, oldTuple, newTuple flinkWorkspaceProtocolTuple) error {
	if oldTuple == newTuple {
		return nil
	}
	if oldTuple.purchase == flinkWorkspacePurchaseRecoveryPending &&
		newTuple.purchase == flinkWorkspacePurchaseManaged &&
		oldTuple.visibility == flinkWorkspaceIdentityStable && newTuple.visibility == flinkWorkspaceIdentityStable &&
		oldTuple.token == newTuple.token && oldTuple.fingerprint == newTuple.fingerprint &&
		((oldTuple.capacity == flinkWorkspaceCapacityLegacy && newTuple.capacity == flinkWorkspaceCapacityLegacy) ||
			(oldTuple.capacity == flinkWorkspaceCapacityInitialAdoptionPending && newTuple.capacity == flinkWorkspaceCapacityInitial)) {
		return nil
	}
	if oldTuple.purchase == flinkWorkspacePurchaseImportedUnknown && newTuple.purchase == flinkWorkspacePurchaseImportedUnknown &&
		oldTuple.capacity == flinkWorkspaceCapacityInitialAdoptionPending && newTuple.capacity == flinkWorkspaceCapacityInitial &&
		oldTuple.visibility == flinkWorkspaceIdentityStable && newTuple.visibility == flinkWorkspaceIdentityStable &&
		oldTuple.token == flinkWorkspaceProtocolUnavailable && newTuple.token == flinkWorkspaceProtocolUnavailable &&
		oldTuple.fingerprint == flinkWorkspaceProtocolUnavailable && newTuple.fingerprint == flinkWorkspaceProtocolUnavailable {
		return nil
	}
	return flinkWorkspaceProtocolError("resource %q attempted unsupported transition %s", id, strings.Join([]string{
		oldTuple.purchase + "/" + oldTuple.capacity + "/" + oldTuple.visibility,
		newTuple.purchase + "/" + newTuple.capacity + "/" + newTuple.visibility,
	}, " -> "))
}
