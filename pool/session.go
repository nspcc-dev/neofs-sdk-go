package pool

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/google/uuid"
	"github.com/nspcc-dev/neofs-sdk-go/client"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofsecdsa "github.com/nspcc-dev/neofs-sdk-go/crypto/ecdsa"
	"github.com/nspcc-dev/neofs-sdk-go/session"
)

var (
	errContainerRequired = errors.New("container required")
)

func initSession(ctx context.Context, dst *session.Object, c *client.Client, dur uint64, signer neofscrypto.Signer) error {
	ni, err := c.NetworkInfo(ctx, client.PrmNetworkInfo{})
	if err != nil {
		return err
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
		return err
	}

	var id uuid.UUID
	if err = id.UnmarshalBinary(res.ID()); err != nil {
		return fmt.Errorf("invalid session token ID: %w", err)
	}

	var key neofsecdsa.PublicKey
	if err = key.Decode(res.PublicKey()); err != nil {
		return fmt.Errorf("invalid public session key: %w", err)
	}

	dst.SetID(id)
	dst.SetAuthKey(&key)
	dst.SetExp(exp)

	return nil
}

func (p *Pool) withinContainerSession(
	ctx context.Context,
	c *client.Client,
	containerID cid.ID,
	signer neofscrypto.Signer,
	verb session.ObjectVerb,
	params containerSessionParams,
) error {
	_, err := params.GetSession()
	if err == nil || errors.Is(err, client.ErrNoSessionExplicitly) {
		return nil
	}

	cacheKey := formCacheKey(fmt.Sprintf("%p", c), signer)

	tok, ok := p.cache.Get(cacheKey)
	if !ok {
		// init new session
		err = initSession(ctx, &tok, c, p.stokenDuration, signer)
		if err != nil {
			return fmt.Errorf("init session: %w", err)
		}

		// cache the opened session
		p.cache.Put(cacheKey, tok)
	}

	tok.ForVerb(verb)
	tok.BindContainer(containerID)

	// sign the token
	if err := tok.Sign(signer); err != nil {
		return fmt.Errorf("sign token: %w", err)
	}

	params.WithinSession(tok)

	return nil
}
