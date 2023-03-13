package eks

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

var ClusterClient = TestClient{
	DescribeClusterOutput: &eks.DescribeClusterOutput{
		Cluster: &types.Cluster{
			Name:               sources.PtrString("dylan"),
			Arn:                sources.PtrString("arn:aws:eks:eu-west-2:801795385023:cluster/dylan"),
			CreatedAt:          sources.PtrTime(time.Now()),
			Version:            sources.PtrString("1.24"),
			Endpoint:           sources.PtrString("https://00D3FF4CC48CBAA9BBC070DAA80BD251.gr7.eu-west-2.eks.amazonaws.com"),
			RoleArn:            sources.PtrString("arn:aws:iam::801795385023:role/dylan-cluster-20221222134106992100000001"),
			ClientRequestToken: sources.PtrString("token"),
			ConnectorConfig: &types.ConnectorConfigResponse{
				ActivationCode:   sources.PtrString("code"),
				ActivationExpiry: sources.PtrTime(time.Now()),
				ActivationId:     sources.PtrString("id"),
				Provider:         sources.PtrString("provider"),
				RoleArn:          sources.PtrString("arn:aws:iam::801795385023:role/dylan-cluster-20221222134106992100000002"),
			},
			Health: &types.ClusterHealth{
				Issues: []types.ClusterIssue{},
			},
			Id: sources.PtrString("id"),
			OutpostConfig: &types.OutpostConfigResponse{
				ControlPlaneInstanceType: sources.PtrString("type"),
				OutpostArns: []string{
					"arn1",
				},
				ControlPlanePlacement: &types.ControlPlanePlacementResponse{
					GroupName: sources.PtrString("groupName"),
				},
			},
			ResourcesVpcConfig: &types.VpcConfigResponse{
				SubnetIds: []string{
					"subnet-0d1fabfe6794b5543",
					"subnet-0865943940092d10a",
					"subnet-00ed8275954eca233",
				},
				SecurityGroupIds: []string{
					"sg-0bf38eb7e14777399",
				},
				ClusterSecurityGroupId: sources.PtrString("sg-08df96f08566d4dda"),
				VpcId:                  sources.PtrString("vpc-0c9152ce7ed2b7305"),
				EndpointPublicAccess:   true,
				EndpointPrivateAccess:  true,
				PublicAccessCidrs: []string{
					"0.0.0.0/0",
				},
			},
			KubernetesNetworkConfig: &types.KubernetesNetworkConfigResponse{
				ServiceIpv4Cidr: sources.PtrString("172.20.0.0/16"),
				IpFamily:        types.IpFamilyIpv4,
				ServiceIpv6Cidr: sources.PtrString("ipv6cidr"),
			},
			Logging: &types.Logging{
				ClusterLogging: []types.LogSetup{
					{
						Types: []types.LogType{
							"api",
							"authenticator",
							"controllerManager",
							"scheduler",
						},
						Enabled: sources.PtrBool(true),
					},
					{
						Types: []types.LogType{
							"audit",
						},
						Enabled: sources.PtrBool(false),
					},
				},
			},
			Identity: &types.Identity{
				Oidc: &types.OIDC{
					Issuer: sources.PtrString("https://oidc.eks.eu-west-2.amazonaws.com/id/00D3FF4CC48CBAA9BBC070DAA80BD251"),
				},
			},
			Status: types.ClusterStatusActive,
			CertificateAuthority: &types.Certificate{
				Data: sources.PtrString("LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUMvakNDQWVhZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFWTVJNd0VRWURWUVFERXdwcmRXSmwKY201bGRHVnpNQjRYRFRJeU1USXlNakV6TkRZME5Gb1hEVE15TVRJeE9URXpORFkwTkZvd0ZURVRNQkVHQTFVRQpBeE1LYTNWaVpYSnVaWFJsY3pDQ0FTSXdEUVlKS29aSWh2Y05BUUVCQlFBRGdnRVBBRENDQVFvQ2dnRUJBTC9tCkN6b25QdUZIUXM1a0xudzdCeXMrak9pNWJscEVCN2RhZUYvQzZqaEVTbkcwdVBVRjVWSFUzbmRyZHRKelBaemQKenM4U1pEMzRsKytGWmw0NFQrYWRqMGFYanpmZ0NTeFo4K0MvaWJUOWIzck5jWU9ZZ3FYT1lXc2JVYmpBSjRadgpnakFqdEl3dTBvUHNYT0JSZU5KTDlhRkl6VFFIcy9QL1hONWI5eGRlSHhwOXN4cnlEREYxQVNuQkxwajduUHMrCmgyNUtvd0hQV1luekV6WVd1T3NZbDQ2RjZacHh4aVhya2hnOGozckR4dXRWZGMvQVBFaVhUdHh3OU9CMjFDMkwKK1VpanpxS2RrZm5idVEvOHF0TTRqbFVGTkgzUG03STlkTEdIMTBTOFdhQkhpODNRMklCd3c0eE5RZ04xNC91dgpXWFZOWkxmM1EwbElkdmtxaCtrQ0F3RUFBYU5aTUZjd0RnWURWUjBQQVFIL0JBUURBZ0trTUE4R0ExVWRFd0VCCi93UUZNQU1CQWY4d0hRWURWUjBPQkJZRUZCa2wvVEJwNVNyMFJrVEk2V1dMVkR4MVdZYUxNQlVHQTFVZEVRUU8KTUF5Q0NtdDFZbVZ5Ym1WMFpYTXdEUVlKS29aSWh2Y05BUUVMQlFBRGdnRUJBQ0FCVWtZUWZSQXlRRFVsc2todgp2NTRZN3lFQ1lUSG00OWVtMWoyV2hyN0JPdXdlUkU4M3g1b0NhWEtjK2tMemlvOEVvY2hxOWN1a1FEYm1KNkpoCmRhUUlyaFFwaG5PMHZSd290YXlhWjdlV2IwTm50WmNxN1ZmNkp5ZU5CR3Y1NTJGdlNNcGprWnh0UXVpTTJ5TXoKbjJWWmtxMzJPb0RjTmxCMERhRVBCSjlIM2ZnbG1qcGdWL0NHZFdMNG1wNEpkb3VPNTFtNkJBMm1ET2JWYzh4VgppNFJIWE9KNG9hSGFTd1B6MHBuQUxabkJoUnpxV0Q1cGlycVlucjBxSlFDamJDWXF1TmJTU3d4c2JMYVFjanNFCjhiUXk0aGxXaEJNWno3UldOeDg1UTBZSjhWNEhKdXVCZ09MaVg1REFtNDZIbndWUy95MHJyN2JTWThoTXErM2QKTmtrPQotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg=="),
			},
			PlatformVersion: sources.PtrString("eks.3"),
			Tags:            map[string]string{},
			EncryptionConfig: []types.EncryptionConfig{
				{
					Resources: []string{
						"secrets",
					},
					Provider: &types.Provider{
						KeyArn: sources.PtrString("arn:aws:kms:eu-west-2:801795385023:key/3a478539-9717-4c20-83a5-19989154dc32"),
					},
				},
			},
		},
	},
}

