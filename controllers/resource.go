package controllers

import (
	"github.com/best/k8s-opdemo/api/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func MutateDeployment(app *v1beta1.AppService, deploy *appsv1.Deployment) {
	labels := map[string]string{
		"app": app.Name,
	}
	selector := &metav1.LabelSelector{
		MatchLabels: labels,
	}

	deploy.Spec = appsv1.DeploymentSpec{
		Replicas: app.Spec.Size,
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: labels,
			},
			Spec: corev1.PodSpec{
				Containers: newContainers(app),
			},
		},
		Selector: selector,
	}
}

func MutateService(app *v1beta1.AppService, svc *corev1.Service) {
	labels := map[string]string{
		"app": app.Name,
	}
	svc.Spec = corev1.ServiceSpec{
		Ports:    app.Spec.Ports,
		Type:     corev1.ServiceTypeNodePort,
		Selector: labels,
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
