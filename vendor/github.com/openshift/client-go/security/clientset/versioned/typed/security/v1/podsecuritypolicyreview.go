// Code generated by client-gen. DO NOT EDIT.

package v1

import (
	"context"

	v1 "github.com/openshift/api/security/v1"
	scheme "github.com/openshift/client-go/security/clientset/versioned/scheme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gentype "k8s.io/client-go/gentype"
)

// PodSecurityPolicyReviewsGetter has a method to return a PodSecurityPolicyReviewInterface.
// A group's client should implement this interface.
type PodSecurityPolicyReviewsGetter interface {
	PodSecurityPolicyReviews(namespace string) PodSecurityPolicyReviewInterface
}

// PodSecurityPolicyReviewInterface has methods to work with PodSecurityPolicyReview resources.
type PodSecurityPolicyReviewInterface interface {
	Create(ctx context.Context, podSecurityPolicyReview *v1.PodSecurityPolicyReview, opts metav1.CreateOptions) (*v1.PodSecurityPolicyReview, error)
	PodSecurityPolicyReviewExpansion
}

// podSecurityPolicyReviews implements PodSecurityPolicyReviewInterface
type podSecurityPolicyReviews struct {
	*gentype.Client[*v1.PodSecurityPolicyReview]
}

// newPodSecurityPolicyReviews returns a PodSecurityPolicyReviews
func newPodSecurityPolicyReviews(c *SecurityV1Client, namespace string) *podSecurityPolicyReviews {
	return &podSecurityPolicyReviews{
		gentype.NewClient[*v1.PodSecurityPolicyReview](
			"podsecuritypolicyreviews",
			c.RESTClient(),
			scheme.ParameterCodec,
			namespace,
			func() *v1.PodSecurityPolicyReview { return &v1.PodSecurityPolicyReview{} }),
	}
}