package microservice

import (
	"context"
	"reflect"

	appsv1 "github.com/openshift/api/apps/v1"
	routev1 "github.com/openshift/api/route/v1"
	paasv1alpha1 "github.com/ykoer/microservice-operator/pkg/apis/paas/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_microservice")

// RouteIpaManaged annotation
const RouteIpaManaged = "cert.patrickeasters.com/ipa-managed"

// ConfigVolume Mount
const ConfigVolume = "/etc/config-volume"

// Labels for BuildConfig, DeploymentConfig, Service and Route
type Labels = map[string]string

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new MicroService Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileMicroService{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("microservice-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource MicroService
	err = c.Watch(&source.Kind{Type: &paasv1alpha1.MicroService{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileMicroService{}

// ReconcileMicroService reconciles a MicroService object
type ReconcileMicroService struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a MicroService object and makes changes based on the state read
// and what is in the MicroService.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileMicroService) Reconcile(request reconcile.Request) (reconcile.Result, error) {

	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling MicroService")

	// Fetch the MicroService instance
	instance := &paasv1alpha1.MicroService{}

	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	labels := Labels{
		"application": instance.Name,
	}

	// Check if the deployment already exists, if not create a new one
	deploymentconfig := r.newDeploymentForCR(instance, labels)
	foundDeploymentconfig := &appsv1.DeploymentConfig{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, foundDeploymentconfig)

	if err != nil && errors.IsNotFound(err) {
		// Define a new DeploymentConfig
		reqLogger.Info("Creating a new DeploymentConfig")
		err = r.client.Create(context.TODO(), deploymentconfig)
		if err != nil {
			reqLogger.Info("Failed to create new DeploymentConfig")
			return reconcile.Result{}, err
		}
		// DeploymentConfig created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Info("Failed to get DeploymentConfig")
		return reconcile.Result{}, err
	}

	if !reflect.DeepEqual(foundDeploymentconfig.Spec, deploymentconfig.Spec) {
		foundDeploymentconfig.Spec.Replicas = instance.Spec.Replicas
		for i, v := range foundDeploymentconfig.Spec.Template.Spec.Containers {
			if v.Name == instance.Name {
				foundDeploymentconfig.Spec.Template.Spec.Containers[i].Image = instance.Spec.Image
			}
		}

		err := r.client.Update(context.TODO(), foundDeploymentconfig)
		if err != nil {
			reqLogger.Info("failed to update deployment:" + err.Error())
		} else if err != nil {
			reqLogger.Info("Failed to update Deployment")
			return reconcile.Result{}, err
		}
	}

	foundService := &corev1.Service{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, foundService)
	if err != nil && errors.IsNotFound(err) {
		// Define a new service
		dep := r.newServiceForCR(instance, labels)
		reqLogger.Info("Creating a new Service")
		err = r.client.Create(context.TODO(), dep)
		if err != nil {
			reqLogger.Info("Failed to create new Service")
			return reconcile.Result{}, err
		}
		// Service created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Info("Failed to get Service")
		return reconcile.Result{}, err
	}

	foundRoute := &routev1.Route{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, foundRoute)
	if err != nil && errors.IsNotFound(err) {
		// Define a new route
		dep := r.newRouteForCR(instance, labels)
		reqLogger.Info("Creating a new Route")
		err = r.client.Create(context.TODO(), dep)
		if err != nil {
			reqLogger.Info("Failed to create new Route")
			return reconcile.Result{}, err
		}
		// Route created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Info("Failed to get Route")
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

// Create newDeploymentForCR method to create a deployment.
func (r *ReconcileMicroService) newDeploymentForCR(cr *paasv1alpha1.MicroService, labels Labels) *appsv1.DeploymentConfig {

	replicas := cr.Spec.Replicas

	dc := &appsv1.DeploymentConfig{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "DeploymentConfig",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        cr.Name,
			Namespace:   cr.Namespace,
			Labels:      labels,
			Annotations: map[string]string{"com.redhat.paas/monitored": "true"},
		},
		Spec: appsv1.DeploymentConfigSpec{
			Replicas: replicas,
			Selector: labels,
			Template: &corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: cr.Spec.Image,
						Name:  cr.Name,
						Ports: []corev1.ContainerPort{{
							ContainerPort: 8080,
							Name:          "http",
						}},
						VolumeMounts: []corev1.VolumeMount{
							{Name: "config-volume", MountPath: ConfigVolume},
						},
						Env: []corev1.EnvVar{
							corev1.EnvVar{
								Name:  "JAVA_OPTIONS",
								Value: "-Dspring.config.additional-location=file:" + ConfigVolume,
							},
						},
					}},
					Volumes: []corev1.Volume{
						r.getConfigVolume(cr),
					},
				},
			},
		},
	}
	// Set Deployment instance as the owner and controller
	controllerutil.SetControllerReference(cr, dc, r.scheme)
	return dc
}

func (r *ReconcileMicroService) getConfigVolume(cr *paasv1alpha1.MicroService) corev1.Volume {

	volume := corev1.Volume{}

	err := r.client.Get(context.TODO(), types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}, &corev1.ConfigMap{})
	if err == nil {
		volume = corev1.Volume{
			Name: "config-volume",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: cr.Name,
					},
				},
			},
		}
	} else if err != nil && errors.IsNotFound(err) {
		volume = corev1.Volume{
			Name: "config-volume",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		}
	}
	return volume
}

func (r *ReconcileMicroService) newServiceForCR(cr *paasv1alpha1.MicroService, labels Labels) *corev1.Service {

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        cr.Name,
			Namespace:   cr.Namespace,
			Labels:      labels,
			Annotations: map[string]string{"service.alpha.openshift.io/serving-cert-secret-name": cr.Name + "-tls"},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:     "http-8080",
					Protocol: "TCP",
					Port:     8080,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 8080,
					},
				},
			},
			Selector: labels,
		},
	}

	// Set Service instance as the owner and controller
	controllerutil.SetControllerReference(cr, service, r.scheme)
	return service
}

func (r *ReconcileMicroService) newRouteForCR(cr *paasv1alpha1.MicroService, labels Labels) *routev1.Route {

	annotations := map[string]string{}
	if cr.ObjectMeta.Annotations[RouteIpaManaged] == "true" {
		annotations[RouteIpaManaged] = "true"
	} else {
		annotations[RouteIpaManaged] = "false"
	}

	route := &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:        cr.Name,
			Namespace:   cr.Namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: routev1.RouteSpec{
			To: routev1.RouteTargetReference{
				Name: cr.Name,
			},
			Host: cr.Spec.Hostname,
			Port: &routev1.RoutePort{
				TargetPort: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: 8080,
				},
			},
			TLS: &routev1.TLSConfig{
				Termination:                   routev1.TLSTerminationEdge,
				InsecureEdgeTerminationPolicy: routev1.InsecureEdgeTerminationPolicyRedirect,
			},
		},
	}

	// Set Route instance as the owner and controller
	controllerutil.SetControllerReference(cr, route, r.scheme)
	return route
}
