package apiservice

import (
	"context"
	"encoding/json"

	"connectrpc.com/connect"
	ssoreadyv1 "github.com/ssoready/ssoready/internal/gen/ssoready/v1"
	"github.com/ssoready/ssoready/internal/store"
)

func (s *Service) RedeemSAMLAccessToken(ctx context.Context, req *connect.Request[ssoreadyv1.RedeemSAMLAccessTokenRequest]) (*connect.Response[ssoreadyv1.RedeemSAMLAccessTokenResponse], error) {
	samlSess, err := s.Store.GetSAMLSessionBySecretAccessToken(ctx, &store.GetSAMLSessionBySecretAccessTokenRequest{
		SecretAccessToken: req.Msg.AccessToken,
	})
	if err != nil {
		return nil, err
	}

	var attrs map[string]string
	if err := json.Unmarshal(samlSess.SubjectIdpAttributes, &attrs); err != nil {
		return nil, err
	}

	return connect.NewResponse(&ssoreadyv1.RedeemSAMLAccessTokenResponse{
		SubjectIdpId:         *samlSess.SubjectID,
		SubjectIdpAttributes: attrs,
	}), nil
}
