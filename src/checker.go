package jwtexporter

import (
	"context"
	"time"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Checker struct {
	period         time.Duration
	labelSelectors []string
	kubeconfigPath string
	exporter       *Exporter
	annotationKey  string
	logger         *zap.Logger
}

// GetChecker is a factory method that returns a new PeriodicSecretChecker
func GetChecker(
	period time.Duration,
	labelSelectors []string,
	kubeconfigPath string,
	e *Exporter,
	annotationKey string,
	logger *zap.Logger,
) *Checker {
	return &Checker{
		period:         period,
		labelSelectors: labelSelectors,
		kubeconfigPath: kubeconfigPath,
		annotationKey:  annotationKey,
		exporter:       e,
		logger:         logger,
	}
}

func (p *Checker) StartChecking() {
	config, err := clientcmd.BuildConfigFromFlags("", p.kubeconfigPath)
	if err != nil {
		p.logger.Fatal("Error building from kubeconfig", zap.String("kubeconfigPath", p.kubeconfigPath), zap.Error(err))
	}

	// creates the clientset
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		p.logger.Fatal("kubernetes.NewForConfig failed", zap.Error(err))
	}

	periodChannel := time.Tick(p.period)
	for {
		p.logger.Info("Begin periodic check")

		p.exporter.ResetMetrics()

		var secrets []corev1.Secret
		var namespacesToCheck []string

		var nss *corev1.NamespaceList
		nss, err = client.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			p.logger.Error("Error requesting namespaces", zap.Error(err))
			ErrorTotal.Inc()
		}
		for _, ns := range nss.Items {
			namespacesToCheck = append(namespacesToCheck, ns.GetObjectMeta().GetName())
			p.logger.Debug("Adding namespace to check", zap.String("namespace", ns.GetObjectMeta().GetName()))
		}

		for _, ns := range namespacesToCheck {
			if len(p.labelSelectors) > 0 {
				for _, labelSelector := range p.labelSelectors {
					var s *corev1.SecretList
					s, err = client.CoreV1().Secrets(ns).List(context.TODO(), metav1.ListOptions{
						LabelSelector: labelSelector,
					})
					if err != nil {
						p.logger.Error("Error requesting secrets", zap.Error(err))
						ErrorTotal.Inc()
						continue
					}
					secrets = append(secrets, s.Items...)
				}
			} else {
				var s *corev1.SecretList
				s, err = client.CoreV1().Secrets(ns).List(context.TODO(), metav1.ListOptions{})
				if err != nil {
					p.logger.Error("Error requesting secrets", zap.Error(err))
					ErrorTotal.Inc()
					continue
				}
				secrets = append(secrets, s.Items...)
			}
		}

		for _, secret := range secrets {
			p.logger.Info(
				"Reviewing secret",
				zap.String("secret_name",
				secret.GetName()),
				zap.String("secret_namespace",
				secret.GetNamespace()),
			)
			var secretKey, token string
			secretKey = secret.Annotations[p.annotationKey]
			token = string(secret.Data[secretKey])
			err = p.exporter.ExportMetrics(token, secretKey, secret.Name, secret.Namespace)
			if err != nil {
				p.logger.Error(
					"Error exporting metrics for secret",
					zap.String("secret_name",
					secret.GetName()),
					zap.String("secret_namespace",
					secret.GetNamespace()),
					zap.Error(err),
				)
				ErrorTotal.Inc()
			} else {
				p.logger.Info(
					"Metrics exported for secret",
					zap.String("secret_name",
					secret.GetName()),
					zap.String("secret_namespace",
					secret.GetNamespace()),
				)
			}
		}
		<-periodChannel
	}
}
