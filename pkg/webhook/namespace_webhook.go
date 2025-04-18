package webhook

import (
	"context"
	"encoding/json"
	"net/http"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// NamespaceCreatorWebhook handles namespace creation
type NamespaceCreatorWebhook struct {
	decoder *admission.Decoder
}

// NewNamespaceCreatorWebhook returns a new NamespaceCreatorWebhook
func NewNamespaceCreatorWebhook(scheme *runtime.Scheme) *NamespaceCreatorWebhook {
	return &NamespaceCreatorWebhook{
		decoder: admission.NewDecoder(scheme),
	}
}

// Handle implements the admission handler
func (a *NamespaceCreatorWebhook) Handle(
	ctx context.Context,
	req admission.Request,
) admission.Response {
	logger := log.FromContext(ctx).WithName("namespace-creator-webhook")

	// Only process create operations
	if req.Operation != admissionv1.Create {
		return admission.Allowed("Not a creation operation")
	}

	namespace := &corev1.Namespace{}
	err := a.decoder.Decode(req, namespace)
	if err != nil {
		logger.Error(err, "Failed to decode namespace")
		return admission.Errored(http.StatusBadRequest, err)
	}

	// Get username from request
	username := req.UserInfo.Username
	logger.Info("Processing namespace creation", "namespace", namespace.Name, "creator", username)

	// Add creator annotation
	if namespace.Annotations == nil {
		namespace.Annotations = make(map[string]string)
	}
	namespace.Annotations["kubernetes.io/creator"] = username

	marshaledNamespace, err := json.Marshal(namespace)
	if err != nil {
		logger.Error(err, "Failed to marshal namespace")
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledNamespace)
}
