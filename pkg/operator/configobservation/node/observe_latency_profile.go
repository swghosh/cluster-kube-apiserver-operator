package node

import (
	"fmt"
	"strconv"

	"github.com/openshift/cluster-kube-apiserver-operator/pkg/operator/configobservation"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	configV1 "github.com/openshift/api/config/v1"
	"github.com/openshift/library-go/pkg/operator/configobserver"
	"github.com/openshift/library-go/pkg/operator/events"
)

var notReadyTolerationSecondsPath = []string{"apiServerArguments", "default-not-ready-toleration-seconds"}
var unreachableTolerationSecondsPath = []string{"apiServerArguments", "default-unreachable-toleration-seconds"}

// ObserveNotReadyTolerationSeconds observes the value that should be set for default-not-ready-toleration-seconds
// api server argument on the basis of provided worker latency profile from config node object.
func ObserveNotReadyTolerationSeconds(genericListers configobserver.Listers, _ events.Recorder, existingConfig map[string]interface{}) (ret map[string]interface{}, errs []error) {
	defer func() {
		// Prune the observed config so that it only contains default-not-ready-toleration-seconds field.
		ret = configobserver.Pruned(ret, notReadyTolerationSecondsPath)
	}()

	nodeLister := genericListers.(configobservation.Listers).NodeLister
	node, err := nodeLister.Get("cluster")
	if err != nil && !apierrors.IsNotFound(err) {
		// we got an error so without the node object we are not able to determine worker latency profile
		return existingConfig, append(errs, err)
	} else if apierrors.IsNotFound(err) {
		// if config/v1/node/cluster object is not found, that can be treated as a non-error case
		return existingConfig, errs
	}

	// read the observed value
	var observedDefaultNotReadyTolerationSeconds string
	switch node.Spec.WorkerLatencyProfile {
	case configV1.DefaultUpdateDefaultReaction:
		observedDefaultNotReadyTolerationSeconds = strconv.Itoa(configV1.DefaultNotReadyTolerationSeconds)
	case configV1.MediumUpdateAverageReaction:
		observedDefaultNotReadyTolerationSeconds = strconv.Itoa(configV1.MediumNotReadyTolerationSeconds)
	case configV1.LowUpdateSlowReaction:
		observedDefaultNotReadyTolerationSeconds = strconv.Itoa(configV1.LowNotReadyTolerationSeconds)
	// in case of empty worker latency profile, do not update config
	case "":
		return existingConfig, errs
	default:
		return existingConfig, append(errs, fmt.Errorf("unknown worker latency profile found: %v", node.Spec.WorkerLatencyProfile))
	}

	// publish the observed value for default-not-ready-toleration-seconds
	observedConfig := map[string]interface{}{}
	if err = unstructured.SetNestedStringSlice(observedConfig,
		[]string{observedDefaultNotReadyTolerationSeconds},
		notReadyTolerationSecondsPath...); err != nil {
		return existingConfig, append(errs, err)
	}
	return observedConfig, errs
}

// ObserveUnreachableTolerationSeconds observes the value that should be set for default-unreachable-toleration-seconds
// api server argument on the basis of provided worker latency profile from config node object.
func ObserveUnreachableTolerationSeconds(genericListers configobserver.Listers, _ events.Recorder, existingConfig map[string]interface{}) (ret map[string]interface{}, errs []error) {
	defer func() {
		// Prune the observed config so that it only contains default-not-ready-toleration-seconds field.
		ret = configobserver.Pruned(ret, unreachableTolerationSecondsPath)
	}()

	nodeLister := genericListers.(configobservation.Listers).NodeLister
	node, err := nodeLister.Get("cluster")
	if err != nil && !apierrors.IsNotFound(err) {
		// we got an error so without the node object we are not able to determine worker latency profile
		return existingConfig, append(errs, err)
	} else if apierrors.IsNotFound(err) {
		// if config/v1/node/cluster object is not found, that can be treated as a non-error case
		return existingConfig, errs
	}

	// read the observed value
	var observedDefaultUnreachableTolerationSeconds string
	switch node.Spec.WorkerLatencyProfile {
	case configV1.DefaultUpdateDefaultReaction:
		observedDefaultUnreachableTolerationSeconds = strconv.Itoa(configV1.DefaultUnreachableTolerationSeconds)
	case configV1.MediumUpdateAverageReaction:
		observedDefaultUnreachableTolerationSeconds = strconv.Itoa(configV1.MediumUnreachableTolerationSeconds)
	case configV1.LowUpdateSlowReaction:
		observedDefaultUnreachableTolerationSeconds = strconv.Itoa(configV1.LowUnreachableTolerationSeconds)
	// in case of empty workerlatencyprofile, do not update config
	case "":
		return existingConfig, errs
	default:
		return existingConfig, append(errs, fmt.Errorf("unknown worker latency profile found: %v", node.Spec.WorkerLatencyProfile))
	}

	// publish the observed value for default-unreachable-toleration-seconds
	observedConfig := map[string]interface{}{}
	if err = unstructured.SetNestedStringSlice(observedConfig,
		[]string{observedDefaultUnreachableTolerationSeconds},
		unreachableTolerationSecondsPath...); err != nil {
		return existingConfig, append(errs, err)
	}
	return observedConfig, errs
}
