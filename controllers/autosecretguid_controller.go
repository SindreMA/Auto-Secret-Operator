package controllers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	autosecretv1alpha1 "github.com/SindreMA/auto-secret-operator/api/v1alpha1"
)

// AutoSecretGuidReconciler reconciles an AutoSecretGuid object
type AutoSecretGuidReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=auto-secret.io,resources=autosecretguids,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=auto-secret.io,resources=autosecretguids/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=auto-secret.io,resources=autosecretguids/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete

// Reconcile handles AutoSecretGuid resources
func (r *AutoSecretGuidReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Fetch the AutoSecretGuid instance
	var autoSecretGuid autosecretv1alpha1.AutoSecretGuid
	if err := r.Get(ctx, req.NamespacedName, &autoSecretGuid); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Check if being deleted
	if !autoSecretGuid.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	// Determine secret name
	secretName := autoSecretGuid.Spec.SecretName
	if secretName == "" {
		secretName = autoSecretGuid.Name
	}

	// Reconcile secret
	guid, err := r.reconcileSecret(ctx, &autoSecretGuid, secretName)
	if err != nil {
		log.Error(err, "Failed to reconcile secret")
		return ctrl.Result{}, err
	}

	// Update status
	autoSecretGuid.Status.SecretName = secretName
	autoSecretGuid.Status.GUID = guid
	if err := r.Status().Update(ctx, &autoSecretGuid); err != nil {
		log.Error(err, "Failed to update AutoSecretGuid status")
		return ctrl.Result{}, err
	}

	log.Info("Successfully reconciled AutoSecretGuid",
		"name", autoSecretGuid.Name,
		"secret", secretName,
		"guid", guid)

	return ctrl.Result{}, nil
}

func (r *AutoSecretGuidReconciler) reconcileSecret(ctx context.Context, autoSecretGuid *autosecretv1alpha1.AutoSecretGuid, secretName string) (string, error) {
	log := log.FromContext(ctx)

	// Check if secret already exists
	var existingSecret corev1.Secret
	err := r.Get(ctx, client.ObjectKey{Name: secretName, Namespace: autoSecretGuid.Namespace}, &existingSecret)

	var guid string

	if err == nil {
		// Secret exists, check if guid is already set
		if existingGuid, hasGuid := existingSecret.Data["guid"]; hasGuid {
			log.Info("Secret already exists with guid", "name", secretName)
			// Still update labels and annotations
			if existingSecret.Labels == nil {
				existingSecret.Labels = make(map[string]string)
			}
			for k, v := range autoSecretGuid.Labels {
				existingSecret.Labels[k] = v
			}
			if existingSecret.Annotations == nil {
				existingSecret.Annotations = make(map[string]string)
			}
			for k, v := range autoSecretGuid.Annotations {
				existingSecret.Annotations[k] = v
			}
			if err := r.Update(ctx, &existingSecret); err != nil {
				return "", fmt.Errorf("failed to update secret metadata: %w", err)
			}
			return string(existingGuid), nil
		}
		// Secret exists but no guid, generate one
		guid, err = r.generateGUID(autoSecretGuid)
		if err != nil {
			return "", fmt.Errorf("failed to generate guid: %w", err)
		}
		existingSecret.Data = map[string][]byte{
			"guid": []byte(guid),
		}
		// Copy labels and annotations from AutoSecretGuid to Secret
		if existingSecret.Labels == nil {
			existingSecret.Labels = make(map[string]string)
		}
		for k, v := range autoSecretGuid.Labels {
			existingSecret.Labels[k] = v
		}
		if existingSecret.Annotations == nil {
			existingSecret.Annotations = make(map[string]string)
		}
		for k, v := range autoSecretGuid.Annotations {
			existingSecret.Annotations[k] = v
		}
		if err := r.Update(ctx, &existingSecret); err != nil {
			return "", fmt.Errorf("failed to update secret: %w", err)
		}
		log.Info("Updated secret with guid", "name", secretName)
		return guid, nil
	}

	if !apierrors.IsNotFound(err) {
		return "", err
	}

	// Secret doesn't exist, create it
	guid, err = r.generateGUID(autoSecretGuid)
	if err != nil {
		return "", fmt.Errorf("failed to generate guid: %w", err)
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        secretName,
			Namespace:   autoSecretGuid.Namespace,
			Labels:      make(map[string]string),
			Annotations: make(map[string]string),
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"guid": []byte(guid),
		},
	}

	// Copy labels and annotations from AutoSecretGuid to Secret
	for k, v := range autoSecretGuid.Labels {
		secret.Labels[k] = v
	}
	for k, v := range autoSecretGuid.Annotations {
		secret.Annotations[k] = v
	}

	// Set owner reference
	if err := controllerutil.SetControllerReference(autoSecretGuid, secret, r.Scheme); err != nil {
		return "", err
	}

	if err := r.Create(ctx, secret); err != nil {
		return "", fmt.Errorf("failed to create secret: %w", err)
	}
	log.Info("Created secret", "name", secretName)

	return guid, nil
}

func (r *AutoSecretGuidReconciler) generateGUID(autoSecretGuid *autosecretv1alpha1.AutoSecretGuid) (string, error) {
	format := autoSecretGuid.Spec.Format
	if format == "" {
		format = "uuidv4"
	}

	switch format {
	case "uuidv4":
		return generateUUIDv4()
	case "uuidv7":
		return generateUUIDv7()
	case "short-uuid":
		return generateShortUUID()
	default:
		return "", fmt.Errorf("unsupported format: %s", format)
	}
}

// generateUUIDv4 generates a random UUID v4
func generateUUIDv4() (string, error) {
	uuid := make([]byte, 16)
	if _, err := rand.Read(uuid); err != nil {
		return "", err
	}

	// Set version (4) and variant bits
	uuid[6] = (uuid[6] & 0x0f) | 0x40 // Version 4
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // Variant RFC4122

	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:16]), nil
}

// generateUUIDv7 generates a time-ordered UUID v7
func generateUUIDv7() (string, error) {
	uuid := make([]byte, 16)

	// Get current timestamp in milliseconds
	now := time.Now().UnixMilli()

	// First 48 bits: timestamp
	uuid[0] = byte(now >> 40)
	uuid[1] = byte(now >> 32)
	uuid[2] = byte(now >> 24)
	uuid[3] = byte(now >> 16)
	uuid[4] = byte(now >> 8)
	uuid[5] = byte(now)

	// Remaining bits: random
	if _, err := rand.Read(uuid[6:]); err != nil {
		return "", err
	}

	// Set version (7) and variant bits
	uuid[6] = (uuid[6] & 0x0f) | 0x70 // Version 7
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // Variant RFC4122

	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:16]), nil
}

// generateShortUUID generates a short base64-encoded UUID
func generateShortUUID() (string, error) {
	uuid := make([]byte, 16)
	if _, err := rand.Read(uuid); err != nil {
		return "", err
	}

	// URL-safe base64 encoding without padding
	return base64.RawURLEncoding.EncodeToString(uuid), nil
}

// SetupWithManager sets up the controller with the Manager
func (r *AutoSecretGuidReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&autosecretv1alpha1.AutoSecretGuid{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}
