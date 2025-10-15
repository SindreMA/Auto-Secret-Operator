package controllers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"

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

// AutoSecretBasicReconciler reconciles an AutoSecretBasic object
type AutoSecretBasicReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=auto-secret.io,resources=autosecretbasics,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=auto-secret.io,resources=autosecretbasics/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=auto-secret.io,resources=autosecretbasics/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete

// Reconcile handles AutoSecretBasic resources
func (r *AutoSecretBasicReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Fetch the AutoSecretBasic instance
	var autoSecretBasic autosecretv1alpha1.AutoSecretBasic
	if err := r.Get(ctx, req.NamespacedName, &autoSecretBasic); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Check if being deleted
	if !autoSecretBasic.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	// Determine secret name
	secretName := autoSecretBasic.Spec.SecretName
	if secretName == "" {
		secretName = autoSecretBasic.Name
	}

	// Reconcile secret
	if err := r.reconcileSecret(ctx, &autoSecretBasic, secretName); err != nil {
		log.Error(err, "Failed to reconcile secret")
		return ctrl.Result{}, err
	}

	// Update status
	autoSecretBasic.Status.SecretName = secretName
	if err := r.Status().Update(ctx, &autoSecretBasic); err != nil {
		log.Error(err, "Failed to update AutoSecretBasic status")
		return ctrl.Result{}, err
	}

	log.Info("Successfully reconciled AutoSecretBasic",
		"name", autoSecretBasic.Name,
		"secret", secretName)

	return ctrl.Result{}, nil
}

func (r *AutoSecretBasicReconciler) reconcileSecret(ctx context.Context, autoSecretBasic *autosecretv1alpha1.AutoSecretBasic, secretName string) error {
	log := log.FromContext(ctx)

	// Check if secret already exists
	var existingSecret corev1.Secret
	err := r.Get(ctx, client.ObjectKey{Name: secretName, Namespace: autoSecretBasic.Namespace}, &existingSecret)

	if err == nil {
		// Secret exists, check if password is already set
		if _, hasPassword := existingSecret.Data["password"]; hasPassword {
			log.Info("Secret already exists with password", "name", secretName)
			// Still update labels and annotations
			if existingSecret.Labels == nil {
				existingSecret.Labels = make(map[string]string)
			}
			for k, v := range autoSecretBasic.Labels {
				existingSecret.Labels[k] = v
			}
			if existingSecret.Annotations == nil {
				existingSecret.Annotations = make(map[string]string)
			}
			for k, v := range autoSecretBasic.Annotations {
				existingSecret.Annotations[k] = v
			}
			if err := r.Update(ctx, &existingSecret); err != nil {
				return fmt.Errorf("failed to update secret metadata: %w", err)
			}
			return nil
		}
		// Secret exists but no password, update it
		password, err := r.generatePassword(autoSecretBasic)
		if err != nil {
			return fmt.Errorf("failed to generate password: %w", err)
		}
		existingSecret.Data = map[string][]byte{
			"username": []byte(autoSecretBasic.Spec.Username),
			"password": []byte(password),
		}
		// Copy labels and annotations from AutoSecretBasic to Secret
		if existingSecret.Labels == nil {
			existingSecret.Labels = make(map[string]string)
		}
		for k, v := range autoSecretBasic.Labels {
			existingSecret.Labels[k] = v
		}
		if existingSecret.Annotations == nil {
			existingSecret.Annotations = make(map[string]string)
		}
		for k, v := range autoSecretBasic.Annotations {
			existingSecret.Annotations[k] = v
		}
		if err := r.Update(ctx, &existingSecret); err != nil {
			return fmt.Errorf("failed to update secret: %w", err)
		}
		log.Info("Updated secret with password", "name", secretName)
		return nil
	}

	if !apierrors.IsNotFound(err) {
		return err
	}

	// Secret doesn't exist, create it
	password, err := r.generatePassword(autoSecretBasic)
	if err != nil {
		return fmt.Errorf("failed to generate password: %w", err)
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        secretName,
			Namespace:   autoSecretBasic.Namespace,
			Labels:      make(map[string]string),
			Annotations: make(map[string]string),
		},
		Type: corev1.SecretTypeBasicAuth,
		Data: map[string][]byte{
			"username": []byte(autoSecretBasic.Spec.Username),
			"password": []byte(password),
		},
	}

	// Copy labels and annotations from AutoSecretBasic to Secret
	for k, v := range autoSecretBasic.Labels {
		secret.Labels[k] = v
	}
	for k, v := range autoSecretBasic.Annotations {
		secret.Annotations[k] = v
	}

	// Set owner reference
	if err := controllerutil.SetControllerReference(autoSecretBasic, secret, r.Scheme); err != nil {
		return err
	}

	if err := r.Create(ctx, secret); err != nil {
		return fmt.Errorf("failed to create secret: %w", err)
	}
	log.Info("Created secret", "name", secretName)

	return nil
}

func (r *AutoSecretBasicReconciler) generatePassword(autoSecretBasic *autosecretv1alpha1.AutoSecretBasic) (string, error) {
	length := autoSecretBasic.Spec.PasswordLength
	if length == 0 {
		length = 30
	}

	charset := autoSecretBasic.Spec.PasswordCharset
	if charset == "" {
		charset = "hex"
	}

	switch charset {
	case "alphanumeric":
		return generateAlphanumericPassword(int(length))
	case "ascii-printable":
		return generateASCIIPrintablePassword(int(length))
	case "hex":
		return generateHexPassword(int(length))
	case "base64":
		return generateBase64Password(int(length))
	default:
		return "", fmt.Errorf("unsupported charset: %s", charset)
	}
}

func generateAlphanumericPassword(n int) (string, error) {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		b[i] = letters[num.Int64()]
	}
	return string(b), nil
}

func generateASCIIPrintablePassword(n int) (string, error) {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()-_=+[]{}|;:,.<>?/"
	b := make([]byte, n)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			return "", err
		}
		b[i] = chars[num.Int64()]
	}
	return string(b), nil
}

func generateHexPassword(n int) (string, error) {
	bytes := make([]byte, (n+1)/2)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes)[:n], nil
}

func generateBase64Password(n int) (string, error) {
	bytes := make([]byte, (n*3+3)/4)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:n], nil
}

// SetupWithManager sets up the controller with the Manager
func (r *AutoSecretBasicReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&autosecretv1alpha1.AutoSecretBasic{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}
