package pool

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/google/uuid"
	"github.com/nspcc-dev/neofs-sdk-go/client"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofsecdsa "github.com/nspcc-dev/neofs-sdk-go/crypto/ecdsa"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/nspcc-dev/neofs-sdk-go/user"
)

var (
	errContainerRequired = errors.New("container required")
)

func initSession(ctx context.Context, c *sdkClientWrapper, dur uint64, signer user.Signer) (session.Object, error) {
	tok := c.nodeSession.GetNodeSession()
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

	c.nodeSession.SetNodeSession(&dst)

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
	_, err := params.GetSession()
	if err == nil || errors.Is(err, client.ErrNoSessionExplicitly) {
		return nil
	}

	cacheKey := cacheKeyForSession(c.addr, signer, verb, containerID)

	tok, ok := p.cache.Get(cacheKey)
	if !ok {
		// init new session or take base session data from cache
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
