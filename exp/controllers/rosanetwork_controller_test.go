/*
Copyright The Kubernetes Authors.
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
	"testing"

	awsSdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	cloudformationtypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	. "github.com/onsi/gomega"
	rosaAWSClient "github.com/openshift/rosa/pkg/aws"
	rosaMocks "github.com/openshift/rosa/pkg/aws/mocks"
	"github.com/sirupsen/logrus"
	gomock "go.uber.org/mock/gomock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"

	expinfrav1 "sigs.k8s.io/cluster-api-provider-aws/v2/exp/api/v1beta2"
)

func TestROSANetworkReconciler_Reconcile(t *testing.T) {
	g := NewWithT(t)

	rosaNetwork := &expinfrav1.ROSANetwork{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-rosa-network",
			Namespace: "test-namespace"},
		Spec: expinfrav1.ROSANetworkSpec{},
	}

	reconciler := &ROSANetworkReconciler{
		Client: testEnv.Client,
	}

	req := ctrl.Request{}
	req.NamespacedName = types.NamespacedName{Name: rosaNetwork.Name, Namespace: rosaNetwork.Namespace}
	_, errReconcile := reconciler.Reconcile(ctx, req)

	g.Expect(errReconcile).ToNot(HaveOccurred())
}

func TestROSANetworkReconciler_updateROSANetworkResources(t *testing.T) {
	g := NewWithT(t)
	mockCtrl := gomock.NewController(t)
	ctx := context.TODO()

	rosaNetwork := &expinfrav1.ROSANetwork{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-rosa-network",
			Namespace: "test-namespace",
		},
		Spec:   expinfrav1.ROSANetworkSpec{},
		Status: expinfrav1.ROSANetworkStatus{},
	}

	t.Run("Handle cloudformation client error", func(t *testing.T) {
		_, mockCFClient, reconciler := createMocks(mockCtrl)

		describeStackResourcesOutput := &cloudformation.DescribeStackResourcesOutput{}
		clientErr := fmt.Errorf("test-error")

		mockCFClient.EXPECT().DescribeStackResources(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, _ *cloudformation.DescribeStackResourcesInput, _ ...func(*cloudformation.Options)) (*cloudformation.DescribeStackResourcesOutput, error) {
				return describeStackResourcesOutput, clientErr
			}).Times(1)

		err := reconciler.updateROSANetworkResources(ctx, rosaNetwork)
		g.Expect(err).To(HaveOccurred())
		g.Expect(len(rosaNetwork.Status.Resources)).To(Equal(0))
	})

	t.Run("Update ROSANetwork.Status.Resources", func(t *testing.T) {
		_, mockCFClient, reconciler := createMocks(mockCtrl)

		logicalResourceID := "logical-resource-id"
		resourceStatus := cloudformationtypes.ResourceStatusCreateComplete
		resourceType := "resource-type"
		resourceStatusReason := "resource-status-reason"
		physicalResourceID := "physical-resource-id"

		describeStackResourcesOutput := &cloudformation.DescribeStackResourcesOutput{
			StackResources: []cloudformationtypes.StackResource{
				{
					LogicalResourceId:    &logicalResourceID,
					ResourceStatus:       resourceStatus,
					ResourceType:         &resourceType,
					ResourceStatusReason: &resourceStatusReason,
					PhysicalResourceId:   &physicalResourceID,
				},
			},
		}

		mockCFClient.EXPECT().DescribeStackResources(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, _ *cloudformation.DescribeStackResourcesInput, _ ...func(*cloudformation.Options)) (*cloudformation.DescribeStackResourcesOutput, error) {
				return describeStackResourcesOutput, nil
			}).Times(1)

		err := reconciler.updateROSANetworkResources(ctx, rosaNetwork)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(rosaNetwork.Status.Resources[0].LogicalID).To(Equal(logicalResourceID))
		g.Expect(rosaNetwork.Status.Resources[0].Status).To(Equal(string(resourceStatus)))
		g.Expect(rosaNetwork.Status.Resources[0].ResourceType).To(Equal(resourceType))
		g.Expect(rosaNetwork.Status.Resources[0].Reason).To(Equal(resourceStatusReason))
		g.Expect(rosaNetwork.Status.Resources[0].PhysicalID).To(Equal(physicalResourceID))
	})
}

func TestROSANetworkReconciler_parseSubnets(t *testing.T) {
	g := NewWithT(t)
	mockCtrl := gomock.NewController(t)

	subnet1Id := "subnet1-physical-id"
	subnet2Id := "subnet2-physical-id"

	rosaNetwork := &expinfrav1.ROSANetwork{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-rosa-network",
			Namespace: "test-namespace",
		},
		Spec: expinfrav1.ROSANetworkSpec{},
		Status: expinfrav1.ROSANetworkStatus{
			Resources: []expinfrav1.CFResource{
				{
					ResourceType: "AWS::EC2::Subnet",
					LogicalID:    "SubnetPrivate",
					PhysicalID:   subnet1Id,
					Status:       "subnet1-status",
					Reason:       "subnet1-reason",
				},
				{
					ResourceType: "AWS::EC2::Subnet",
					LogicalID:    "SubnetPublic",
					PhysicalID:   subnet2Id,
					Status:       "subnet2-status",
					Reason:       "subnet2-reason",
				},
				{
					ResourceType: "bogus-type",
					LogicalID:    "bogus-logical-id",
					PhysicalID:   "bugus-physical-id",
					Status:       "bogus-status",
					Reason:       "bogus-reason",
				},
			},
		},
	}

	t.Run("Handle EC2 client error", func(t *testing.T) {
		mockEC2Client, _, reconciler := createMocks(mockCtrl)

		describeSubnetsOutput := &ec2.DescribeSubnetsOutput{}

		mockEC2Client.EXPECT().DescribeSubnets(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, _ *ec2.DescribeSubnetsInput, _ ...func(*ec2.Options)) (*ec2.DescribeSubnetsOutput, error) {
				return describeSubnetsOutput, fmt.Errorf("test-error")
			}).Times(1)

		err := reconciler.parseSubnets(rosaNetwork)
		g.Expect(err).To(HaveOccurred())
		g.Expect(len(rosaNetwork.Status.Subnets)).To(Equal(0))
	})

	t.Run("Update ROSANetwork.Status.Subnets", func(t *testing.T) {
		mockEC2Client, _, reconciler := createMocks(mockCtrl)

		az := "az01"

		describeSubnetsOutput := &ec2.DescribeSubnetsOutput{
			Subnets: []ec2Types.Subnet{
				{
					AvailabilityZone: &az,
				},
			},
		}

		mockEC2Client.EXPECT().DescribeSubnets(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, _ *ec2.DescribeSubnetsInput, _ ...func(*ec2.Options)) (*ec2.DescribeSubnetsOutput, error) {
				return describeSubnetsOutput, nil
			}).Times(2)

		err := reconciler.parseSubnets(rosaNetwork)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(rosaNetwork.Status.Subnets[0].AvailabilityZone).To(Equal(az))
		g.Expect(rosaNetwork.Status.Subnets[0].PrivateSubnet).To(Equal(subnet1Id))
		g.Expect(rosaNetwork.Status.Subnets[0].PublicSubnet).To(Equal(subnet2Id))
	})
}

func createMocks(mockCtrl *gomock.Controller) (*rosaMocks.MockEc2ApiClient, *rosaMocks.MockCloudFormationApiClient, *ROSANetworkReconciler) {
	mockEC2Client := rosaMocks.NewMockEc2ApiClient(mockCtrl)
	mockCFClient := rosaMocks.NewMockCloudFormationApiClient(mockCtrl)
	awsClient := rosaAWSClient.New(
		awsSdk.Config{},
		rosaAWSClient.NewLoggerWrapper(logrus.New(), nil),
		rosaMocks.NewMockIamApiClient(mockCtrl),
		mockEC2Client,
		rosaMocks.NewMockOrganizationsApiClient(mockCtrl),
		rosaMocks.NewMockS3ApiClient(mockCtrl),
		rosaMocks.NewMockSecretsManagerApiClient(mockCtrl),
		rosaMocks.NewMockStsApiClient(mockCtrl),
		mockCFClient,
		rosaMocks.NewMockServiceQuotasApiClient(mockCtrl),
		rosaMocks.NewMockServiceQuotasApiClient(mockCtrl),
		&rosaAWSClient.AccessKey{},
		false,
	)

	reconciler := &ROSANetworkReconciler{
		Client:    testEnv.Client,
		awsClient: awsClient,
	}

	return mockEC2Client, mockCFClient, reconciler
}
