package controllers

import (
	"context"
	"fmt"
	"net/url"

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

// AutoSecretDbReconciler reconciles an AutoSecretDb object
type AutoSecretDbReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=auto-secret.io,resources=autosecretdbs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=auto-secret.io,resources=autosecretdbs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=auto-secret.io,resources=autosecretdbs/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete

// Reconcile handles AutoSecretDb resources
func (r *AutoSecretDbReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Fetch the AutoSecretDb instance
	var autoSecretDb autosecretv1alpha1.AutoSecretDb
	if err := r.Get(ctx, req.NamespacedName, &autoSecretDb); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Check if being deleted
	if !autoSecretDb.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	// Determine secret name
	secretName := autoSecretDb.Spec.SecretName
	if secretName == "" {
		secretName = autoSecretDb.Name
	}

	// Reconcile secret
	if err := r.reconcileSecret(ctx, &autoSecretDb, secretName); err != nil {
		log.Error(err, "Failed to reconcile secret")
		return ctrl.Result{}, err
	}

	// Update status
	autoSecretDb.Status.SecretName = secretName
	if err := r.Status().Update(ctx, &autoSecretDb); err != nil {
		log.Error(err, "Failed to update AutoSecretDb status")
		return ctrl.Result{}, err
	}

	log.Info("Successfully reconciled AutoSecretDb",
		"name", autoSecretDb.Name,
		"secret", secretName)

	return ctrl.Result{}, nil
}

func (r *AutoSecretDbReconciler) reconcileSecret(ctx context.Context, autoSecretDb *autosecretv1alpha1.AutoSecretDb, secretName string) error {
	log := log.FromContext(ctx)

	// Check if secret already exists
	var existingSecret corev1.Secret
	err := r.Get(ctx, client.ObjectKey{Name: secretName, Namespace: autoSecretDb.Namespace}, &existingSecret)

	var password string

	if err == nil {
		// Secret exists, check if password is already set
		if existingPassword, hasPassword := existingSecret.Data["password"]; hasPassword {
			log.Info("Secret already exists with password", "name", secretName)
			password = string(existingPassword)
		} else {
			// Generate new password
			var err error
			password, err = r.generatePassword(autoSecretDb)
			if err != nil {
				return fmt.Errorf("failed to generate password: %w", err)
			}
		}
	} else if apierrors.IsNotFound(err) {
		// Secret doesn't exist, generate password
		var err error
		password, err = r.generatePassword(autoSecretDb)
		if err != nil {
			return fmt.Errorf("failed to generate password: %w", err)
		}
	} else {
		return err
	}

	// Build secret data
	secretData := r.buildSecretData(autoSecretDb, password)

	if err == nil {
		// Update existing secret
		existingSecret.Data = secretData
		// Copy labels and annotations from AutoSecretDb to Secret
		if existingSecret.Labels == nil {
			existingSecret.Labels = make(map[string]string)
		}
		for k, v := range autoSecretDb.Labels {
			existingSecret.Labels[k] = v
		}
		if existingSecret.Annotations == nil {
			existingSecret.Annotations = make(map[string]string)
		}
		for k, v := range autoSecretDb.Annotations {
			existingSecret.Annotations[k] = v
		}
		if err := r.Update(ctx, &existingSecret); err != nil {
			return fmt.Errorf("failed to update secret: %w", err)
		}
		log.Info("Updated secret", "name", secretName)
	} else {
		// Create new secret
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:        secretName,
				Namespace:   autoSecretDb.Namespace,
				Labels:      make(map[string]string),
				Annotations: make(map[string]string),
			},
			Type: corev1.SecretTypeBasicAuth,
			Data: secretData,
		}

		// Copy labels and annotations from AutoSecretDb to Secret
		for k, v := range autoSecretDb.Labels {
			secret.Labels[k] = v
		}
		for k, v := range autoSecretDb.Annotations {
			secret.Annotations[k] = v
		}

		// Set owner reference
		if err := controllerutil.SetControllerReference(autoSecretDb, secret, r.Scheme); err != nil {
			return err
		}

		if err := r.Create(ctx, secret); err != nil {
			return fmt.Errorf("failed to create secret: %w", err)
		}
		log.Info("Created secret", "name", secretName)
	}

	return nil
}

func (r *AutoSecretDbReconciler) generatePassword(autoSecretDb *autosecretv1alpha1.AutoSecretDb) (string, error) {
	length := autoSecretDb.Spec.PasswordLength
	if length == 0 {
		length = 30
	}

	charset := autoSecretDb.Spec.PasswordCharset
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

func (r *AutoSecretDbReconciler) buildSecretData(autoSecretDb *autosecretv1alpha1.AutoSecretDb, password string) map[string][]byte {
	port := autoSecretDb.Spec.Port
	if port == 0 {
		port = 5432
	}

	dbType := autoSecretDb.Spec.DBType
	if dbType == "" {
		dbType = "postgresql"
	}

	username := autoSecretDb.Spec.Username
	dbname := autoSecretDb.Spec.DBName
	dbhost := autoSecretDb.Spec.DBHost

	// URL-encode credentials
	encodedUser := url.QueryEscape(username)
	encodedPass := url.QueryEscape(password)

	// Build URIs
	uri := fmt.Sprintf("%s://%s:%s@%s:%d/%s",
		dbType, encodedUser, encodedPass, dbhost, port, dbname)

	if autoSecretDb.Spec.AdditionalParams != "" {
		uri += autoSecretDb.Spec.AdditionalParams
	}

	jdbcURI := fmt.Sprintf("jdbc:%s://%s:%d/%s?password=%s&user=%s",
		dbType, dbhost, port, dbname, encodedPass, encodedUser)

	if autoSecretDb.Spec.AdditionalParams != "" {
		jdbcURI += "&" + autoSecretDb.Spec.AdditionalParams[1:] // skip leading '?'
	}

	// Extract short hostname (before first dot)
	shortHost := dbhost
	for i, ch := range dbhost {
		if ch == '.' {
			shortHost = dbhost[:i]
			break
		}
	}

	// pgpass format: hostname:port:database:username:password
	pgpass := fmt.Sprintf("%s:%d:%s:%s:%s", shortHost, port, dbname, username, password)

	return map[string][]byte{
		"dbname":       []byte(dbname),
		"fqdn-jdbc-uri": []byte(jdbcURI),
		"fqdn-uri":     []byte(uri),
		"host":         []byte(shortHost),
		"jdbc-uri":     []byte(jdbcURI),
		"password":     []byte(password),
		"pgpass":       []byte(pgpass),
		"port":         []byte(fmt.Sprintf("%d", port)),
		"uri":          []byte(uri),
		"user":         []byte(username),
		"username":     []byte(username),
	}
}

// SetupWithManager sets up the controller with the Manager
func (r *AutoSecretDbReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&autosecretv1alpha1.AutoSecretDb{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}
