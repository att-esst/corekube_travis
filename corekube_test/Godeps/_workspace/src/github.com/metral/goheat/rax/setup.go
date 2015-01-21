package rax

import (
	"github.com/metral/goheat/util"
	"github.com/metral/goutils"
	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/openstack"
	"github.com/rackspace/gophercloud/openstack/identity/v2/tokens"
)

func IdentitySetup(c *util.HeatConfig) *tokens.Token {
	authOpts := gophercloud.AuthOptions{
		IdentityEndpoint: c.OSAuthUrl,
		Username:         c.OSUsername,
		Password:         c.OSPassword,
		TenantID:         c.OSTenantId,
	}

	provider, err := openstack.AuthenticatedClient(authOpts)
	goutils.PrintErrors(
		goutils.ErrorParams{Err: err, CallerNum: 2, Fatal: false})

	client := openstack.NewIdentityV2(provider)

	opts := tokens.WrapOptions(authOpts)
	token, err := tokens.Create(client, opts).ExtractToken()
	goutils.PrintErrors(
		goutils.ErrorParams{Err: err, CallerNum: 2, Fatal: false})

	return token
}
