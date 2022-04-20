package node

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"

	configv1 "github.com/openshift/api/config/v1"
	configlistersv1 "github.com/openshift/client-go/config/listers/config/v1"
	"github.com/openshift/cluster-kube-apiserver-operator/pkg/operator/configobservation"
	"github.com/openshift/library-go/pkg/operator/configobserver"
	"github.com/openshift/library-go/pkg/operator/events"
)

type workerLatencyProfileTestScenario struct {
	name                  string
	existingKubeAPIConfig map[string]interface{}
	expectedKubeAPIConfig map[string]interface{}
	workerLatencyProfile  configv1.WorkerLatencyProfileType
}

func multiScenarioLatencyProfilesTest(
	configObserveFn func(configobserver.Listers, events.Recorder, map[string]interface{}) (ret map[string]interface{}, errs []error),
	scenarios []workerLatencyProfileTestScenario,
) func(*testing.T) {
	return func(t *testing.T) {
		for _, scenario := range scenarios {
			t.Run(scenario.name, func(t *testing.T) {
				// test data
				eventRecorder := events.NewInMemoryRecorder("")
				configNodeIndexer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{})
				configNodeIndexer.Add(&configv1.Node{
					ObjectMeta: metav1.ObjectMeta{Name: "cluster"},
					Spec:       configv1.NodeSpec{WorkerLatencyProfile: scenario.workerLatencyProfile},
				})
				listers := configobservation.Listers{
					NodeLister: configlistersv1.NewNodeLister(configNodeIndexer),
				}

				// act
				observedKubeAPIConfig, err := configObserveFn(listers, eventRecorder, scenario.existingKubeAPIConfig)

				// validate
				if len(err) > 0 {
					t.Fatal(err)
				}
				if !cmp.Equal(scenario.expectedKubeAPIConfig, observedKubeAPIConfig) {
					t.Fatalf("unexpected configuration, diff = %v", cmp.Diff(scenario.expectedKubeAPIConfig, observedKubeAPIConfig))
				}
			})
		}
	}
}

func TestObserveNotReadyTolerationSeconds(t *testing.T) {
	scenarios := []workerLatencyProfileTestScenario{
		// scenario 1: empty worker latency profile
		{
			name:                  "default value is not applied when worker latency profile is unset",
			expectedKubeAPIConfig: nil,
			workerLatencyProfile:  "", // empty worker latency profile
		},

		// scenario 2: Default
		{
			name: "worker latency profile Default: config with default-not-ready-toleration-seconds=300",
			expectedKubeAPIConfig: map[string]interface{}{
				"apiServerArguments": map[string]interface{}{
					"default-not-ready-toleration-seconds": []interface{}{"300"},
				},
			},
			workerLatencyProfile: configv1.DefaultUpdateDefaultReaction,
		},

		// scenario 3: MediumUpdateAverageReaction
		{
			name: "worker latency profile MediumUpdateAverageReaction: config with default-not-ready-toleration-seconds=60",
			expectedKubeAPIConfig: map[string]interface{}{
				"apiServerArguments": map[string]interface{}{
					"default-not-ready-toleration-seconds": []interface{}{"60"},
				},
			},
			workerLatencyProfile: configv1.MediumUpdateAverageReaction,
		},

		// scenario 4: LowUpdateSlowReaction
		{
			name: "worker latency profile LowUpdateSlowReaction: config with default-not-ready-toleration-seconds=60",
			expectedKubeAPIConfig: map[string]interface{}{
				"apiServerArguments": map[string]interface{}{
					"default-not-ready-toleration-seconds": []interface{}{"60"},
				},
			},
			workerLatencyProfile: configv1.LowUpdateSlowReaction,
		},
	}
	multiScenarioLatencyProfilesTest(ObserveNotReadyTolerationSeconds, scenarios)(t)
}

func TestObserveUnreachableTolerationSeconds(t *testing.T) {
	scenarios := []workerLatencyProfileTestScenario{
		// scenario 1: empty worker latency profile
		{
			name:                  "default value is not applied when worker latency profile is unset",
			expectedKubeAPIConfig: nil,
			workerLatencyProfile:  "", // empty worker latency profile
		},

		// scenario 2: Default
		{
			name: "worker latency profile Default: config with default-unreachable-toleration-seconds=300",
			expectedKubeAPIConfig: map[string]interface{}{
				"apiServerArguments": map[string]interface{}{
					"default-unreachable-toleration-seconds": []interface{}{"300"},
				},
			},
			workerLatencyProfile: configv1.DefaultUpdateDefaultReaction,
		},

		// scenario 3: MediumUpdateAverageReaction
		{
			name: "worker latency profile MediumUpdateAverageReaction: config with default-unreachable-toleration-seconds=60",
			expectedKubeAPIConfig: map[string]interface{}{
				"apiServerArguments": map[string]interface{}{
					"default-unreachable-toleration-seconds": []interface{}{"60"},
				},
			},
			workerLatencyProfile: configv1.MediumUpdateAverageReaction,
		},

		// scenario 4: LowUpdateSlowReaction
		{
			name: "worker latency profile LowUpdateSlowReaction: config with default-unreachable-toleration-seconds=60",
			expectedKubeAPIConfig: map[string]interface{}{
				"apiServerArguments": map[string]interface{}{
					"default-unreachable-toleration-seconds": []interface{}{"60"},
				},
			},
			workerLatencyProfile: configv1.LowUpdateSlowReaction,
		},
	}
	multiScenarioLatencyProfilesTest(ObserveUnreachableTolerationSeconds, scenarios)(t)
}
