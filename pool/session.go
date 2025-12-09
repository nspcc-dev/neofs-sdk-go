package pool

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neofs-sdk-go/client"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/session/v2"
	"github.com/nspcc-dev/neofs-sdk-go/user"
)

func initSession(ctx context.Context, c *sdkClientWrapper, dur uint64, signer user.Signer) (session.Token, error) {
	currentTime := uint64(time.Now().Unix())

	var exp uint64
	if math.MaxUint64-currentTime < dur {
		exp = math.MaxUint64
	} else {
		exp = currentTime + dur
	}

	var dst session.Token
	dst.SetVersion(session.TokenCurrentVersion)
	dst.SetExp(exp)
	dst.SetIssuer(session.NewTargetUser(signer.UserID()))

	// set random nonce
	dst.SetNonce(session.NewNonce())

	nm, err := c.NetMapSnapshot(ctx, client.PrmNetMapSnapshot{})
	if err != nil {
		return dst, fmt.Errorf("get netmap snapshot: %w", err)
	}
	for _, node := range nm.Nodes() {
		neoPubKey, err := keys.NewPublicKeyFromBytes(node.PublicKey(), elliptic.P256())
		if err != nil {
			return dst, fmt.Errorf("parse node public key: %w", err)
		}

		ecdsaPubKey := (*ecdsa.PublicKey)(neoPubKey)

		userID := user.NewFromECDSAPublicKey(*ecdsaPubKey)
		if err = dst.AddSubject(session.NewTargetUser(userID)); err != nil {
			return dst, fmt.Errorf("add subject: %w", err)
		}
	}

	dst.SetIat(currentTime)
	dst.SetNbf(currentTime)
	dst.SetExp(exp)

	return dst, nil
}

func (p *Pool) withinContainerSession(
	ctx context.Context,
	c *sdkClientWrapper,
	containerID cid.ID,
	signer user.Signer,
	verb session.Verb,
	params containerSessionParams,
) error {
	_, errV1 := params.GetSession()
	_, errV2 := params.GetSessionV2()

	switch {
	case errV1 == nil && errV2 == nil:
		return client.ErrSessionTokenBothVersionsSet
	case errV2 == nil || errors.Is(errV2, client.ErrNoSessionExplicitly):
		return nil
	case errV1 == nil || errors.Is(errV1, client.ErrNoSessionExplicitly):
		return nil
	default:
	}

	cacheKey := cacheKeyForSession(c.addr, signer, verb, containerID)

	tokV2, ok := p.cache.Get(cacheKey)
	if !ok {
		var err error
		tokV2, err = initSession(ctx, c, p.stokenDuration, signer)
		if err != nil {
			return fmt.Errorf("init session v2: %w", err)
		}

		ctxV2, err := session.NewContext(containerID, []session.Verb{verb})
		if err != nil {
			return fmt.Errorf("create context v2: %w", err)
		}

		if err = tokV2.AddContext(ctxV2); err != nil {
			return fmt.Errorf("add context v2: %w", err)
		}

		if err = tokV2.Sign(signer); err != nil {
			return fmt.Errorf("sign token v2: %w", err)
		}

		p.cache.Put(cacheKey, tokV2)
	}

	params.WithinSessionV2(tokV2)

	return nil
}
