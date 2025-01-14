package latencyprofilecontroller

import (
	"encoding/json"
	"fmt"
	"reflect"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	listersv1 "k8s.io/client-go/listers/core/v1"

	configv1 "github.com/openshift/api/config/v1"
	"github.com/openshift/library-go/pkg/operator/configobserver"
	nodeobserver "github.com/openshift/library-go/pkg/operator/configobserver/node"
)

const (
	// static pod revision config maps used by installer controller to track
	// configs across different revisions
	revisionConfigMapName = "config"
	revisionConfigMapKey  = "config.yaml"
)

type revisionConfigMatcher struct {
	configMapLister listersv1.ConfigMapNamespaceLister
	latencyConfigs  []nodeobserver.LatencyConfigProfileTuple
}

// NewInstallerRevisionConfigMatcher is used to create a MatchProfileRevisionConfigsFunc that can be matches
// config maps generated by installer controller for various static pod revisions and match if each of
// the active revisions specified contain arg val pairs specific to given latency profile or not.
func NewInstallerRevisionConfigMatcher(
	configMapLister listersv1.ConfigMapNamespaceLister,
	latencyConfigs []nodeobserver.LatencyConfigProfileTuple,
) MatchProfileRevisionConfigsFunc {

	ret := revisionConfigMatcher{
		configMapLister: configMapLister,
		latencyConfigs:  latencyConfigs,
	}
	return ret.matchProfileForActiveRevisions
}

func (r *revisionConfigMatcher) matchProfileForActiveRevisions(profile configv1.WorkerLatencyProfileType, activeRevisions []int32) (match bool, err error) {
	// For each revision, check that the configmap for that revision have correct arg val pairs or not
	for _, revision := range activeRevisions {
		configMapNameWithRevision := fmt.Sprintf("%s-%d", revisionConfigMapName, revision)
		configMap, err := r.configMapLister.Get(configMapNameWithRevision)
		if err != nil {
			return false, err
		}

		match, err := r.configMatchProfileArguments(configMap, profile)
		if err != nil {
			return false, err
		}
		// in case a single revision doesn't match return false
		if !match {
			return false, nil
		}
	}
	return true, nil
}

func (r *revisionConfigMatcher) configMatchProfileArguments(
	configMap *corev1.ConfigMap,
	currentProfile configv1.WorkerLatencyProfileType,
) (bool, error) {

	configData, ok := configMap.Data[revisionConfigMapKey]
	if !ok {
		return false, fmt.Errorf("could not find %s in %s config map from %s namespace", revisionConfigMapKey, configMap.Name, configMap.Namespace)
	}

	var currentConfig map[string]interface{}
	if err := json.Unmarshal([]byte(configData), &currentConfig); err != nil {
		return false, err
	}

	// set the desiredConfig with expected values
	// also, use the same loop to get a list of configPaths that could be used for pruning
	desiredConfig := make(map[string]interface{})
	usedConfigPaths := make([][]string, len(r.latencyConfigs))

	for i, latencyConfig := range r.latencyConfigs {
		profileValue, ok := latencyConfig.ProfileConfigValues[currentProfile]
		// in case an unknown latency profile is encountered
		// Note: In case new latency profiles are added in the future in openshift/api
		// this could break cluster upgrades and set the controller invoking this function into
		// degraded state.
		if !ok {
			err := fmt.Errorf("unknown worker latency profile found, Profile = %q", currentProfile)
			return false, err
		}
		// set each arg val pair on the desiredConfig
		err := unstructured.SetNestedStringSlice(desiredConfig, []string{profileValue}, latencyConfig.ConfigPath...)
		if err != nil {
			return false, err
		}

		usedConfigPaths[i] = latencyConfig.ConfigPath
	}

	// prune currently observed config to get an object that only has arg val pairs that we like to monitor
	currentConfigPruned := configobserver.Pruned(currentConfig, usedConfigPaths...)

	// desired and current config does match
	if reflect.DeepEqual(currentConfigPruned, desiredConfig) {
		return true, nil
	}

	// desired and current config does not match
	return false, nil
}
