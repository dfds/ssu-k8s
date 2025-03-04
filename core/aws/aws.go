package aws

import (
	"context"
	"errors"
	"fmt"
	awsCore "github.com/aws/aws-sdk-go-v2/aws"
	awsHttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	orgTypes "github.com/aws/aws-sdk-go-v2/service/organizations/types"
	"go.dfds.cloud/ssu-k8s/core/logging"
	"go.uber.org/zap"
	"net/http"
)

type OrgClient struct {
	client *organizations.Client
	ctx    context.Context
}

func (c *OrgClient) GetAllAccountsFromOuRecursive(parentId string) ([]orgTypes.Account, error) {
	orgUnits, err := c.GetAllOUsFromParent(parentId)
	if err != nil {
		return nil, err
	}
	orgUnits = append(orgUnits, orgTypes.OrganizationalUnit{Id: &parentId})

	var allAccounts []orgTypes.Account

	for _, ou := range orgUnits {
		orgUnitAccounts, err := c.GetAccounts(*ou.Id)
		if err != nil {
			return nil, err
		}
		allAccounts = append(allAccounts, orgUnitAccounts...)
	}

	return allAccounts, nil
}

func (c *OrgClient) GetAccounts(parentId string) ([]orgTypes.Account, error) {
	var maxResults int32 = 20
	var accounts []orgTypes.Account
	resps := organizations.NewListAccountsForParentPaginator(c.client, &organizations.ListAccountsForParentInput{MaxResults: &maxResults, ParentId: &parentId})
	for resps.HasMorePages() { // Due to the limit of only 20 accounts per query and wanting to avoid getting hit by a rate limit, this will take a while if you have a decent amount of AWS accounts
		page, err := resps.NextPage(c.ctx)
		if err != nil {
			logging.Logger.Error("Error getting accounts:", zap.Error(err))
			return accounts, err
		}

		accounts = append(accounts, page.Accounts...)
	}

	return accounts, nil
}

func (c *OrgClient) GetAllOUsFromParent(parentId string) ([]orgTypes.OrganizationalUnit, error) {
	var maxResults int32 = 20
	var ou []orgTypes.OrganizationalUnit
	resp := organizations.NewListOrganizationalUnitsForParentPaginator(c.client, &organizations.ListOrganizationalUnitsForParentInput{
		ParentId:   &parentId,
		MaxResults: &maxResults,
	})

	for resp.HasMorePages() {
		page, err := resp.NextPage(c.ctx)
		if err != nil {
			return nil, err
		}

		ou = append(ou, page.OrganizationalUnits...)
	}

	for _, o := range ou {
		recursiveResp, err := c.GetAllOUsFromParent(*o.Id)
		if err != nil {
			return nil, err
		}
		ou = append(ou, recursiveResp...)
	}

	return ou, nil
}

func NewAwsOrgClient(ctx context.Context, region string) (*OrgClient, error) {
	var cfg awsCore.Config
	cfg, err := awsConfig.LoadDefaultConfig(ctx, awsConfig.WithRegion(region), awsConfig.WithHTTPClient(CreateHttpClientWithoutKeepAlive()))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("unable to load SDK config, %v", err))
	}
	orgClient := organizations.NewFromConfig(cfg)

	return &OrgClient{
		client: orgClient,
		ctx:    ctx,
	}, nil
}

func CreateHttpClientWithoutKeepAlive() *awsHttp.BuildableClient {
	client := awsHttp.NewBuildableClient().WithTransportOptions(func(transport *http.Transport) {
		transport.DisableKeepAlives = true
	})

	return client
}
