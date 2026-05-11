package api

import (
	"context"
	"fmt"
	"time"

	"github.com/stashapp/stash/internal/api/loaders"
	"github.com/stashapp/stash/internal/api/urlbuilders"
	"github.com/stashapp/stash/internal/manager"
	"github.com/stashapp/stash/pkg/models"
)

func convertVideoFile(f models.File) (*models.VideoFile, error) {
	vf, ok := f.(*models.VideoFile)
	if !ok {
		return nil, fmt.Errorf("file %T is not a video file", f)
	}
	return vf, nil
}

func (r *sceneResolver) getPrimaryFile(ctx context.Context, obj *models.Scene) (*models.VideoFile, error) {
	if obj.PrimaryFileID != nil {
		f, err := loaders.From(ctx).FileByID.Load(*obj.PrimaryFileID)
		if err != nil {
			return nil, err
		}

		ret, err := convertVideoFile(f)
		if err != nil {
			return nil, err
		}

		obj.Files.SetPrimary(ret)

		return ret, nil
	} else {
		_ = obj.LoadPrimaryFile(ctx, r.repository.File)
	}

	return nil, nil
}

func (r *sceneResolver) getFiles(ctx context.Context, obj *models.Scene) ([]*models.VideoFile, error) {
	fileIDs, err := loaders.From(ctx).SceneFiles.Load(obj.ID)
	if err != nil {
		return nil, err
	}

	files, errs := loaders.From(ctx).FileByID.LoadAll(fileIDs)
	err = firstError(errs)
	if err != nil {
		return nil, err
	}

	ret := make([]*models.VideoFile, len(files))
	for i, f := range files {
		ret[i], err = convertVideoFile(f)
		if err != nil {
			return nil, err
		}
	}

	obj.Files.Set(ret)

	return ret, nil
}

func (r *sceneResolver) Date(ctx context.Context, obj *models.Scene) (*string, error) {
	if obj.Date != nil {
		result := obj.Date.String()
		return &result, nil
	}
	return nil, nil
}

func (r *sceneResolver) Files(ctx context.Context, obj *models.Scene) ([]*VideoFile, error) {
	files, err := r.getFiles(ctx, obj)
	if err != nil {
		return nil, err
	}

	ret := make([]*VideoFile, len(files))

	for i, f := range files {
		ret[i] = &VideoFile{
			VideoFile: f,
		}
	}

	return ret, nil
}

func (r *sceneResolver) Rating(ctx context.Context, obj *models.Scene) (*int, error) {
	if obj.Rating != nil {
		rating := models.Rating100To5(*obj.Rating)
		return &rating, nil
	}
	return nil, nil
}

func (r *sceneResolver) Rating100(ctx context.Context, obj *models.Scene) (*int, error) {
	return obj.Rating, nil
}

func (r *sceneResolver) Paths(ctx context.Context, obj *models.Scene) (*ScenePathsType, error) {
	baseURL, _ := ctx.Value(BaseURLCtxKey).(string)
	config := manager.GetInstance().Config
	builder := urlbuilders.NewSceneURLBuilder(baseURL, obj)
	screenshotPath := builder.GetScreenshotURL()
	previewPath := builder.GetStreamPreviewURL()
	streamPath := builder.GetStreamURL(config.GetAPIKey()).String()
	webpPath := builder.GetStreamPreviewImageURL()
	objHash := obj.GetHash(config.GetVideoFileNamingAlgorithm())
	vttPath := builder.GetSpriteVTTURL(objHash)
	spritePath := builder.GetSpriteURL(objHash)
	funscriptPath := builder.GetFunscriptURL(config.GetAPIKey()).String()
	captionBasePath := builder.GetCaptionURL()
	interactiveHeatmap := builder.GetInteractiveHeatmapURL()

	return &ScenePathsType{
		Screenshot:         &screenshotPath,
		Preview:            &previewPath,
		Stream:             &streamPath,
		Webp:               &webpPath,
		Vtt:                &vttPath,
		Sprite:             &spritePath,
		Funscript:          &funscriptPath,
		InteractiveHeatmap: &interactiveHeatmap,
		Caption:            &captionBasePath,
	}, nil
}

func (r *sceneResolver) SceneMarkers(ctx context.Context, obj *models.Scene) ([]*models.SceneMarker, error) {
	return loaders.From(ctx).SceneMarkers.Load(obj.ID)
}