func TestClusterGetFunc(t *testing.T) {
	item, err := ClusterGetFunc(context.Background(), ClusterClient, "foo", &eks.DescribeClusterInput{})

	if err != nil {
		t.Error(err)
	}

	if err = item.Validate(); err != nil {
		t.Error(err)
	}

	// It doesn't really make sense to test anything other than the linked items
	// since the attributes are converted automatically
	tests := sources.ItemRequestTests{
		{
			ExpectedType:   "iam-role",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:aws:iam::801795385023:role/dylan-cluster-20221222134106992100000002",
			ExpectedScope:  "801795385023",
		},
		{
			ExpectedType:   "kms-key",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:aws:kms:eu-west-2:801795385023:key/3a478539-9717-4c20-83a5-19989154dc32",
			ExpectedScope:  "801795385023.eu-west-2",
		},
		{
			ExpectedType:   "http",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "https://00D3FF4CC48CBAA9BBC070DAA80BD251.gr7.eu-west-2.eks.amazonaws.com",
			ExpectedScope:  "global",
		},
		{
			ExpectedType:   "ec2-security-group",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "sg-0bf38eb7e14777399",
			ExpectedScope:  item.Scope,
		},
		{
			ExpectedType:   "ec2-security-group",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "sg-08df96f08566d4dda",
			ExpectedScope:  item.Scope,
		},
		{
			ExpectedType:   "ec2-subnet",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "subnet-0d1fabfe6794b5543",
			ExpectedScope:  item.Scope,
		},
		{
			ExpectedType:   "ec2-subnet",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "subnet-0865943940092d10a",
			ExpectedScope:  item.Scope,
		},
		{
			ExpectedType:   "ec2-subnet",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "subnet-00ed8275954eca233",
			ExpectedScope:  item.Scope,
		},
		{
			ExpectedType:   "ec2-vpc",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "vpc-0c9152ce7ed2b7305",
			ExpectedScope:  item.Scope,
		},
		{
			ExpectedType:   "iam-role",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:aws:iam::801795385023:role/dylan-cluster-20221222134106992100000001",
			ExpectedScope:  "801795385023",
		},
		{
			ExpectedType:   "eks-fargate-profile",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "dylan",
			ExpectedScope:  item.Scope,
		},
		{
			ExpectedType:   "eks-addon",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "dylan",
			ExpectedScope:  item.Scope,
		},
		{
			ExpectedType:   "eks-nodegroup",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "dylan",
			ExpectedScope:  item.Scope,
		},
	}

	tests.Execute(t, item)
}

func TestNewClusterSource(t *testing.T) {
	config, account, region := sources.GetAutoConfig(t)

	source := NewClusterSource(config, account, region)

	test := sources.E2ETest{
		Source:  source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}
