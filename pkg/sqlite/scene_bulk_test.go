//go:build integration
// +build integration

package sqlite_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSceneGetManyBulkLoaders(t *testing.T) {
	runWithRollbackTxn(t, "GetManyTagIDs matches singular GetTagIDs", func(t *testing.T, ctx context.Context) {
		ids := []int{
			sceneIDs[sceneIdxWithTag],
			sceneIDs[sceneIdxWithTwoTags],
			sceneIDs[sceneIdxWithThreeTags],
			sceneIDs[sceneIdxWithGallery], // no tags
		}

		bulk, err := db.Scene.GetManyTagIDs(ctx, ids)
		if !assert.NoError(t, err) {
			return
		}
		if !assert.Equal(t, len(ids), len(bulk)) {
			return
		}

		for i, id := range ids {
			want, err := db.Scene.GetTagIDs(ctx, id)
			if !assert.NoError(t, err) {
				return
			}
			assert.ElementsMatch(t, want, bulk[i], "scene %d (idx %d)", id, i)
		}
	})

	runWithRollbackTxn(t, "GetManyPerformerIDs matches singular", func(t *testing.T, ctx context.Context) {
		ids := []int{
			sceneIDs[sceneIdxWithPerformer],
			sceneIDs[sceneIdxWithTwoPerformers],
			sceneIDs[sceneIdxWithThreePerformers],
			sceneIDs[sceneIdxWithGallery], // no performers
		}

		bulk, err := db.Scene.GetManyPerformerIDs(ctx, ids)
		if !assert.NoError(t, err) {
			return
		}
		if !assert.Equal(t, len(ids), len(bulk)) {
			return
		}

		for i, id := range ids {
			want, err := db.Scene.GetPerformerIDs(ctx, id)
			if !assert.NoError(t, err) {
				return
			}
			assert.ElementsMatch(t, want, bulk[i], "scene %d (idx %d)", id, i)
		}
	})

	runWithRollbackTxn(t, "GetManyGalleryIDs matches singular", func(t *testing.T, ctx context.Context) {
		ids := []int{
			sceneIDs[sceneIdxWithGallery],
			sceneIDs[sceneIdxWithTag], // no galleries
		}

		bulk, err := db.Scene.GetManyGalleryIDs(ctx, ids)
		if !assert.NoError(t, err) {
			return
		}
		if !assert.Equal(t, len(ids), len(bulk)) {
			return
		}

		for i, id := range ids {
			want, err := db.Scene.GetGalleryIDs(ctx, id)
			if !assert.NoError(t, err) {
				return
			}
			assert.ElementsMatch(t, want, bulk[i], "scene %d (idx %d)", id, i)
		}
	})

	runWithRollbackTxn(t, "GetManyGroups matches singular", func(t *testing.T, ctx context.Context) {
		ids := []int{
			sceneIDs[sceneIdxWithGroup],
			sceneIDs[sceneIdxWithGroupWithParent],
			sceneIDs[sceneIdxWithTag], // no groups
		}

		bulk, err := db.Scene.GetManyGroups(ctx, ids)
		if !assert.NoError(t, err) {
			return
		}
		if !assert.Equal(t, len(ids), len(bulk)) {
			return
		}

		for i, id := range ids {
			want, err := db.Scene.GetGroups(ctx, id)
			if !assert.NoError(t, err) {
				return
			}
			assert.Equal(t, want, bulk[i], "scene %d (idx %d)", id, i)
		}
	})

	runWithRollbackTxn(t, "GetManyURLs matches singular", func(t *testing.T, ctx context.Context) {
		// All scenes set URL fields in the fixture loader.
		ids := []int{
			sceneIDs[sceneIdxWithTag],
			sceneIDs[sceneIdxWithGallery],
			sceneIDs[sceneIdxWithGroup],
		}

		bulk, err := db.Scene.GetManyURLs(ctx, ids)
		if !assert.NoError(t, err) {
			return
		}
		if !assert.Equal(t, len(ids), len(bulk)) {
			return
		}

		for i, id := range ids {
			want, err := db.Scene.GetURLs(ctx, id)
			if !assert.NoError(t, err) {
				return
			}
			// URL order must be preserved (position-sorted).
			assert.Equal(t, want, bulk[i], "scene %d (idx %d)", id, i)
		}
	})

	runWithRollbackTxn(t, "GetManyStashIDs matches singular", func(t *testing.T, ctx context.Context) {
		ids := []int{
			sceneIDs[sceneIdxWithTag],
			sceneIDs[sceneIdxWithGallery],
			sceneIDs[sceneIdxWithGroup],
		}

		bulk, err := db.Scene.GetManyStashIDs(ctx, ids)
		if !assert.NoError(t, err) {
			return
		}
		if !assert.Equal(t, len(ids), len(bulk)) {
			return
		}

		for i, id := range ids {
			want, err := db.Scene.GetStashIDs(ctx, id)
			if !assert.NoError(t, err) {
				return
			}
			assert.ElementsMatch(t, want, bulk[i], "scene %d (idx %d)", id, i)
		}
	})

	runWithRollbackTxn(t, "GetManyTagIDs empty input returns empty slice", func(t *testing.T, ctx context.Context) {
		bulk, err := db.Scene.GetManyTagIDs(ctx, nil)
		assert.NoError(t, err)
		assert.Empty(t, bulk)
	})

	runWithRollbackTxn(t, "SceneMarker FindManyBySceneIDs matches FindBySceneID", func(t *testing.T, ctx context.Context) {
		ids := []int{
			sceneIDs[sceneIdxWithMarkerAndTag],
			sceneIDs[sceneIdxWithMarkerTwoTags],
			sceneIDs[sceneIdxWithTag], // no markers
		}

		bulk, err := db.SceneMarker.FindManyBySceneIDs(ctx, ids)
		if !assert.NoError(t, err) {
			return
		}
		if !assert.Equal(t, len(ids), len(bulk)) {
			return
		}

		for i, id := range ids {
			want, err := db.SceneMarker.FindBySceneID(ctx, id)
			if !assert.NoError(t, err) {
				return
			}
			if !assert.Equal(t, len(want), len(bulk[i]), "scene %d (idx %d)", id, i) {
				continue
			}
			for j := range want {
				assert.Equal(t, want[j].ID, bulk[i][j].ID, "scene %d, marker %d", id, j)
				assert.Equal(t, want[j].SceneID, bulk[i][j].SceneID, "scene %d, marker %d", id, j)
			}
		}
	})
}