func (r *sceneResolver) Captions(ctx context.Context, obj *models.Scene) (ret []*models.VideoCaption, err error) {
	primaryFile, err := r.getPrimaryFile(ctx, obj)
	if err != nil {
		return nil, err
	}
	if primaryFile == nil {
		return nil, nil
	}

	if err := r.withReadTxn(ctx, func(ctx context.Context) error {
		ret, err = r.repository.File.GetCaptions(ctx, primaryFile.Base().ID)
		return err
	}); err != nil {
		return nil, err
	}

	return ret, err
}

func (r *sceneResolver) Galleries(ctx context.Context, obj *models.Scene) ([]*models.Gallery, error) {
	var galleryIDs []int
	if obj.GalleryIDs.Loaded() {
		galleryIDs = obj.GalleryIDs.List()
	} else {
		ids, err := loaders.From(ctx).SceneGalleryIDs.Load(obj.ID)
		if err != nil {
			return nil, err
		}
		obj.GalleryIDs = models.NewRelatedIDs(ids)
		galleryIDs = ids
	}

	ret, errs := loaders.From(ctx).GalleryByID.LoadAll(galleryIDs)
	return ret, firstError(errs)
}

func (r *sceneResolver) Studio(ctx context.Context, obj *models.Scene) (ret *models.Studio, err error) {
	if obj.StudioID == nil {
		return nil, nil
	}

	return loaders.From(ctx).StudioByID.Load(*obj.StudioID)
}

func (r *sceneResolver) loadSceneGroups(ctx context.Context, obj *models.Scene) ([]models.GroupsScenes, error) {
	if obj.Groups.Loaded() {
		return obj.Groups.List(), nil
	}
	groups, err := loaders.From(ctx).SceneGroups.Load(obj.ID)
	if err != nil {
		return nil, err
	}
	obj.Groups = models.NewRelatedGroups(groups)
	return groups, nil
}

func (r *sceneResolver) Movies(ctx context.Context, obj *models.Scene) ([]*SceneMovie, error) {
	groups, err := r.loadSceneGroups(ctx, obj)
	if err != nil {
		return nil, err
	}

	loader := loaders.From(ctx).GroupByID
	ret := make([]*SceneMovie, 0, len(groups))
	for _, sm := range groups {
		movie, err := loader.Load(sm.GroupID)
		if err != nil {
			return nil, err
		}

		sceneIdx := sm.SceneIndex
		ret = append(ret, &SceneMovie{
			Movie:      movie,
			SceneIndex: sceneIdx,
		})
	}

	return ret, nil
}

func (r *sceneResolver) Groups(ctx context.Context, obj *models.Scene) ([]*SceneGroup, error) {
	groups, err := r.loadSceneGroups(ctx, obj)
	if err != nil {
		return nil, err
	}

	loader := loaders.From(ctx).GroupByID
	ret := make([]*SceneGroup, 0, len(groups))
	for _, sm := range groups {
		group, err := loader.Load(sm.GroupID)
		if err != nil {
			return nil, err
		}

		sceneIdx := sm.SceneIndex
		ret = append(ret, &SceneGroup{
			Group:      group,
			SceneIndex: sceneIdx,
		})
	}

	return ret, nil
}

func (r *sceneResolver) Tags(ctx context.Context, obj *models.Scene) ([]*models.Tag, error) {
	var tagIDs []int
	if obj.TagIDs.Loaded() {
		tagIDs = obj.TagIDs.List()
	} else {
		ids, err := loaders.From(ctx).SceneTagIDs.Load(obj.ID)
		if err != nil {
			return nil, err
		}
		obj.TagIDs = models.NewRelatedIDs(ids)
		tagIDs = ids
	}

	ret, errs := loaders.From(ctx).TagByID.LoadAll(tagIDs)
	return ret, firstError(errs)
}

func (r *sceneResolver) Performers(ctx context.Context, obj *models.Scene) ([]*models.Performer, error) {
	var performerIDs []int
	if obj.PerformerIDs.Loaded() {
		performerIDs = obj.PerformerIDs.List()
	} else {
		ids, err := loaders.From(ctx).ScenePerformerIDs.Load(obj.ID)
		if err != nil {
			return nil, err
		}
		obj.PerformerIDs = models.NewRelatedIDs(ids)
		performerIDs = ids
	}

	ret, errs := loaders.From(ctx).PerformerByID.LoadAll(performerIDs)
	return ret, firstError(errs)
}

