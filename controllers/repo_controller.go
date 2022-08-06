/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"

	"github.com/aws/smithy-go/ptr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	srcv1 "bwag.me/git-stats-exporter/api/v1"
)

// RepoReconciler reconciles a Repo object
type RepoReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=src.bwag.me,resources=repoes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=src.bwag.me,resources=repoes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=src.bwag.me,resources=repoes/finalizers,verbs=update

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *RepoReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	var repo srcv1.Repo
	if err := r.Get(ctx, req.NamespacedName, &repo); err != nil {
		log.Error(err, "unable to fetch Repo")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	//TODO: retrieve github stats

	repo.Status.State = ptr.String("Synchronized")
	repo.Status.LastQuery = metav1.Now()
	if err := r.Status().Update(ctx, &repo); err != nil {
		log.Error(err, "unable to update Repo status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RepoReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&srcv1.Repo{}).
		Complete(r)
}
