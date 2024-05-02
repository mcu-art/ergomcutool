package assets

import (
	"embed"

	"github.com/mcu-art/ergomcutool/utils"
)

//go:embed all:assets
var EmbeddedAssets embed.FS

func CopyAssets(dest string, dirPerm, filePerm uint32) error {
	return utils.CopyEmbeddedDir(EmbeddedAssets, dest, dirPerm, filePerm)
}
