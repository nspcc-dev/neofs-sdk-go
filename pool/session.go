package pool

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neofs-sdk-go/client"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofsecdsa "github.com/nspcc-dev/neofs-sdk-go/crypto/ecdsa"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	sessionv2 "github.com/nspcc-dev/neofs-sdk-go/session/v2"
	"github.com/nspcc-dev/neofs-sdk-go/user"
)

func initSession(ctx context.Context, c *sdkClientWrapper, dur uint64, signer user.Signer) (session.Object, error) {
	tok := c.nodeSession.GetNodeSession(signer.Public())
	if tok != nil {
		return *tok, nil
	}

	var dst session.Object
	ni, err := c.NetworkInfo(ctx, client.PrmNetworkInfo{})
	if err != nil {
		return dst, err
	}

	epoch := ni.CurrentEpoch()

	var exp uint64
	if math.MaxUint64-epoch < dur {
		exp = math.MaxUint64
	} else {
		exp = epoch + dur
	}

	var prm client.PrmSessionCreate
	prm.SetExp(exp)

	res, err := c.SessionCreate(ctx, signer, prm)

	if err != nil {
		return dst, err
	}

	var id uuid.UUID
	if err = id.UnmarshalBinary(res.ID()); err != nil {
		return dst, fmt.Errorf("invalid session token ID: %w", err)
	}

	var key neofsecdsa.PublicKey
	if err = key.Decode(res.PublicKey()); err != nil {
		return dst, fmt.Errorf("invalid public session key: %w", err)
	}

	dst.SetID(id)
	dst.SetAuthKey(&key)
	dst.SetExp(exp)

	c.nodeSession.SetNodeSession(&dst, signer.Public())

	return dst, nil
}

func initSessionV2(ctx context.Context, c *sdkClientWrapper) (sessionv2.Token, error) {
	var dst sessionv2.Token
	dst.SetVersion(sessionv2.TokenCurrentVersion)
	dst.SetNonce(sessionv2.RandomNonce())

	ei, err := c.EndpointInfo(ctx, client.PrmEndpointInfo{})
	if err != nil {
		return dst, fmt.Errorf("get endpoint info: %w", err)
	}
	neoPubKey, err := keys.NewPublicKeyFromBytes(ei.NodeInfo().PublicKey(), elliptic.P256())
	if err != nil {
		return dst, fmt.Errorf("parse node public key: %w", err)
	}

	ecdsaPubKey := (*ecdsa.PublicKey)(neoPubKey)

	userID := user.NewFromECDSAPublicKey(*ecdsaPubKey)
	if err = dst.AddSubject(sessionv2.NewTargetUser(userID)); err != nil {
		return dst, fmt.Errorf("add subject: %w", err)
	}
	return dst, nil
}

func (p *Pool) withinContainerSession(
	ctx context.Context,
	c *sdkClientWrapper,
	containerID cid.ID,
	signer user.Signer,
	verb session.ObjectVerb,
	params containerSessionParams,
) error {
	_, errV1 := params.GetSession()
	originV2, errV2 := params.GetSessionV2()

	switch {
	case errV1 == nil && errV2 == nil:
		return errors.New("both session versions are set")
	case errV1 == nil || errors.Is(errV1, client.ErrNoSessionExplicitly):
		return nil
	case errors.Is(errV2, client.ErrNoSessionExplicitly):
		return nil
	case errV2 == nil:
		if p.disableDelegateSessionV2 {
			return nil
		}

		tokenHash := hex.EncodeToString(sha256.New().Sum(originV2.SignedData()))
		cacheKey := cacheKeyForSessionV2(c.addr, signer, containerID, tokenHash)

		tokV2, ok := p.cache.GetV2(cacheKey)
		if !ok {
			var err error
			tokV2, err = initSessionV2(ctx, c)
			if err != nil {
				return fmt.Errorf("init session v2: %w", err)
			}
			if err = tokV2.SetContexts(originV2.Contexts()); err != nil {
				return fmt.Errorf("set contexts token v2: %w", err)
			}
			tokV2.SetExp(originV2.Exp())
			tokV2.SetNbf(originV2.Nbf())
			tokV2.SetIat(originV2.Iat())

			tokV2.SetOrigin(originV2)

			if err = tokV2.Sign(p.signer); err != nil {
				return fmt.Errorf("sign token v2: %w", err)
			}

			p.cache.PutV2(cacheKey, tokV2)
		}

		params.WithinSessionV2(tokV2)
		return nil
	default:
	}

	// Use v2 tokens if configured
	if p.useV2Sessions {
		return p.withinContainerSessionV2(ctx, c, containerID, signer, params)
	}

	// Default to v1 tokens
	return p.withinContainerSessionV1(ctx, c, containerID, signer, verb, params)
}

func (p *Pool) withinContainerSessionV1(
	ctx context.Context,
	c *sdkClientWrapper,
	containerID cid.ID,
	signer user.Signer,
	verb session.ObjectVerb,
	params containerSessionParams,
) error {
	cacheKey := cacheKeyForSession(c.addr, signer, verb, containerID)

	tok, ok := p.cache.Get(cacheKey)
	if !ok {
		// init new session or take base session data from cache
		var err error
		tok, err = initSession(ctx, c, p.stokenDuration, signer)
		if err != nil {
			return fmt.Errorf("init session: %w", err)
		}

		tok.ForVerb(verb)
		tok.BindContainer(containerID)

		// sign the token
		if err = tok.Sign(signer); err != nil {
			return fmt.Errorf("sign token: %w", err)
		}

		// cache the opened session
		p.cache.Put(cacheKey, tok)
	}

	params.WithinSession(tok)

	return nil
}

func (p *Pool) withinContainerSessionV2(
	ctx context.Context,
	c *sdkClientWrapper,
	containerID cid.ID,
	signer user.Signer,
	params containerSessionParams,
) error {
	cacheKey := cacheKeyForSessionV2(c.addr, signer, containerID, "")

	tokV2, ok := p.cache.GetV2(cacheKey)
	if !ok {
		var err error
		tokV2, err = initSessionV2(ctx, c)
		if err != nil {
			return fmt.Errorf("init session v2: %w", err)
		}

		ctxV2, err := sessionv2.NewContext(containerID, []sessionv2.Verb{
			sessionv2.VerbObjectPut,
			sessionv2.VerbObjectGet,
			sessionv2.VerbObjectHead,
			sessionv2.VerbObjectSearch,
			sessionv2.VerbObjectDelete,
			sessionv2.VerbObjectRange,
			sessionv2.VerbObjectRangeHash,
		})
		if err != nil {
			return fmt.Errorf("create context v2: %w", err)
		}

		if err = tokV2.AddContext(ctxV2); err != nil {
			return fmt.Errorf("add context v2: %w", err)
		}

		currentTime := time.Now()
		tokV2.SetIat(currentTime)
		tokV2.SetNbf(currentTime)
		tokV2.SetExp(currentTime.Add(time.Duration(p.stokenDuration) * time.Second))

		if err = tokV2.Sign(signer); err != nil {
			return fmt.Errorf("sign token v2: %w", err)
		}

		p.cache.PutV2(cacheKey, tokV2)
	}

	params.WithinSessionV2(tokV2)

	return nil
}
