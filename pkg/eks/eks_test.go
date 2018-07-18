package eks_test

import (
	"github.com/aws/aws-sdk-go/aws"
	cfn "github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/kubicorn/kubicorn/pkg/logger"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	. "github.com/weaveworks/eksctl/pkg/eks"
	"github.com/weaveworks/eksctl/pkg/eks/mocks"
)

var _ = Describe("Eks", func() {
	var (
		c *ClusterProvider
		p *MockProvider
	)

	BeforeEach(func() {

	})

	Describe("ListAll", func() {
		Context("With a cluster name", func() {
			var (
				clusterName string
				err         error
			)

			BeforeEach(func() {
				clusterName = "test-cluster"

				p = &MockProvider{
					cfn: &mocks.CloudFormationAPI{},
					eks: &mocks.EKSAPI{},
					ec2: &mocks.EC2API{},
					sts: &mocks.STSAPI{},
				}

				c = &ClusterProvider{
					Spec: &ClusterConfig{
						ClusterName: clusterName,
					},
					Provider: p,
				}

				p.mockEKS().On("DescribeCluster", mock.MatchedBy(func(input *eks.DescribeClusterInput) bool {
					return *input.Name == clusterName
				})).Return(&eks.DescribeClusterOutput{
					Cluster: &eks.Cluster{
						Name:   aws.String(clusterName),
						Status: aws.String(eks.ClusterStatusActive),
					},
				}, nil)
			})

			Context("and normal log level", func() {
				BeforeEach(func() {
					logger.Level = 3
				})

				JustBeforeEach(func() {
					err = c.ListClusters()
				})

				It("should not error", func() {
					Expect(err).NotTo(HaveOccurred())
				})

				It("should have called AWS EKS service once", func() {
					Expect(p.mockEKS().AssertNumberOfCalls(GinkgoT(), "DescribeCluster", 1)).To(BeTrue())
				})

				It("should not call AWS CFN ListStackPages", func() {
					Expect(p.mockCloudFormation().AssertNumberOfCalls(GinkgoT(), "ListStacksPages", 0)).To(BeTrue())
				})
			})

			Context("and debug log level", func() {
				var (
					expectedStatusFilter string
				)
				BeforeEach(func() {
					expectedStatusFilter = "CREATE_COMPLETE"

					logger.Level = 4

					p.mockCloudFormation().On("ListStacksPages", mock.MatchedBy(func(input *cfn.ListStacksInput) bool {
						return *input.StackStatusFilter[0] == expectedStatusFilter
					}), mock.Anything).Return(nil)
				})

				JustBeforeEach(func() {
					err = c.ListClusters()
				})

				It("should not error", func() {
					Expect(err).NotTo(HaveOccurred())
				})

				It("should have called AWS EKS service once", func() {
					Expect(p.mockEKS().AssertNumberOfCalls(GinkgoT(), "DescribeCluster", 1)).To(BeTrue())
				})

				It("should have called AWS CFN ListStackPages", func() {
					Expect(p.mockCloudFormation().AssertNumberOfCalls(GinkgoT(), "ListStacksPages", 1)).To(BeTrue())
				})
			})
		})
	})
})