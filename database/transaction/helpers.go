package transaction

import "github.com/beaconsoftwarellc/gadget/v2/database/qb"

func appendIfMissing(slice []qb.TableField, i qb.TableField) []qb.TableField {
	if contains(slice, i) {
		return slice
	}
	return append(slice, i)
}

// I think this can be replaced with lo
func contains(slice []qb.TableField, i qb.TableField) bool {
	for _, ele := range slice {
		if ele == i {
			return true
		}
	}
	return false
}
