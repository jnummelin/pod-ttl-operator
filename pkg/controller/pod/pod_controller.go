package pod

import (
	"context"
	"log"
	"strconv"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// Add creates a new Pod Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcilePod{client: mgr.GetClient(), scheme: mgr.GetScheme(), timers: make(map[string]*time.Timer)}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("pod-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to Pod
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForObject{})
	log.Printf("Created watch for Pods")
	if err != nil {
		return err
	}

	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcilePod{}

// ReconcilePod reconciles a Pod object
type ReconcilePod struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
	timers map[string]*time.Timer
}

// Reconcile managed the timers for the Pod TTLs
//
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcilePod) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the Pod instance
	pod := &corev1.Pod{}
	err := r.client.Get(context.TODO(), request.NamespacedName, pod)
	if err != nil {
		if errors.IsNotFound(err) {
			// Pod already gone
			timer, found := r.timers[request.Namespace+"/"+request.Name]
			if found {
				log.Printf("Found existing timer for %s/%s, stopping&removing it as the pod is gone.", request.Namespace, request.Name)
				timer.Stop()
				delete(r.timers, request.Namespace+"/"+request.Name)
			}
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Printf("Getting Pod %s/%s failed:%s\n", request.Namespace, request.Name, err)
		return reconcile.Result{}, err
	}

	ttl := pod.Annotations["nummel.in/pod-ttl"]
	if ttl == "" {
		log.Printf("Pod does not have TTL annotation, ignoring")
		return reconcile.Result{}, nil
	}

	var podReadyAt *time.Time
	for _, c := range pod.Status.Conditions {
		if c.Type == "Ready" && c.Status == "True" {
			podReadyAt = &c.LastTransitionTime.Time
			log.Printf("Pod %s/%s ready at:%s\n", request.Namespace, request.Name, podReadyAt)
		}
	}

	if podReadyAt == nil {
		log.Printf("Pod does not have ready condition, ignoring")
		return reconcile.Result{}, nil
	}

	ttl_secs, err := strconv.Atoi(ttl)
	if err != nil {
		log.Printf("Error reading TTL annotation: %s\n", err)
		return reconcile.Result{}, err
	}

	timer, found := r.timers[request.Namespace+"/"+request.Name]
	if found {
		log.Printf("Timer already existing for %s/%s\n", request.Namespace, request.Name)
		return reconcile.Result{}, nil
	}

	pod_ttl := podReadyAt.Add(time.Duration(ttl_secs) * time.Second).Sub(time.Now())
	log.Printf("Creating timer with duration: %s\n", pod_ttl)

	timer = time.NewTimer(pod_ttl)
	r.timers[request.Namespace+"/"+request.Name] = timer
	go func() {
		<-timer.C
		log.Printf("Timer for Pod %s/%s expired, should go and kill the pod now\n", request.Namespace, request.Name)
		err := r.client.Delete(context.TODO(), pod)
		if err != nil {
			log.Printf("ERROR deleting the pod: %s\n", err)
		}
	}()

	return reconcile.Result{}, nil
}
