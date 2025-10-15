package controllers

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	autosecretv1alpha1 "github.com/SindreMA/auto-secret-operator/api/v1alpha1"
)

// AutoSecretDbSecretRedirectReconciler reconciles an AutoSecretDbSecretRedirect object
type AutoSecretDbSecretRedirectReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=auto-secret.io,resources=autosecretdbsecretredirects,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=auto-secret.io,resources=autosecretdbsecretredirects/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=auto-secret.io,resources=autosecretdbsecretredirects/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete

// Reconcile handles AutoSecretDbSecretRedirect resources
func (r *AutoSecretDbSecretRedirectReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Fetch the AutoSecretDbSecretRedirect instance
	var redirect autosecretv1alpha1.AutoSecretDbSecretRedirect
	if err := r.Get(ctx, req.NamespacedName, &redirect); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Check if being deleted
	if !redirect.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	// Determine target secret name
	targetSecretName := redirect.Spec.TargetSecretName
	if targetSecretName == "" {
		targetSecretName = redirect.Spec.SecretName + "-redirect"
	}

	// Get the source secret
	var sourceSecret corev1.Secret
	err := r.Get(ctx, client.ObjectKey{
		Name:      redirect.Spec.SecretName,
		Namespace: redirect.Namespace,
	}, &sourceSecret)

	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("Source secret not found", "secret", redirect.Spec.SecretName)
			return ctrl.Result{Requeue: true}, nil
		}
		return ctrl.Result{}, err
	}

	// Check if we need to update (source secret has changed)
	if redirect.Status.SourceSecretResourceVersion == sourceSecret.ResourceVersion {
		log.V(1).Info("Source secret unchanged, skipping reconciliation")
		return ctrl.Result{}, nil
	}

	// Reconcile the target secret
	if err := r.reconcileTargetSecret(ctx, &redirect, &sourceSecret, targetSecretName); err != nil {
		log.Error(err, "Failed to reconcile target secret")
		return ctrl.Result{}, err
	}

	// Update status
	redirect.Status.TargetSecretName = targetSecretName
	redirect.Status.SourceSecretResourceVersion = sourceSecret.ResourceVersion
	if err := r.Status().Update(ctx, &redirect); err != nil {
		log.Error(err, "Failed to update status")
		return ctrl.Result{}, err
	}

	log.Info("Successfully reconciled AutoSecretDbSecretRedirect",
		"name", redirect.Name,
		"source", redirect.Spec.SecretName,
		"target", targetSecretName)

	return ctrl.Result{}, nil
}

func (r *AutoSecretDbSecretRedirectReconciler) reconcileTargetSecret(
	ctx context.Context,
	redirect *autosecretv1alpha1.AutoSecretDbSecretRedirect,
	sourceSecret *corev1.Secret,
	targetSecretName string,
) error {
	log := log.FromContext(ctx)

	// Extract URI from source secret
	uriData, hasURI := sourceSecret.Data["uri"]
	if !hasURI {
		return fmt.Errorf("source secret does not contain 'uri' field")
	}

	uri := string(uriData)

	// Transform URI to different formats
	transformedData, err := r.transformURI(uri, sourceSecret.Data)
	if err != nil {
		return fmt.Errorf("failed to transform URI: %w", err)
	}

	// Check if target secret exists
	var existingSecret corev1.Secret
	err = r.Get(ctx, client.ObjectKey{
		Name:      targetSecretName,
		Namespace: redirect.Namespace,
	}, &existingSecret)

	if err == nil {
		// Update existing secret
		existingSecret.Data = transformedData
		if err := r.Update(ctx, &existingSecret); err != nil {
			return fmt.Errorf("failed to update target secret: %w", err)
		}
		log.Info("Updated target secret", "name", targetSecretName)
	} else if apierrors.IsNotFound(err) {
		// Create new secret
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      targetSecretName,
				Namespace: redirect.Namespace,
			},
			Type: corev1.SecretTypeOpaque,
			Data: transformedData,
		}

		// Set owner reference
		if err := controllerutil.SetControllerReference(redirect, secret, r.Scheme); err != nil {
			return err
		}

		if err := r.Create(ctx, secret); err != nil {
			return fmt.Errorf("failed to create target secret: %w", err)
		}
		log.Info("Created target secret", "name", targetSecretName)
	} else {
		return err
	}

	return nil
}

