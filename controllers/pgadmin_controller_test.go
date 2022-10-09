package controllers

import (
	"context"
	pgadminv1alpha1 "github.com/dhope-nagesh/pgadmin-operator/api/v1alpha1"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	v12 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"time"
)

const (
	PgadminName            = "pgadmin-test"
	PgadminNamespace       = "default"
	PgadminReplicas  int32 = 10
	timeout                = time.Second * 10
	interval               = time.Millisecond * 250
)

var _ = ginkgo.Describe("Pgadmin controller", func() {
	ginkgo.Context("When creating Pgadmin", func() {
		ginkgo.It("Create pgadmin", func() {
			ginkgo.By("By create pgadmin")
			ctx := context.Background()
			pgadmin := &pgadminv1alpha1.Pgadmin{
				ObjectMeta: metav1.ObjectMeta{
					Name:      PgadminName,
					Namespace: PgadminNamespace,
				},
				Spec: pgadminv1alpha1.PgadminSpec{
					Replicas:        PgadminReplicas,
					CredsSecretName: "test-creds",
				},
			}
			gomega.Expect(k8sClient.Create(ctx, pgadmin)).Should(gomega.Succeed())

			deploymentName := "deployment-" + pgadmin.Name
			serviceName := "service-" + pgadmin.Name

			deploymentNameLookUpKey := types.NamespacedName{Name: deploymentName, Namespace: pgadmin.Namespace}
			serviceNameLookUpKey := types.NamespacedName{Name: serviceName, Namespace: pgadmin.Namespace}

			deployment := &appsv1.Deployment{}

			gomega.Eventually(func() bool {
				err := k8sClient.Get(ctx, deploymentNameLookUpKey, deployment)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(gomega.BeTrue())

			gomega.Expect(deployment.Spec.Replicas).To(gomega.Equal(Int32ToPtr(PgadminReplicas)))

			service := &v12.Service{}

			gomega.Eventually(func() bool {
				err := k8sClient.Get(ctx, serviceNameLookUpKey, service)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(gomega.BeTrue())

		})
	})
})
