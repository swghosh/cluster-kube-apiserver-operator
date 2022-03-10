package node

// import (
// 	"fmt"

// 	apierrors "k8s.io/apimachinery/pkg/api/errors"
// 	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

// 	configv1 "github.com/openshift/api/config/v1"
// 	"github.com/openshift/cluster-kube-apiserver-operator/pkg/operator/configobservation"
// 	"github.com/openshift/library-go/pkg/operator/configobserver"
// 	"github.com/openshift/library-go/pkg/operator/events"
// )

// var notReadyTolerationSecondsPath = []string{"apiServerArguments", "default-not-ready-toleration-seconds"}
// var unreachableTolerationSecondsPath = []string{"apiServerArguments", "default-unreachable-toleration-seconds"}

// func ObserveNotReadyTolerationSeconds(genericListers configobserver.Listers, _ events.Recorder, existingConfig map[string]interface{}) (ret map[string]interface{}, errs []error) {
// 	defer func() {
// 		// Prune the observed config so that it only contains shutdown-delay-duration field.
// 		ret = configobserver.Pruned(ret, shutdownDelayDurationPath)
// 	}()

// 	// read the observed value
// 	var observedShutdownDelayDuration string
// 	listers := genericListers.(configobservation.Listers)
// 	infra, err := listers.InfrastructureLister().Get("cluster")
// 	if err != nil && !apierrors.IsNotFound(err) {
// 		// we got an error so without the infrastructure object we are not able to determine the type of platform we are running on
// 		return existingConfig, append(errs, err)
// 	}

// 	// see if the current and the observed value differ
// 	observedConfig := map[string]interface{}{}
// 	if currentShutdownDelayDuration != observedShutdownDelayDuration {
// 		if err = unstructured.SetNestedStringSlice(observedConfig, []string{observedShutdownDelayDuration}, shutdownDelayDurationPath...); err != nil {
// 			return existingConfig, append(errs, err)
// 		}
// 		return observedConfig, errs
// 	}

// 	// nothing has changed return the original configuration
// 	return existingConfig, errs
// }

// func ObserveUnreachableTolerationSeconds(genericListers configobserver.Listers, _ events.Recorder, existingConfig map[string]interface{}) (ret map[string]interface{}, errs []error) {
// 	defer func() {
// 		// Prune the observed config so that it only contains gracefulTerminationDuration field.
// 		ret = configobserver.Pruned(ret, gracefulTerminationDurationPath)
// 	}()

// 	// read the observed value
// 	var observedGracefulTerminationDuration string
// 	listers := genericListers.(configobservation.Listers)
// 	infra, err := listers.InfrastructureLister().Get("cluster")
// 	if err != nil && !apierrors.IsNotFound(err) {
// 		// we got an error so without the infrastructure object we are not able to determine the type of platform we are running on
// 		return existingConfig, append(errs, err)
// 	}

// 	switch {
// 	case infra.Status.ControlPlaneTopology == configv1.SingleReplicaTopologyMode:
// 		// reduce termination duration from 135s (default) to 15s to reach the maximum downtime for SNO:
// 		// - the shutdown-delay-duration is set to 0s because there is no load-balancer, and no fallback apiserver
// 		//   anyway that could benefit from a service network taking out the endpoint gracefully
// 		// - additional 15s is for in-flight requests
// 		observedGracefulTerminationDuration = "15"
// 	case infra.Spec.PlatformSpec.Type == configv1.AWSPlatformType:
// 		// AWS has a known issue: https://bugzilla.redhat.com/show_bug.cgi?id=1943804
// 		// We need to extend the shutdown-delay-duration so that an NLB has a chance to notice and remove unhealthy instance.
// 		// Once the mentioned issue is resolved this code must be removed and default values applied
// 		//
// 		// 194s is calculated as follows:
// 		//   the initial 129s is reserved fo the minimal termination period - the time needed for an LB to take an instance out of rotation
// 		//   additional 60s for finishing all in-flight requests
// 		//   an extra 5s to make sure the potential SIGTERM will be sent after the server terminates itself
// 		observedGracefulTerminationDuration = "194"
// 	default:
// 		// don't override default value
// 		return map[string]interface{}{}, errs
// 	}

// 	// read the current value
// 	currentGracefulTerminationDuration, _, err := unstructured.NestedString(existingConfig, gracefulTerminationDurationPath...)
// 	if err != nil {
// 		errs = append(errs, fmt.Errorf("unable to extract gracefulTerminationDuration from the existing config: %v, path = %v", err, gracefulTerminationDurationPath))
// 		// keep going, we are only interested in the observed value which will overwrite the current configuration anyway
// 	}

// 	// see if the current and the observed value differ
// 	observedConfig := map[string]interface{}{}
// 	if currentGracefulTerminationDuration != observedGracefulTerminationDuration {
// 		if err = unstructured.SetNestedField(observedConfig, observedGracefulTerminationDuration, gracefulTerminationDurationPath...); err != nil {
// 			return existingConfig, append(errs, err)
// 		}
// 		return observedConfig, errs
// 	}

// 	// nothing has changed return the original configuration
// 	return existingConfig, errs
// }