// transformURI takes a database URI and creates multiple format variations
func (r *AutoSecretDbSecretRedirectReconciler) transformURI(uri string, sourceData map[string][]byte) (map[string][]byte, error) {
	// Parse the URI
	parsedURI, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("invalid URI format: %w", err)
	}

	result := make(map[string][]byte)

	// Copy original URI
	result["uri"] = []byte(uri)
	result["original-uri"] = []byte(uri)

	// Extract components
	scheme := parsedURI.Scheme
	username := parsedURI.User.Username()
	password, _ := parsedURI.User.Password()
	host := parsedURI.Hostname()
	port := parsedURI.Port()
	if port == "" {
		port = "5432" // default PostgreSQL port
	}
	dbname := strings.TrimPrefix(parsedURI.Path, "/")
	query := parsedURI.RawQuery

	// Store individual components
	result["username"] = []byte(username)
	result["password"] = []byte(password)
	result["host"] = []byte(host)
	result["port"] = []byte(port)
	result["dbname"] = []byte(dbname)

	// Build ms-uri (Microsoft SQL connection string format for PostgreSQL)
	// Format: Server=host;Port=port;Database=dbname;User Id=username;Password=password;
	msURI := fmt.Sprintf("Server=%s;Port=%s;Database=%s;User Id=%s;Password=%s;",
		host, port, dbname, username, password)
	result["ms-uri"] = []byte(msURI)

	// Build odbc-uri (ODBC connection string)
	odbcURI := fmt.Sprintf("Driver={PostgreSQL Unicode};Server=%s;Port=%s;Database=%s;Uid=%s;Pwd=%s;",
		host, port, dbname, username, password)
	result["odbc-uri"] = []byte(odbcURI)

	// Build ADO.NET connection string
	adoNetURI := fmt.Sprintf("Host=%s;Port=%s;Database=%s;Username=%s;Password=%s;",
		host, port, dbname, username, password)
	result["adonet-uri"] = []byte(adoNetURI)

	// Build JDBC URI
	jdbcURI := fmt.Sprintf("jdbc:%s://%s:%s/%s?user=%s&password=%s",
		scheme, host, port, dbname, url.QueryEscape(username), url.QueryEscape(password))
	if query != "" {
		jdbcURI += "&" + query
	}
	result["jdbc-uri"] = []byte(jdbcURI)

	// Copy other useful fields from source if they exist
	for _, key := range []string{"fqdn-uri", "fqdn-jdbc-uri", "pgpass", "user"} {
		if val, exists := sourceData[key]; exists {
			result[key] = val
		}
	}

	return result, nil
}

// SetupWithManager sets up the controller with the Manager
func (r *AutoSecretDbSecretRedirectReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&autosecretv1alpha1.AutoSecretDbSecretRedirect{}).
		Owns(&corev1.Secret{}).
		Watches(
			&corev1.Secret{},
			handler.EnqueueRequestsFromMapFunc(r.findRedirectsForSecret),
		).
		Complete(r)
}

// findRedirectsForSecret finds all AutoSecretDbSecretRedirect resources that reference a given Secret
func (r *AutoSecretDbSecretRedirectReconciler) findRedirectsForSecret(ctx context.Context, obj client.Object) []reconcile.Request {
	secret := obj.(*corev1.Secret)

	var redirectList autosecretv1alpha1.AutoSecretDbSecretRedirectList
	if err := r.List(ctx, &redirectList, client.InNamespace(secret.Namespace)); err != nil {
		return []reconcile.Request{}
	}

	var requests []reconcile.Request
	for _, redirect := range redirectList.Items {
		if redirect.Spec.SecretName == secret.Name {
			requests = append(requests, reconcile.Request{
				NamespacedName: client.ObjectKey{
					Name:      redirect.Name,
					Namespace: redirect.Namespace,
				},
			})
		}
	}

	return requests
}
