package latencyprofilecontroller

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	controlplanev1 "github.com/openshift/api/kubecontrolplane/v1"

	"github.com/openshift/cluster-kube-apiserver-operator/pkg/operator/operatorclient"
)

func TestConfigMatchesControllerManagerArgument(t *testing.T) {
	createConfigMapFromKASConfig := func(
		config controlplanev1.KubeAPIServerConfig,
		configMapName, configMapNamespace string,
	) (configMap corev1.ConfigMap) {

		configAsJsonBytes, err := json.MarshalIndent(config, "", "")
		require.NoError(t, err)

		configMap = corev1.ConfigMap{
			ObjectMeta: v1.ObjectMeta{Name: configMapName, Namespace: configMapNamespace},
			Data: map[string]string{
				apiServerConfigMapKey: string(configAsJsonBytes),
			},
		}
		return configMap
	}

	kasConfigs := []controlplanev1.KubeAPIServerConfig{
		// config 1
		{
			APIServerArguments: map[string]controlplanev1.Arguments{},
		},
		// config 2
		{
			APIServerArguments: map[string]controlplanev1.Arguments{
				"default-not-ready-toleration-seconds": []string{"60"},
				"default-watch-cache-size":             []string{"100"},
			},
		},
		// config 3
		{
			APIServerArguments: map[string]controlplanev1.Arguments{
				"default-watch-cache-size": []string{"100"},
			},
		},
		// config 4
		{
			APIServerArguments: map[string]controlplanev1.Arguments{
				"default-not-ready-toleration-seconds":   []string{"300"},
				"default-unreachable-toleration-seconds": []string{"300"},
			},
		},
	}
	kasConfigMaps := make([]corev1.ConfigMap, len(kasConfigs))
	for i, kasConfig := range kasConfigs {
		kasConfigMaps[i] = createConfigMapFromKASConfig(
			kasConfig, fmt.Sprintf("%s-%d", apiServerConfigMapName, i),
			operatorclient.TargetNamespace)
	}

	scenarios := []struct {
		name               string
		apiServerConfig    *controlplanev1.KubeAPIServerConfig
		apiServerConfigMap *corev1.ConfigMap
		argVals            map[string]string
		expectedMatch      bool
	}{
		{
			name: "arg=default-unreachable-toleration-seconds should not be found in config with empty extendedArgs",

			// config with empty extendedArgs
			apiServerConfig:    &kasConfigs[0],
			apiServerConfigMap: &kasConfigMaps[0],

			argVals:       map[string]string{"default-unreachable-toleration-seconds": "300"},
			expectedMatch: false,
		},
		{
			name: "arg=default-not-ready-toleration-seconds with value=40s should be found in config",

			// config with extendedArgs{default-not-ready-toleration-seconds=40s,default-watch-cache-size}
			apiServerConfig:    &kasConfigs[1],
			apiServerConfigMap: &kasConfigMaps[1],

			argVals:       map[string]string{"default-not-ready-toleration-seconds": "60"},
			expectedMatch: true,
		},
		{
			name: "arg=default-not-ready-toleration-seconds should not be found in config which does not contain that arg",

			// config with extendedArgs{default-watch-cache-size}
			apiServerConfig:    &kasConfigs[2],
			apiServerConfigMap: &kasConfigMaps[2],

			argVals:       map[string]string{"default-not-ready-toleration-seconds": "300"},
			expectedMatch: false,
		},
		{
			name: "arg=default-not-ready-toleration-seconds with value=40s should not be found in config which contains that arg but different value",

			// config with extendedArgs{default-not-ready-toleration-seconds=2m,default-unreachable-toleration-seconds}
			apiServerConfig:    &kasConfigs[3],
			apiServerConfigMap: &kasConfigMaps[3],

			argVals:       map[string]string{"default-not-ready-toleration-seconds": "40"},
			expectedMatch: false,
		},
	}
	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			// act
			actualMatch, err := configMatchApiServerArguments(scenario.apiServerConfigMap, scenario.argVals)
			if err != nil {
				// in case error is encountered during matching
				t.Fatal(err)
			}
			// validate
			if !(actualMatch == scenario.expectedMatch) {
				containStr := "should contain"
				if !scenario.expectedMatch {
					containStr = "should not contain"
				}
				t.Fatalf("unexpected matching, expected = %v but got %v; api-server-config=%v %s %v",
					scenario.expectedMatch, actualMatch,
					*scenario.apiServerConfig, containStr,
					scenario.argVals)
			}
		})
	}
}
