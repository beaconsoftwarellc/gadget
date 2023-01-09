package transaction

import (
	"github.com/beaconsoftwarellc/gadget/v2/database/qb"
	"github.com/samber/lo"
)

func appendIfMissing(slice []qb.TableField, i qb.TableField) []qb.TableField {
	if lo.Contains(slice, i) {
		return slice
	}
	return append(slice, i)
}
