package latencyprofilecontroller

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	apiconfigv1 "github.com/openshift/api/config/v1"
	configv1 "github.com/openshift/client-go/config/clientset/versioned/typed/config/v1"
	listerv1 "github.com/openshift/client-go/config/listers/config/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"

	configv1informers "github.com/openshift/client-go/config/informers/externalversions/config/v1"
	"github.com/openshift/library-go/pkg/controller/factory"
	"github.com/openshift/library-go/pkg/operator/events"
	"github.com/openshift/library-go/pkg/operator/v1helpers"

	controlplanev1 "github.com/openshift/api/kubecontrolplane/v1"
	"github.com/openshift/cluster-kube-apiserver-operator/pkg/operator/operatorclient"
)

const (
	apiServerConfigMapName = "config"
	apiServerConfigMapKey  = "config.yaml"

	DefaultNotReadyTolerationSecondsArgument    = "default-not-ready-toleration-seconds"
	DefaultUnreachableTolerationSecondsArgument = "default-unreachable-toleration-seconds"
)

// LatencyProfileController periodically lists the config/v1/node object
// and fetches the worker latency profile applied on the cluster which is used to
// updates the status of the config node object. The status updates reflect the
// state of kube-api-server(s) running on control plane node(s) and their
// observed config for default-not-ready-toleration-seconds and
// default-unreachable-toleration-seconds match the applied arguments.

type latencyProfileController struct {
	operatorClient   v1helpers.StaticPodOperatorClient
	configClient     configv1.ConfigV1Interface
	configMapClient  corev1client.ConfigMapsGetter
	configNodeLister listerv1.NodeLister
}

func NewLatencyProfileController(
	operatorClient v1helpers.StaticPodOperatorClient,
	configClient configv1.ConfigV1Interface,
	nodeInformer configv1informers.NodeInformer,
	kubeInformersForNamespaces v1helpers.KubeInformersForNamespaces,
	kubeClient kubernetes.Interface,
	eventRecorder events.Recorder,
) factory.Controller {

	ret := &latencyProfileController{
		operatorClient:   operatorClient,
		configClient:     configClient,
		configMapClient:  v1helpers.CachedConfigMapGetter(kubeClient.CoreV1(), kubeInformersForNamespaces),
		configNodeLister: nodeInformer.Lister(),
	}

	return factory.New().WithInformers(
		// this is for our general configuration input and our status output in case another actor changes it
		operatorClient.Informer(),

		// We use nodeInformer for observing current worker latency profile
		nodeInformer.Informer(),

		// for configmaps of operator client target namespace
		kubeInformersForNamespaces.InformersFor(operatorclient.TargetNamespace).Core().V1().ConfigMaps().Informer(),
	).ResyncEvery(time.Minute).WithSync(ret.sync).WithSyncDegradedOnError(operatorClient).ToController(
		"LatencyProfileController",
		eventRecorder.WithComponentSuffix("latency-profile-controller"),
	)
}

