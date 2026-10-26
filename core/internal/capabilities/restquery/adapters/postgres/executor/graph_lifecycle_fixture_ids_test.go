package executor

import "github.com/google/uuid"

type graphLifecycleIDs struct {
	rootOne, rootTwo       uuid.UUID
	alpha, middle, zulu    uuid.UUID
	missing, owner, owner2 uuid.UUID
	itemAlpha, itemMiddle  uuid.UUID
	itemZulu               uuid.UUID
}

func canonicalGraphIDs() graphLifecycleIDs {
	return graphLifecycleIDs{
		rootOne: materialUUID(101), rootTwo: materialUUID(102),
		alpha: materialUUID(201), middle: materialUUID(202),
		zulu: materialUUID(203), missing: materialUUID(204),
		owner: materialUUID(301), owner2: materialUUID(302),
		itemAlpha: materialUUID(401), itemMiddle: materialUUID(402),
		itemZulu: materialUUID(403),
	}
}

func canonicalGraphEdges(ids graphLifecycleIDs) map[string][][]any {
	return map[string][][]any{
		"children": {
			{ids.rootOne, ids.zulu, int16(1)},
			{ids.rootOne, ids.alpha, int16(2)},
			{ids.rootOne, ids.missing, int16(3)},
			{ids.rootTwo, ids.middle, int16(1)},
		},
		"owner": {{ids.rootOne, ids.owner, int16(1)}},
		"items": {
			{ids.alpha, ids.itemZulu, int16(1)},
			{ids.alpha, ids.itemAlpha, int16(2)},
			{ids.middle, ids.itemMiddle, int16(1)},
		},
	}
}

func canonicalGraphCandidates(ids graphLifecycleIDs) map[string][]uuid.UUID {
	return map[string][]uuid.UUID{
		"children": {ids.alpha, ids.middle, ids.zulu, ids.missing},
		"owner":    {ids.owner},
		"items":    {ids.itemAlpha, ids.itemMiddle, ids.itemZulu},
	}
}
