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
)

type DeploymentSpec struct {
	DeploymentName    string
	SlackToken        string
	VerificationToken string
	MayaServerIP      string
	SlackWebHook      string
	APIKey            string
}

func NewDeploymentSpec(DeploymentName, SlackToken, VerificationToken, MayaServerIP, SlackWebHook, APIKey string) *DeploymentSpec {
	return &DeploymentSpec{
		DeploymentName:    DeploymentName,
		SlackToken:        SlackToken,
		VerificationToken: VerificationToken,
		MayaServerIP:      MayaServerIP,
		SlackWebHook:      SlackWebHook,
		APIKey:            APIKey,
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
							Name:  "Chat Server",
							Image: "mulebot/chatserver:v02",
							Ports: []apiv1.ContainerPort{
								{
									Name:          "chat-server-http-port",
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

func CreateDeployment(client kubernetes.Interface, deploySpec *DeploymentSpec) {

	deployClient := client.AppsV1beta1().Deployments(apiv1.NamespaceDefault)
	deploymentInstance := NewDeploymentInstance(deploySpec)
	deployClient.Create(deploymentInstance)
	fmt.Printf("Deployment %s is created!\n", deploySpec.DeploymentName)
}

func main() {
	flag.Parse()
	DeploymentSpecObj := NewDeploymentSpec(*deploymentName, *slackToken, *verificationToken, *mayaServerIP, *slackWebHook, *apiKey)

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

}

func int32ptr(number int32) *int32 {
	return &number
}