func (r *sceneResolver) StashIds(ctx context.Context, obj *models.Scene) ([]*models.StashID, error) {
	if !obj.StashIDs.Loaded() {
		ids, err := loaders.From(ctx).SceneStashIDs.Load(obj.ID)
		if err != nil {
			return nil, err
		}
		obj.StashIDs = models.NewRelatedStashIDs(ids)
	}

	return stashIDsSliceToPtrSlice(obj.StashIDs.List()), nil
}

func (r *sceneResolver) SceneStreams(ctx context.Context, obj *models.Scene) ([]*manager.SceneStreamEndpoint, error) {
	// load the primary file into the scene
	_, err := r.getPrimaryFile(ctx, obj)
	if err != nil {
		return nil, err
	}

	config := manager.GetInstance().Config

	baseURL, _ := ctx.Value(BaseURLCtxKey).(string)
	builder := urlbuilders.NewSceneURLBuilder(baseURL, obj)
	apiKey := config.GetAPIKey()

	return manager.GetSceneStreamPaths(obj, builder.GetStreamURL(apiKey), config.GetMaxStreamingTranscodeSize())
}

func (r *sceneResolver) Interactive(ctx context.Context, obj *models.Scene) (bool, error) {
	primaryFile, err := r.getPrimaryFile(ctx, obj)
	if err != nil {
		return false, err
	}
	if primaryFile == nil {
		return false, nil
	}

	return primaryFile.Interactive, nil
}

func (r *sceneResolver) InteractiveSpeed(ctx context.Context, obj *models.Scene) (*int, error) {
	primaryFile, err := r.getPrimaryFile(ctx, obj)
	if err != nil {
		return nil, err
	}
	if primaryFile == nil {
		return nil, nil
	}

	return primaryFile.InteractiveSpeed, nil
}

func (r *sceneResolver) loadSceneURLs(ctx context.Context, obj *models.Scene) ([]string, error) {
	if obj.URLs.Loaded() {
		return obj.URLs.List(), nil
	}
	urls, err := loaders.From(ctx).SceneURLs.Load(obj.ID)
	if err != nil {
		return nil, err
	}
	obj.URLs = models.NewRelatedStrings(urls)
	return urls, nil
}

func (r *sceneResolver) URL(ctx context.Context, obj *models.Scene) (*string, error) {
	urls, err := r.loadSceneURLs(ctx, obj)
	if err != nil {
		return nil, err
	}
	if len(urls) == 0 {
		return nil, nil
	}
	return &urls[0], nil
}

func (r *sceneResolver) Urls(ctx context.Context, obj *models.Scene) ([]string, error) {
	return r.loadSceneURLs(ctx, obj)
}

func (r *sceneResolver) OCounter(ctx context.Context, obj *models.Scene) (*int, error) {
	ret, err := loaders.From(ctx).SceneOCount.Load(obj.ID)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func (r *sceneResolver) LastPlayedAt(ctx context.Context, obj *models.Scene) (*time.Time, error) {
	ret, err := loaders.From(ctx).SceneLastPlayed.Load(obj.ID)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (r *sceneResolver) PlayCount(ctx context.Context, obj *models.Scene) (*int, error) {
	ret, err := loaders.From(ctx).ScenePlayCount.Load(obj.ID)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func (r *sceneResolver) PlayHistory(ctx context.Context, obj *models.Scene) ([]*time.Time, error) {
	ret, err := loaders.From(ctx).ScenePlayHistory.Load(obj.ID)
	if err != nil {
		return nil, err
	}

	// convert to pointer slice
	ptrRet := make([]*time.Time, len(ret))
	for i, t := range ret {
		tt := t
		ptrRet[i] = &tt
	}

	return ptrRet, nil
}

func (r *sceneResolver) OHistory(ctx context.Context, obj *models.Scene) ([]*time.Time, error) {
	ret, err := loaders.From(ctx).SceneOHistory.Load(obj.ID)
	if err != nil {
		return nil, err
	}

	// convert to pointer slice
	ptrRet := make([]*time.Time, len(ret))
	for i, t := range ret {
		tt := t
		ptrRet[i] = &tt
	}

	return ptrRet, nil
}

func (r *sceneResolver) CustomFields(ctx context.Context, obj *models.Scene) (map[string]interface{}, error) {
	m, err := loaders.From(ctx).SceneCustomFields.Load(obj.ID)
	if err != nil {
		return nil, err
	}

	if m == nil {
		return make(map[string]interface{}), nil
	}

	return m, nil
}