func (c *latencyProfileController) sync(ctx context.Context, syncCtx factory.SyncContext) error {
	// Collect the current latency profile
	configNodeObj, err := c.configNodeLister.Get("cluster")
	if err != nil && !errors.IsNotFound(err) {
		return err
	} else if errors.IsNotFound(err) {
		// if config/v1/node/cluster object is not found this controller should do nothing
		return nil
	}

	degradedCondition := metav1.Condition{
		Type:   apiconfigv1.KubeAPIServerDegraded,
		Status: metav1.ConditionUnknown,
	}
	progressingCondition := metav1.Condition{
		Type:   apiconfigv1.KubeAPIServerProgressing,
		Status: metav1.ConditionUnknown,
	}
	completedCondition := metav1.Condition{
		Type:   apiconfigv1.KubeAPIServerComplete,
		Status: metav1.ConditionUnknown,
	}
	if configNodeObj.Spec.WorkerLatencyProfile == "" {
		// TODO: in case worker latency profile is transitioned
		// from Default/Medium/Low to ""
		// check that the required apiServerArgs should vanish
		degradedCondition.Status = metav1.ConditionFalse
		progressingCondition.Status = metav1.ConditionFalse
		completedCondition.Status = metav1.ConditionTrue

		degradedCondition.Reason = "AsExpected"
		progressingCondition.Reason = "AsExpected"
		completedCondition.Reason = reasonLatencyProfileEmpty

		completedCondition.Message = "worker latency profile not set on cluster"

		_, _ = c.updateConfigNodeStatus(ctx, degradedCondition, progressingCondition, completedCondition)
		_, err = c.alternateUpdateStatus(ctx, copyConditions(degradedCondition, progressingCondition, completedCondition)...)
		return err
	}

	_, operatorStatus, _, err := c.operatorClient.GetStaticPodOperatorState()
	if err != nil {
		return err
	}

	// Collect the unique set of revisions of the apiserver nodes
	revisionMap := map[int32]struct{}{}
	uniqueRevisions := []int32{}
	for _, nodeStatus := range operatorStatus.NodeStatuses {
		revision := nodeStatus.CurrentRevision
		if _, ok := revisionMap[revision]; !ok {
			revisionMap[revision] = struct{}{}
			uniqueRevisions = append(uniqueRevisions, revision)
		}
	}

	// Collect the current latency profile
	desiredApiServerArgumentVals := map[string]string{}

	switch configNodeObj.Spec.WorkerLatencyProfile {
	case apiconfigv1.DefaultUpdateDefaultReaction:
		desiredApiServerArgumentVals[DefaultNotReadyTolerationSecondsArgument] = strconv.Itoa(apiconfigv1.DefaultNotReadyTolerationSeconds)
		desiredApiServerArgumentVals[DefaultUnreachableTolerationSecondsArgument] = strconv.Itoa(apiconfigv1.DefaultUnreachableTolerationSeconds)
	case apiconfigv1.MediumUpdateAverageReaction:
		desiredApiServerArgumentVals[DefaultNotReadyTolerationSecondsArgument] = strconv.Itoa(apiconfigv1.MediumNotReadyTolerationSeconds)
		desiredApiServerArgumentVals[DefaultUnreachableTolerationSecondsArgument] = strconv.Itoa(apiconfigv1.MediumUnreachableTolerationSeconds)
	case apiconfigv1.LowUpdateSlowReaction:
		desiredApiServerArgumentVals[DefaultNotReadyTolerationSecondsArgument] = strconv.Itoa(apiconfigv1.LowNotReadyTolerationSeconds)
		desiredApiServerArgumentVals[DefaultUnreachableTolerationSecondsArgument] = strconv.Itoa(apiconfigv1.LowUnreachableTolerationSeconds)
	}

	// For each revision, check that the configmap for that revision have
	// correct argument values or not
	revisionsHaveSynced := true
	for _, revision := range uniqueRevisions {
		configMapNameWithRevision := fmt.Sprintf("%s-%d", apiServerConfigMapName, revision)
		configMap, err := c.configMapClient.ConfigMaps(operatorclient.TargetNamespace).Get(ctx, configMapNameWithRevision, metav1.GetOptions{})
		if err != nil {
			return err
		}

		match, err := configMatchApiServerArguments(configMap, desiredApiServerArgumentVals)
		if err != nil {
			return err
		}
		if !match {
			revisionsHaveSynced = false
			break
		}
	}

	if revisionsHaveSynced {
		// APIServer has Completed rollout
		completedCondition.Status = metav1.ConditionTrue
		completedCondition.Message = "all kube-apiserver(s) have updated latency profile"
		completedCondition.Reason = reasonLatencyProfileUpdated

		// APIServer is not Progressing rollout
		progressingCondition.Status = metav1.ConditionFalse
		progressingCondition.Reason = reasonLatencyProfileUpdated

		// APIServer is not Degraded
		degradedCondition.Status = metav1.ConditionFalse
		degradedCondition.Reason = reasonLatencyProfileUpdated
	} else {
		// APIServer has not Completed rollout
		completedCondition.Status = metav1.ConditionFalse
		completedCondition.Reason = reasonLatencyProfileUpdateTriggered

		// APIServer is Progressing rollout
		progressingCondition.Status = metav1.ConditionTrue
		progressingCondition.Message = "kube-apiserver(s) are updating latency profile"
		progressingCondition.Reason = reasonLatencyProfileUpdateTriggered

		// APIServer is not Degraded
		degradedCondition.Status = metav1.ConditionFalse
		degradedCondition.Reason = reasonLatencyProfileUpdateTriggered
	}
	_, _ = c.updateConfigNodeStatus(ctx, degradedCondition, progressingCondition, completedCondition)
	_, err = c.alternateUpdateStatus(ctx, copyConditions(degradedCondition, progressingCondition, completedCondition)...)
	return err
}

// configMatchApiServerArguments checks if the specified config map containing kube-apiserver node
// config contains all the specified arguments and values in observedconfig.apiServerArguments field
func configMatchApiServerArguments(configMap *corev1.ConfigMap, argValMap map[string]string) (bool, error) {
	configData, ok := configMap.Data[apiServerConfigMapKey]
	if !ok {
		return false, fmt.Errorf("could not find %s in %s config map from %s namespace", apiServerConfigMapKey, configMap.Name, configMap.Namespace)
	}
	var kubeApiServerConfig controlplanev1.KubeAPIServerConfig
	if err := json.Unmarshal([]byte(configData), &kubeApiServerConfig); err != nil {
		return false, err
	}
	for arg := range argValMap {
		expectedValue := argValMap[arg]
		apiServerArgumentFetchedValues, ok := kubeApiServerConfig.APIServerArguments[arg]

		// such an argument does not exist in config
		if !ok {
			return false, nil
		}
		if len(apiServerArgumentFetchedValues) > 0 {
			if !(apiServerArgumentFetchedValues[0] == expectedValue) {
				return false, nil
			}
		}
	}
	return true, nil
}
