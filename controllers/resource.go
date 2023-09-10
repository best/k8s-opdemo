package controllers

import (
	"github.com/best/k8s-opdemo/api/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func NewDeploy(app *v1beta1.AppService) *appsv1.Deployment {
	labels := map[string]string{
		"app": app.Name,
	}
	selector := &metav1.LabelSelector{
		MatchLabels: labels,
	}

	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Name,
			Namespace: app.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(app, schema.GroupVersionKind{
					Kind:    v1beta1.Kind,
					Group:   v1beta1.GroupVersion.Group,
					Version: v1beta1.GroupVersion.Version,
				}),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: app.Spec.Size,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: newContainers(app),
				},
			},
			Selector:                selector,
			Strategy:                appsv1.DeploymentStrategy{},
			MinReadySeconds:         0,
			RevisionHistoryLimit:    nil,
			Paused:                  false,
			ProgressDeadlineSeconds: nil,
		},
	}

}

func newContainers(app *v1beta1.AppService) []corev1.Container {
	var containerPorts []corev1.ContainerPort
	for _, port := range app.Spec.Ports {
		containerPorts = append(containerPorts, corev1.ContainerPort{
			Name:          port.Name,
			ContainerPort: port.TargetPort.IntVal,
			Protocol:      port.Protocol,
		})
	}

	container := corev1.Container{
		Name:      app.Name,
		Image:     app.Spec.Image,
		Ports:     containerPorts,
		Resources: app.Spec.Resources,
		Env:       app.Spec.Envs,
	}

	return []corev1.Container{container}
}

func NewService(app *v1beta1.AppService) *corev1.Service {
	labels := map[string]string{
		"app": app.Name,
	}

	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Name,
			Namespace: app.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(app, schema.GroupVersionKind{
					Kind:    v1beta1.Kind,
					Group:   v1beta1.GroupVersion.Group,
					Version: v1beta1.GroupVersion.Version,
				}),
			},
		},
		Spec: corev1.ServiceSpec{
			Ports:    app.Spec.Ports,
			Type:     corev1.ServiceTypeNodePort,
			Selector: labels,
		},
	}
}
