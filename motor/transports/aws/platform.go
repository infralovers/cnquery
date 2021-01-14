package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/iam"
)

func (t *Transport) Identifier() (string, error) {
	return "//platformid.api.mondoo.app/runtime/aws/accounts/" + t.info.Account, nil
}

// Info returns the connection information
func (t *Transport) Info() Info {
	return t.info
}

type Account struct {
	ID      string
	Aliases []string
}

func (t *Transport) Account() (Account, error) {
	accountid := t.info.Account
	ctx := context.Background()
	res, err := t.Iam("").ListAccountAliasesRequest(&iam.ListAccountAliasesInput{}).Send(ctx)
	if err != nil {
		return Account{
			ID: accountid,
		}, err
	}
	return Account{
		ID:      accountid,
		Aliases: res.AccountAliases,
	}, nil
}
