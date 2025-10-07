package controllers

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"net/url"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	autosecretv1alpha1 "github.com/yourusername/db-secret-operator/api/v1alpha1"
)

// randomPassword generates a cryptographically secure random password
func randomPassword(n int) (string, error) {
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

// AutoSecretReconciler reconciles an AutoSecret object
type AutoSecretReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=auto-secret.io,resources=autosecrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=auto-secret.io,resources=autosecrets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=auto-secret.io,resources=autosecrets/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete

// Reconcile handles AutoSecret resources
func (r *AutoSecretReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Fetch the AutoSecret instance
	var autoSecret autosecretv1alpha1.AutoSecret
	if err := r.Get(ctx, req.NamespacedName, &autoSecret); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Check if being deleted
	if !autoSecret.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	// Create basic-auth secret
	basicAuthSecretName := fmt.Sprintf("%s-basic-auth", autoSecret.Name)
	basicAuthSecret, err := r.reconcileBasicAuthSecret(ctx, &autoSecret, basicAuthSecretName)
	if err != nil {
		log.Error(err, "Failed to reconcile basic-auth secret")
		return ctrl.Result{}, err
	}

	// Create DB URI secret
	dbURISecretName := fmt.Sprintf("%s-db-uri", autoSecret.Name)
	_, err = r.reconcileDBURISecret(ctx, &autoSecret, dbURISecretName, basicAuthSecret)
	if err != nil {
		log.Error(err, "Failed to reconcile DB URI secret")
		return ctrl.Result{}, err
	}

	// Update status
	autoSecret.Status.BasicAuthSecretName = basicAuthSecretName
	autoSecret.Status.DBURISecretName = dbURISecretName
	if err := r.Status().Update(ctx, &autoSecret); err != nil {
		log.Error(err, "Failed to update AutoSecret status")
		return ctrl.Result{}, err
	}

	log.Info("Successfully reconciled AutoSecret",
		"name", autoSecret.Name,
		"basicAuthSecret", basicAuthSecretName,
		"dbURISecret", dbURISecretName)

	return ctrl.Result{}, nil
}

func (r *AutoSecretReconciler) reconcileBasicAuthSecret(ctx context.Context, autoSecret *autosecretv1alpha1.AutoSecret, secretName string) (*corev1.Secret, error) {
	log := log.FromContext(ctx)

	// Check if secret already exists
	var existingSecret corev1.Secret
	err := r.Get(ctx, client.ObjectKey{Name: secretName, Namespace: autoSecret.Namespace}, &existingSecret)

	if err == nil {
		// Secret exists, check if password is already set
		if _, hasPassword := existingSecret.Data["password"]; hasPassword {
			log.Info("Basic-auth secret already exists with password", "name", secretName)
			return &existingSecret, nil
		}
		// Secret exists but no password, update it
		password, err := randomPassword(32)
		if err != nil {
			return nil, fmt.Errorf("failed to generate password: %w", err)
		}
		existingSecret.Data = map[string][]byte{
			"username": []byte(autoSecret.Spec.Username),
			"password": []byte(password),
		}
		if err := r.Update(ctx, &existingSecret); err != nil {
			return nil, fmt.Errorf("failed to update basic-auth secret: %w", err)
		}
		log.Info("Updated basic-auth secret with password", "name", secretName)
		return &existingSecret, nil
	}

	if !apierrors.IsNotFound(err) {
		return nil, err
	}

	// Secret doesn't exist, create it
	password, err := randomPassword(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate password: %w", err)
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: autoSecret.Namespace,
		},
		Type: corev1.SecretTypeBasicAuth,
		Data: map[string][]byte{
			"username": []byte(autoSecret.Spec.Username),
			"password": []byte(password),
		},
	}

	// Set owner reference
	if err := controllerutil.SetControllerReference(autoSecret, secret, r.Scheme); err != nil {
		return nil, err
	}

	if err := r.Create(ctx, secret); err != nil {
		return nil, fmt.Errorf("failed to create basic-auth secret: %w", err)
	}
	log.Info("Created basic-auth secret", "name", secretName)

	return secret, nil
}

func (r *AutoSecretReconciler) reconcileDBURISecret(ctx context.Context, autoSecret *autosecretv1alpha1.AutoSecret, secretName string, basicAuthSecret *corev1.Secret) (*corev1.Secret, error) {
	log := log.FromContext(ctx)

	username := string(basicAuthSecret.Data["username"])
	password := string(basicAuthSecret.Data["password"])

	port := autoSecret.Spec.Port
	if port == 0 {
		port = 5432
	}

	// Build PostgreSQL URI
	uri := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s",
		url.QueryEscape(username),
		url.QueryEscape(password),
		autoSecret.Spec.DBHost,
		port,
		autoSecret.Spec.DBName)

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: autoSecret.Namespace,
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"DATABASE_URI": []byte(uri),
			"DB_HOST":      []byte(autoSecret.Spec.DBHost),
			"DB_NAME":      []byte(autoSecret.Spec.DBName),
			"DB_PORT":      []byte(fmt.Sprintf("%d", port)),
			"DB_USER":      []byte(username),
			"DB_PASSWORD":  []byte(password),
		},
	}

	// Set owner reference
	if err := controllerutil.SetControllerReference(autoSecret, secret, r.Scheme); err != nil {
		return nil, err
	}

	// Check if secret already exists
	var existingSecret corev1.Secret
	err := r.Get(ctx, client.ObjectKey{Name: secretName, Namespace: autoSecret.Namespace}, &existingSecret)

	if apierrors.IsNotFound(err) {
		// Create new secret
		if err := r.Create(ctx, secret); err != nil {
			return nil, fmt.Errorf("failed to create DB URI secret: %w", err)
		}
		log.Info("Created DB URI secret", "name", secretName)
	} else if err == nil {
		// Update existing secret
		existingSecret.Data = secret.Data
		if err := r.Update(ctx, &existingSecret); err != nil {
			return nil, fmt.Errorf("failed to update DB URI secret: %w", err)
		}
		log.Info("Updated DB URI secret", "name", secretName)
		return &existingSecret, nil
	} else {
		return nil, err
	}

	return secret, nil
}

// SetupWithManager sets up the controller with the Manager
func (r *AutoSecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&autosecretv1alpha1.AutoSecret{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}
