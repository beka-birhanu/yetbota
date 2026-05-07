package postphoto

import (
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmodels"
	"github.com/beka-birhanu/yetbota/content-service/internal/domain/postphoto"
)

func buildQueryMods(opts *postphoto.Options) []qm.QueryMod {
	mods := []qm.QueryMod{}
	if opts == nil {
		return mods
	}

	if opts.LoadPhoto {
		mods = append(mods, qm.Load(dbmodels.PostPhotoRels.Photo))
	}

	if len(opts.PostIDs) > 0 {
		mods = append(mods, dbmodels.PostPhotoWhere.PostID.IN(opts.PostIDs))
	}

	return mods
}
