package relations

import (
	"context"
	"errors"
	"fmt"

	"github.com/nspcc-dev/neofs-sdk-go/bearer"
	"github.com/nspcc-dev/neofs-sdk-go/client"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/session"
)

// Tokens contains different tokens to perform requests in Relations implementations.
type Tokens struct {
	Session *session.Object
	Bearer  *bearer.Token
}

var (
	// ErrNoLeftSibling is an error that must be returned if object doesn't have left sibling in objects chain.
	ErrNoLeftSibling = errors.New("no left siblings")

	// ErrNoSplitInfo is an error that must be returned if requested object isn't virtual.
	ErrNoSplitInfo = errors.New("no split info")
)

// HeadExecutor describes methods to get object head.
type HeadExecutor interface {
	ObjectHead(ctx context.Context, containerID cid.ID, objectID oid.ID, prm client.PrmObjectHead) (*client.ResObjectHead, error)
}

// SearchExecutor describes methods to search objects.
type SearchExecutor interface {
	ObjectSearchInit(ctx context.Context, containerID cid.ID, prm client.PrmObjectSearch) (*client.ObjectListReader, error)
}

// Executor describes all methods required to find all siblings for object.
type Executor interface {
	HeadExecutor
	SearchExecutor
}

// Get returns all related phy objects for provided root object ID in split-chain order, without linking object id.
// If linking object is found its id will be returned in the second result variable.
//
// Result doesn't include root object ID itself.
func Get(ctx context.Context, executor Executor, containerID cid.ID, rootObjectID oid.ID, tokens Tokens) ([]oid.ID, *oid.ID, error) {
	splitInfo, err := getSplitInfo(ctx, executor, containerID, rootObjectID, tokens)
	if err != nil {
		if errors.Is(err, ErrNoSplitInfo) {
			return []oid.ID{}, nil, nil
		}

		return nil, nil, err
	}

	// collect split chain by the descending ease of operations (ease is evaluated heuristically).
	// If any approach fails, we don't try the next since we assume that it will fail too.
	if _, ok := splitInfo.Link(); !ok {
		// the list is expected to contain last part and (probably) split info
		list, err := findSiblingByParentID(ctx, executor, containerID, rootObjectID, tokens)
		if err != nil {
			return nil, nil, fmt.Errorf("children: %w", err)
		}

		for _, id := range list {
			split, err := getSplitInfo(ctx, executor, containerID, id, tokens)
			if err != nil {
				if errors.Is(err, ErrNoSplitInfo) {
					continue
				}
				return nil, nil, fmt.Errorf("split info: %w", err)
			}
			if link, ok := split.Link(); ok {
				splitInfo.SetLink(link)
				break
			}
			if last, ok := split.LastPart(); ok {
				splitInfo.SetLastPart(last)
			}
		}
	}

	if idLinking, ok := splitInfo.Link(); ok {
		children, err := listChildrenByLinker(ctx, executor, containerID, idLinking, tokens)
		if err != nil {
			return nil, nil, fmt.Errorf("linking object's header: %w", err)
		}

		return children, &idLinking, nil
	}

	idMember, ok := splitInfo.LastPart()
	if !ok {
		return nil, nil, errors.New("missing any data in received object split information")
	}

	chain := []oid.ID{idMember}
	chainSet := map[oid.ID]struct{}{idMember: {}}

	for {
		idMember, err = getLeftSibling(ctx, executor, containerID, idMember, tokens)
		if err != nil {
			if errors.Is(err, ErrNoLeftSibling) {
				break
			}
			return nil, nil, fmt.Errorf("split chain member's header: %w", err)
		}

		if _, ok = chainSet[idMember]; ok {
			return nil, nil, fmt.Errorf("duplicated member in the split chain %s", idMember)
		}

		chain = append([]oid.ID{idMember}, chain...)
		chainSet[idMember] = struct{}{}
	}

	return chain, nil, nil
}

func getSplitInfo(ctx context.Context, header HeadExecutor, cnrID cid.ID, objID oid.ID, tokens Tokens) (*object.SplitInfo, error) {
	var prmHead client.PrmObjectHead
	if tokens.Bearer != nil {
		prmHead.WithBearerToken(*tokens.Bearer)
	}
	if tokens.Session != nil {
		prmHead.WithinSession(*tokens.Session)
	}
	prmHead.MarkRaw()
	res, err := header.ObjectHead(ctx, cnrID, objID, prmHead)

	if err != nil {
		var errSplit *object.SplitInfoError
		if errors.As(err, &errSplit) {
			return errSplit.SplitInfo(), nil
		}

		return nil, fmt.Errorf("raw object header: %w", err)
	}

	var hdr object.Object
	if !res.ReadHeader(&hdr) {
		return nil, errors.New("header")
	}

	if hdr.SplitID() == nil {
		return nil, ErrNoSplitInfo
	}

	si := object.NewSplitInfo()
	si.SetSplitID(hdr.SplitID())

	if hdr.HasParent() {
		if len(hdr.Children()) > 0 {
			si.SetLink(objID)
		} else {
			si.SetLastPart(objID)
		}
	}

	return si, nil
}

func findSiblingByParentID(ctx context.Context, searcher SearchExecutor, cnrID cid.ID, objID oid.ID, tokens Tokens) ([]oid.ID, error) {
	var query object.SearchFilters
	var prm client.PrmObjectSearch

	query.AddParentIDFilter(object.MatchStringEqual, objID)
	prm.SetFilters(query)

	if tokens.Bearer != nil {
		prm.WithBearerToken(*tokens.Bearer)
	}
	if tokens.Session != nil {
		prm.WithinSession(*tokens.Session)
	}

	resSearch, err := searcher.ObjectSearchInit(ctx, cnrID, prm)
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}

	var res []oid.ID
	err = resSearch.Iterate(func(id oid.ID) bool {
		res = append(res, id)
		return false
	})

	if err != nil {
		return nil, fmt.Errorf("iterate: %w", err)
	}

	return res, nil
}

func listChildrenByLinker(ctx context.Context, header HeadExecutor, cnrID cid.ID, objID oid.ID, tokens Tokens) ([]oid.ID, error) {
	var prm client.PrmObjectHead
	if tokens.Bearer != nil {
		prm.WithBearerToken(*tokens.Bearer)
	}
	if tokens.Session != nil {
		prm.WithinSession(*tokens.Session)
	}

	res, err := header.ObjectHead(ctx, cnrID, objID, prm)
	if err != nil {
		return nil, fmt.Errorf("linking object's header: %w", err)
	}

	var hdr object.Object
	if !res.ReadHeader(&hdr) {
		return nil, errors.New("header")
	}

	return hdr.Children(), nil
}

func getLeftSibling(ctx context.Context, header HeadExecutor, cnrID cid.ID, objID oid.ID, tokens Tokens) (oid.ID, error) {
	var prm client.PrmObjectHead
	if tokens.Bearer != nil {
		prm.WithBearerToken(*tokens.Bearer)
	}
	if tokens.Session != nil {
		prm.WithinSession(*tokens.Session)
	}

	res, err := header.ObjectHead(ctx, cnrID, objID, prm)
	if err != nil {
		return oid.ID{}, fmt.Errorf("split chain member's header: %w", err)
	}

	var hdr object.Object
	if !res.ReadHeader(&hdr) {
		return oid.ID{}, errors.New("header")
	}

	idMember, ok := hdr.PreviousID()
	if !ok {
		return oid.ID{}, ErrNoLeftSibling
	}

	return idMember, nil
}
