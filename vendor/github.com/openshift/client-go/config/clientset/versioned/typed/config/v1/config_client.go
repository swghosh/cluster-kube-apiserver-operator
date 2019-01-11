// Code generated by client-gen. DO NOT EDIT.

package v1

import (
	v1 "github.com/openshift/api/config/v1"
	"github.com/openshift/client-go/config/clientset/versioned/scheme"
	serializer "k8s.io/apimachinery/pkg/runtime/serializer"
	rest "k8s.io/client-go/rest"
)

type ConfigV1Interface interface {
	RESTClient() rest.Interface
	AuthenticationsGetter
	BuildsGetter
	ClusterOperatorsGetter
	ClusterVersionsGetter
	ConsolesGetter
	DNSsGetter
	ImagesGetter
	InfrastructuresGetter
	IngressesGetter
	NetworksGetter
	OAuthsGetter
	ProjectsGetter
	ProxiesGetter
	SchedulingsGetter
}

// ConfigV1Client is used to interact with features provided by the config.openshift.io group.
type ConfigV1Client struct {
	restClient rest.Interface
}

func (c *ConfigV1Client) Authentications() AuthenticationInterface {
	return newAuthentications(c)
}

func (c *ConfigV1Client) Builds() BuildInterface {
	return newBuilds(c)
}

func (c *ConfigV1Client) ClusterOperators() ClusterOperatorInterface {
	return newClusterOperators(c)
}

func (c *ConfigV1Client) ClusterVersions() ClusterVersionInterface {
	return newClusterVersions(c)
}

func (c *ConfigV1Client) Consoles() ConsoleInterface {
	return newConsoles(c)
}

func (c *ConfigV1Client) DNSs() DNSInterface {
	return newDNSs(c)
}

func (c *ConfigV1Client) Images() ImageInterface {
	return newImages(c)
}

func (c *ConfigV1Client) Infrastructures() InfrastructureInterface {
	return newInfrastructures(c)
}

func (c *ConfigV1Client) Ingresses() IngressInterface {
	return newIngresses(c)
}

func (c *ConfigV1Client) Networks() NetworkInterface {
	return newNetworks(c)
}

func (c *ConfigV1Client) OAuths() OAuthInterface {
	return newOAuths(c)
}

func (c *ConfigV1Client) Projects() ProjectInterface {
	return newProjects(c)
}

func (c *ConfigV1Client) Proxies() ProxyInterface {
	return newProxies(c)
}

func (c *ConfigV1Client) Schedulings() SchedulingInterface {
	return newSchedulings(c)
}

// NewForConfig creates a new ConfigV1Client for the given config.
func NewForConfig(c *rest.Config) (*ConfigV1Client, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	return &ConfigV1Client{client}, nil
}

// NewForConfigOrDie creates a new ConfigV1Client for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *ConfigV1Client {
	client, err := NewForConfig(c)
	if err != nil {
		panic(err)
	}
	return client
}

// New creates a new ConfigV1Client for the given RESTClient.
func New(c rest.Interface) *ConfigV1Client {
	return &ConfigV1Client{c}
}

func setConfigDefaults(config *rest.Config) error {
	gv := v1.SchemeGroupVersion
	config.GroupVersion = &gv
	config.APIPath = "/apis"
	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: scheme.Codecs}

	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	return nil
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *ConfigV1Client) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}
