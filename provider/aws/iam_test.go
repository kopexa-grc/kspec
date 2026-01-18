// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package aws

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"go.uber.org/mock/gomock"

	"github.com/kopexa-grc/kspec/core"
	"github.com/kopexa-grc/kspec/provider/aws/resources"
)

func TestIAMUserResource_Name(t *testing.T) {
	r := resources.NewIAMUser(nil, "")
	if got := r.Name(); got != "aws_iam_user" {
		t.Errorf("Name() = %v, want aws_iam_user", got)
	}
}

func TestIAMUserResource_Fetch_EmptyUsers(t *testing.T) {
	ctrl := gomock.NewController(t)
	mock := NewMockIAMAPI(ctrl)

	mock.EXPECT().
		ListUsers(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&iam.ListUsersOutput{Users: []types.User{}}, nil)

	r := resources.NewIAMUser(mock, "123456789012")
	rs, err := r.Fetch(context.Background(), core.Asset{})

	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if len(rs) != 0 {
		t.Errorf("Fetch() returned %d resources, want 0", len(rs))
	}
}

func TestIAMUserResource_Fetch_SingleUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	mock := NewMockIAMAPI(ctrl)

	createDate := time.Now().Add(-30 * 24 * time.Hour)
	userName := "test-user"
	userID := "AIDATEST123"
	userArn := "arn:aws:iam::123456789012:user/test-user"

	mock.EXPECT().
		ListUsers(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&iam.ListUsersOutput{
			Users: []types.User{
				{
					UserName:   aws.String(userName),
					UserId:     aws.String(userID),
					Arn:        aws.String(userArn),
					Path:       aws.String("/"),
					CreateDate: &createDate,
				},
			},
		}, nil)

	// No console access
	mock.EXPECT().
		GetLoginProfile(gomock.Any(), gomock.Any()).
		Return(nil, &types.NoSuchEntityException{})

	// No MFA
	mock.EXPECT().
		ListMFADevices(gomock.Any(), gomock.Any()).
		Return(&iam.ListMFADevicesOutput{MFADevices: []types.MFADevice{}}, nil)

	// No access keys
	mock.EXPECT().
		ListAccessKeys(gomock.Any(), gomock.Any()).
		Return(&iam.ListAccessKeysOutput{AccessKeyMetadata: []types.AccessKeyMetadata{}}, nil)

	// No groups
	mock.EXPECT().
		ListGroupsForUser(gomock.Any(), gomock.Any()).
		Return(&iam.ListGroupsForUserOutput{Groups: []types.Group{}}, nil)

	// No attached policies
	mock.EXPECT().
		ListAttachedUserPolicies(gomock.Any(), gomock.Any()).
		Return(&iam.ListAttachedUserPoliciesOutput{AttachedPolicies: []types.AttachedPolicy{}}, nil)

	// No inline policies
	mock.EXPECT().
		ListUserPolicies(gomock.Any(), gomock.Any()).
		Return(&iam.ListUserPoliciesOutput{PolicyNames: []string{}}, nil)

	r := resources.NewIAMUser(mock, "123456789012")
	rs, err := r.Fetch(context.Background(), core.Asset{})

	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if len(rs) != 1 {
		t.Fatalf("Fetch() returned %d resources, want 1", len(rs))
	}

	res := rs[0]
	if res["name"] != userName {
		t.Errorf("name = %v, want %v", res["name"], userName)
	}
	if res["id"] != userID {
		t.Errorf("id = %v, want %v", res["id"], userID)
	}
	if res["has_mfa"] != false {
		t.Errorf("has_mfa = %v, want false", res["has_mfa"])
	}
	if res["has_console_access"] != false {
		t.Errorf("has_console_access = %v, want false", res["has_console_access"])
	}
}

