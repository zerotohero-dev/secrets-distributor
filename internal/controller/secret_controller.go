/*
 */

package controller

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"strings"
)

// SecretReconciler reconciles a Secret object
type SecretReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=secrets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;create

func (r *SecretReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the Secret instance
	secret := &corev1.Secret{}
	if err := r.Get(ctx, req.NamespacedName, secret); err != nil {
		if client.IgnoreNotFound(err) == nil {
			// This means the secret was deleted
			logger.Info("secret was deleted",
				"name", req.Name,
				"namespace", req.Namespace,
			)
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	types := secret.Annotations["secrets-distributor.z2h.dev/types"]
	targetNs := secret.Annotations["secrets-distributor.z2h.dev/target-namespace"]

	if types == "" || targetNs == "" {
		logger.Info("skipping secret - missing required annotations",
			"name", secret.Name,
			"namespace", secret.Namespace,
		)
		return ctrl.Result{}, nil
	}

	// If no k8s in types, skip
	if !strings.Contains(types, "k8s") {
		logger.Info("skipping secret - k8s not in target types",
			"name", secret.Name,
			"types", types,
		)
		return ctrl.Result{}, nil
	}

	// Create namespace if it doesn't exist
	targetNamespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: targetNs,
		},
	}

	if err := r.Create(ctx, targetNamespace); err != nil {
		if !errors.IsAlreadyExists(err) {
			logger.Error(err, "failed to create namespace")
			return ctrl.Result{}, err
		}
	}

	// Prepare target secret
	targetSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secret.Name,
			Namespace: targetNs,
			Labels:    secret.Labels,
			// Copy only relevant annotations
			Annotations: map[string]string{
				"secrets-distributor.z2h.dev/owner": secret.Annotations["secrets-distributor.z2h.dev/owner"],
			},
		},
		Type: secret.Type,
		Data: secret.Data,
	}

	// Try to get existing secret
	existingSecret := &corev1.Secret{}
	err := r.Get(ctx, client.ObjectKey{
		Name:      secret.Name,
		Namespace: targetNs}, existingSecret)

	if err != nil {
		if errors.IsNotFound(err) {
			// Create a new secret
			if err := r.Create(ctx, targetSecret); err != nil {
				logger.Error(err, "failed to create target secret")
				return ctrl.Result{}, err
			}
			logger.Info("created target secret",
				"name", targetSecret.Name,
				"namespace", targetSecret.Namespace,
			)
			return ctrl.Result{}, nil
		}
		logger.Error(err, "failed to get existing secret")
		return ctrl.Result{}, err
	}

	// Check if update is needed
	if reflect.DeepEqual(existingSecret.Data, targetSecret.Data) &&
		reflect.DeepEqual(existingSecret.Labels, targetSecret.Labels) {
		logger.Info("target secret is up to date, skipping update",
			"name", targetSecret.Name,
			"namespace", targetSecret.Namespace,
		)
		return ctrl.Result{}, nil
	}

	// Update existing secret
	if err := r.Update(ctx, targetSecret); err != nil {
		logger.Error(err, "failed to update target secret")
		return ctrl.Result{}, err
	}

	logger.Info("updated target secret",
		"name", targetSecret.Name,
		"namespace", targetSecret.Namespace,
	)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Secret{}).
		Complete(r)
}
