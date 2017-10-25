package comms

import "github.com/ericchiang/k8s/apis/extensions/v1beta1"

type MockK8sCommunicator struct {
	OnGetDeployment    func(string) (*v1beta1.Deployment, error)
	OnListDeployments  func() (*v1beta1.DeploymentList, error)
	OnUpdateDeployment func(*v1beta1.Deployment) (*v1beta1.Deployment, error)
}

func (mock *MockK8sCommunicator) GetDeployment(deploymentName string) (*v1beta1.Deployment, error) {
	return mock.OnGetDeployment(deploymentName)
}

func (mock *MockK8sCommunicator) ListDeployments() (*v1beta1.DeploymentList, error) {
	return mock.OnListDeployments()
}

func (mock *MockK8sCommunicator) UpdateDeployment(deployment *v1beta1.Deployment) (*v1beta1.Deployment, error) {
	return mock.OnUpdateDeployment(deployment)
}