func TestIAMUserResource_Fetch_UserWithMFA(t *testing.T) {
	ctrl := gomock.NewController(t)
	mock := NewMockIAMAPI(ctrl)

	userName := "secure-user"
	userID := "AIDATEST456"

	mock.EXPECT().
		ListUsers(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&iam.ListUsersOutput{
			Users: []types.User{
				{
					UserName: aws.String(userName),
					UserId:   aws.String(userID),
					Arn:      aws.String("arn:aws:iam::123456789012:user/secure-user"),
					Path:     aws.String("/"),
				},
			},
		}, nil)

	// Has console access
	mock.EXPECT().
		GetLoginProfile(gomock.Any(), gomock.Any()).
		Return(&iam.GetLoginProfileOutput{}, nil)

	// Has MFA
	mock.EXPECT().
		ListMFADevices(gomock.Any(), gomock.Any()).
		Return(&iam.ListMFADevicesOutput{
			MFADevices: []types.MFADevice{
				{SerialNumber: aws.String("arn:aws:iam::123456789012:mfa/secure-user")},
			},
		}, nil)

	mock.EXPECT().
		ListAccessKeys(gomock.Any(), gomock.Any()).
		Return(&iam.ListAccessKeysOutput{AccessKeyMetadata: []types.AccessKeyMetadata{}}, nil)

	mock.EXPECT().
		ListGroupsForUser(gomock.Any(), gomock.Any()).
		Return(&iam.ListGroupsForUserOutput{Groups: []types.Group{}}, nil)

	mock.EXPECT().
		ListAttachedUserPolicies(gomock.Any(), gomock.Any()).
		Return(&iam.ListAttachedUserPoliciesOutput{AttachedPolicies: []types.AttachedPolicy{}}, nil)

	mock.EXPECT().
		ListUserPolicies(gomock.Any(), gomock.Any()).
		Return(&iam.ListUserPoliciesOutput{PolicyNames: []string{}}, nil)

	r := resources.NewIAMUser(mock, "123456789012")
	rs, err := r.Fetch(context.Background(), core.Asset{})

	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if len(rs) != 1 {
		t.Fatalf("Fetch() returned %d resources, want 1", len(rs))
	}

	res := rs[0]
	if res["has_mfa"] != true {
		t.Errorf("has_mfa = %v, want true", res["has_mfa"])
	}
	if res["has_console_access"] != true {
		t.Errorf("has_console_access = %v, want true", res["has_console_access"])
	}
	if res["console_without_mfa"] != false {
		t.Errorf("console_without_mfa = %v, want false (has MFA)", res["console_without_mfa"])
	}
}

func TestIAMUserResource_Fetch_AdminUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	mock := NewMockIAMAPI(ctrl)

	userName := "admin-user"

	mock.EXPECT().
		ListUsers(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&iam.ListUsersOutput{
			Users: []types.User{
				{
					UserName: aws.String(userName),
					UserId:   aws.String("AIDAADMIN"),
					Arn:      aws.String("arn:aws:iam::123456789012:user/admin-user"),
					Path:     aws.String("/"),
				},
			},
		}, nil)

	mock.EXPECT().GetLoginProfile(gomock.Any(), gomock.Any()).Return(nil, &types.NoSuchEntityException{})
	mock.EXPECT().ListMFADevices(gomock.Any(), gomock.Any()).Return(&iam.ListMFADevicesOutput{}, nil)
	mock.EXPECT().ListAccessKeys(gomock.Any(), gomock.Any()).Return(&iam.ListAccessKeysOutput{}, nil)
	mock.EXPECT().ListGroupsForUser(gomock.Any(), gomock.Any()).Return(&iam.ListGroupsForUserOutput{}, nil)

	// Admin policy attached
	mock.EXPECT().
		ListAttachedUserPolicies(gomock.Any(), gomock.Any()).
		Return(&iam.ListAttachedUserPoliciesOutput{
			AttachedPolicies: []types.AttachedPolicy{
				{PolicyArn: aws.String("arn:aws:iam::aws:policy/AdministratorAccess")},
			},
		}, nil)

	mock.EXPECT().ListUserPolicies(gomock.Any(), gomock.Any()).Return(&iam.ListUserPoliciesOutput{}, nil)

	r := resources.NewIAMUser(mock, "123456789012")
	rs, err := r.Fetch(context.Background(), core.Asset{})

	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}

	res := rs[0]
	if res["is_admin"] != true {
		t.Errorf("is_admin = %v, want true", res["is_admin"])
	}
}

func TestIAMRoleResource_Name(t *testing.T) {
	r := resources.NewIAMRole(nil, "")
	if got := r.Name(); got != "aws_iam_role" {
		t.Errorf("Name() = %v, want aws_iam_role", got)
	}
}

