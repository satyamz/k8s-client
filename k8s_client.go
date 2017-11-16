package main

import (
	"flag"
	"fmt"

	appsv1beta1 "k8s.io/api/apps/v1beta1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	deploymentName    = flag.String("deploy", "", "Deployment name")
	slackToken        = flag.String("slack-token", "", "Slack Token")
	verificationToken = flag.String("verify-token", "", "Slack Verification Token")
	mayaServerIP      = flag.String("maya-server-ip", "", "Maya server IP with port (e.g. http://192.168.0.0:8000)")
	slackWebHook      = flag.String("slack-web-hook", "", "Slack incoming web hook")
	apiKey            = flag.String("api-key", "", "API Key to authenticate with Maya Server")
	externalIP        = flag.String("external-ip", "", "External IP for chat server")
)

type DeploymentSpec struct {
	DeploymentName    string
	SlackToken        string
	VerificationToken string
	MayaServerIP      string
	SlackWebHook      string
	APIKey            string
	ExternalIP        string
}

func NewDeploymentSpec(DeploymentName, SlackToken, VerificationToken, MayaServerIP, SlackWebHook, APIKey, ExternalIP string) *DeploymentSpec {
	return &DeploymentSpec{
		DeploymentName:    DeploymentName,
		SlackToken:        SlackToken,
		VerificationToken: VerificationToken,
		MayaServerIP:      MayaServerIP,
		SlackWebHook:      SlackWebHook,
		APIKey:            APIKey,
		ExternalIP:        ExternalIP,
	}

}

func NewDeploymentInstance(deploySpec *DeploymentSpec) *appsv1beta1.Deployment {
	return &appsv1beta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: deploySpec.DeploymentName,
		},
		Spec: appsv1beta1.DeploymentSpec{
			Replicas: int32ptr(1),
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "chat-server",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "chat-server",
							Image: "mulebot/chat-server:v03",
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 8000,
								},
							},
							Env: []apiv1.EnvVar{
								{
									Name:  "SLACK_TOKEN",
									Value: deploySpec.SlackToken,
								},
								{
									Name:  "VERIFICATION_TOKEN",
									Value: deploySpec.VerificationToken,
								},
								{
									Name:  "MAYA_SERVER_IP",
									Value: deploySpec.MayaServerIP,
								},
								{
									Name:  "SLACK_INCOMING_WEB_HOOK",
									Value: deploySpec.SlackWebHook,
								},
								{
									Name:  "API_KEY",
									Value: deploySpec.APIKey,
								},
							},
						},
					},
				},
			},
		},
	}

}

func NewSeviceInstance(deploySpec *DeploymentSpec) *apiv1.Service {
	return &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"app": "chat-server",
			},
			Name: deploySpec.DeploymentName + "-service",
		},
		Spec: apiv1.ServiceSpec{
			Type: apiv1.ServiceTypeNodePort,
			Ports: []apiv1.ServicePort{
				{
					Name:     "chatserver-port",
					Port:     8000, //TODO: Need to change. Accept env var as of now.
					NodePort: 30550,
					Protocol: apiv1.ProtocolTCP,
				},
			},
			ExternalIPs: []string{
				deploySpec.ExternalIP,
			},
			Selector: map[string]string{
				"app": "chat-server",
			},
		},
	}
}

func CreateDeployment(client kubernetes.Interface, deploySpec *DeploymentSpec) {

	deployClient := client.AppsV1beta1().Deployments(apiv1.NamespaceDefault)
	deploymentInstance := NewDeploymentInstance(deploySpec)
	deployRes, err := deployClient.Create(deploymentInstance)
	if err != nil {
		fmt.Printf("[Error]: %+v", err)
		return
	}

	fmt.Printf("Deployment details:\n %+v\n", deployRes)
}

func CreateService(client kubernetes.Interface, deploySpec *DeploymentSpec) {
	serviceClient := client.Core().Services(apiv1.NamespaceDefault)
	serviceInstance := NewSeviceInstance(deploySpec)
	serveiceRes, err := serviceClient.Create(serviceInstance)
	if err != nil {
		fmt.Printf("[Error]: %+v", err)
		return
	}

	fmt.Printf("Service details \n %+v \n", serveiceRes)
}

func main() {
	flag.Parse()
	DeploymentSpecObj := NewDeploymentSpec(*deploymentName, *slackToken, *verificationToken, *mayaServerIP, *slackWebHook, *apiKey, *externalIP)

	config, err := rest.InClusterConfig()
	if err != nil {
		fmt.Println("Failed to get config  ", err)
	}
	fmt.Println(config)

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Println(" Failed to create config ", err)
	}

	CreateDeployment(clientset, DeploymentSpecObj)
	CreateService(clientset, DeploymentSpecObj)
}

func int32ptr(number int32) *int32 {
	return &number
}
