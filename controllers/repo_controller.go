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
	"fmt"
	"time"

	"github.com/aws/smithy-go/ptr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	srcv1 "bwag.me/git-stats-exporter/api/v1"
	"bwag.me/git-stats-exporter/pkg/repos"
)

// RepoReconciler reconciles a Repo object
type RepoReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=src.bwag.me,resources=repos,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=src.bwag.me,resources=repos/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=src.bwag.me,resources=repos/finalizers,verbs=update

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

	var repoMetrics *repos.Repos
	if repo.Spec.GHTokenSecretRef != nil {
		var ghtoken corev1.Secret
		if err := r.Get(ctx, types.NamespacedName{Namespace: repo.Namespace, Name: *repo.Spec.GHTokenSecretRef}, &ghtoken); err != nil {
			log.Error(err, "unable to fetch the secret reference")
			return ctrl.Result{}, err
		}
		if token, ok := ghtoken.Data["token"]; ok {
			repoMetrics = repos.New(ctx, token)
		} else {
			err := fmt.Errorf("unable to fetch the token key from the secret reference")
			log.Error(err, "provide the token key under data in the secret")
			return ctrl.Result{}, err
		}
	} else {
		repoMetrics = repos.New(ctx, nil)
	}

	if err := repoMetrics.EmitMetrics(ctx, repo.Spec.Owner, repo.Spec.Name); err != nil {
		log.Error(err, fmt.Sprintf("unable to emit Repo (%s/%s) metrics", repo.Spec.Owner, repo.Spec.Name))
		return ctrl.Result{}, err
	}

	repo.Status.State = ptr.String("Synchronized")
	repo.Status.LastQuery = metav1.Now()
	if err := r.Status().Update(ctx, &repo); err != nil {
		log.Error(err, "unable to update Repo status")
		return ctrl.Result{}, err
	}

	log.Info(fmt.Sprintf("Reconciled \"%s/%s\"", repo.Spec.Owner, repo.Spec.Name))
	return ctrl.Result{RequeueAfter: 5 * time.Minute}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RepoReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&srcv1.Repo{}).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(r)
}