func TestIAMRoleResource_Fetch_SingleRole(t *testing.T) {
	ctrl := gomock.NewController(t)
	mock := NewMockIAMAPI(ctrl)

	roleName := "test-role"
	assumeRolePolicy := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"Service":"ec2.amazonaws.com"},"Action":"sts:AssumeRole"}]}`

	mock.EXPECT().
		ListRoles(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&iam.ListRolesOutput{
			Roles: []types.Role{
				{
					RoleName:                 aws.String(roleName),
					RoleId:                   aws.String("AROATEST123"),
					Arn:                      aws.String("arn:aws:iam::123456789012:role/test-role"),
					Path:                     aws.String("/"),
					AssumeRolePolicyDocument: aws.String(assumeRolePolicy),
				},
			},
		}, nil)

	mock.EXPECT().
		ListAttachedRolePolicies(gomock.Any(), gomock.Any()).
		Return(&iam.ListAttachedRolePoliciesOutput{AttachedPolicies: []types.AttachedPolicy{}}, nil)

	mock.EXPECT().
		ListRolePolicies(gomock.Any(), gomock.Any()).
		Return(&iam.ListRolePoliciesOutput{PolicyNames: []string{}}, nil)

	r := resources.NewIAMRole(mock, "123456789012")
	rs, err := r.Fetch(context.Background(), core.Asset{})

	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if len(rs) != 1 {
		t.Fatalf("Fetch() returned %d resources, want 1", len(rs))
	}

	res := rs[0]
	if res["name"] != roleName {
		t.Errorf("name = %v, want %v", res["name"], roleName)
	}
	if res["is_service_role"] != false {
		t.Errorf("is_service_role = %v, want false", res["is_service_role"])
	}
}

func TestIAMGroupResource_Name(t *testing.T) {
	r := resources.NewIAMGroup(nil)
	if got := r.Name(); got != "aws_iam_group" {
		t.Errorf("Name() = %v, want aws_iam_group", got)
	}
}

func TestIAMGroupResource_Fetch_SingleGroup(t *testing.T) {
	ctrl := gomock.NewController(t)
	mock := NewMockIAMAPI(ctrl)

	groupName := "developers"

	mock.EXPECT().
		ListGroups(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&iam.ListGroupsOutput{
			Groups: []types.Group{
				{
					GroupName: aws.String(groupName),
					GroupId:   aws.String("AGPATEST123"),
					Arn:       aws.String("arn:aws:iam::123456789012:group/developers"),
					Path:      aws.String("/"),
				},
			},
		}, nil)

	mock.EXPECT().
		GetGroup(gomock.Any(), gomock.Any()).
		Return(&iam.GetGroupOutput{
			Users: []types.User{
				{UserName: aws.String("user1")},
				{UserName: aws.String("user2")},
			},
		}, nil)

	mock.EXPECT().
		ListAttachedGroupPolicies(gomock.Any(), gomock.Any()).
		Return(&iam.ListAttachedGroupPoliciesOutput{AttachedPolicies: []types.AttachedPolicy{}}, nil)

	mock.EXPECT().
		ListGroupPolicies(gomock.Any(), gomock.Any()).
		Return(&iam.ListGroupPoliciesOutput{PolicyNames: []string{}}, nil)

	r := resources.NewIAMGroup(mock)
	rs, err := r.Fetch(context.Background(), core.Asset{})

	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if len(rs) != 1 {
		t.Fatalf("Fetch() returned %d resources, want 1", len(rs))
	}

	res := rs[0]
	if res["name"] != groupName {
		t.Errorf("name = %v, want %v", res["name"], groupName)
	}
	if res["member_count"] != 2 {
		t.Errorf("member_count = %v, want 2", res["member_count"])
	}
	if res["has_members"] != true {
		t.Errorf("has_members = %v, want true", res["has_members"])
	}
}

func TestIAMPolicyResource_Name(t *testing.T) {
	r := resources.NewIAMPolicy(nil)
	if got := r.Name(); got != "aws_iam_policy" {
		t.Errorf("Name() = %v, want aws_iam_policy", got)
	}
}
