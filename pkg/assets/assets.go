package assets

import (
	"time"

	"github.com/jessevdk/go-assets"
)

var _Assets97b9446a0df6936070d46bfa8dbe9801fcb7f8eb = "<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n    <meta charset=\"UTF-8\">\n    <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n    <title>Empires of Avalon - Admin area</title>\n    <style>\n        .main {\n            width: 960px;\n        }\n\n        .section {\n            padding: 0 30px;\n            margin: 30px auto;\n        }\n\n        .section article {\n            margin: 10px auto;\n            width: 100%;\n        }\n\n        .section article .label {\n            font-weight: bold;\n        }\n\n        .section article .context .attribute {\n            margin-left: 10px;\n        }\n    </style>\n</head>\n<body>\n    <h1>Empires of Avalon - Admin Dashboard</h1>\n\n    <em>({{ .Timestamp }})</em>\n\n    <div class=\"main\">\n        <h2>Registry</h2>\n        <h3>Inventories</h3>\n        {{ with .Data.registry.inventories }}\n        <div class=\"section\">\n            {{ range . }}\n                <article>\n                    <div class=\"attribute\"><span class=\"label\">Grain ID:</span> <span><a href=\"/admin/inventory/{{ .grain_id }}\" target=\"_blank\">{{ .grain_id }}</a></span></div>\n                    <div class=\"attribute\"><span class=\"label\">Identity:</span> <span>{{ .identity }}</span></div>\n                    <div class=\"attribute\"><span class=\"label\">Last seen:</span> <span>{{ .last_seen }}</span></div>\n                    <div class=\"attribute\"><span class=\"label\">Tolerations:</span> <span>{{ .tolerations }}</span></div>\n                    {{ with .context }}\n                    <div class=\"context\">\n                        {{ range $k, $v := . }}\n                        <div class=\"attribute\"><span class=\"label\">{{ $k }}:</span> <span>{{ $v }}</span></div>\n                        {{ end }}\n                    </div>\n                    {{ end }}\n                </article>\n            {{ end }}\n        </div>\n        {{ else }}\n        <b>No inventories found</b>\n        {{ end }}\n\n        <h3>Timers</h3>\n        {{ with .Data.registry.timers }}\n        <div class=\"section\">\n            {{ range . }}\n                <article>\n                    <div class=\"attribute\"><span class=\"label\">Grain ID:</span> <span><a href=\"/admin/timer/{{ .grain_id }}\" target=\"_blank\">{{ .grain_id }}</a></span></div>\n                    <div class=\"attribute\"><span class=\"label\">Identity:</span> <span>{{ .identity }}</span></div>\n                    <div class=\"attribute\"><span class=\"label\">Last seen:</span> <span>{{ .last_seen }}</span></div>\n                    <div class=\"attribute\"><span class=\"label\">Tolerations:</span> <span>{{ .tolerations }}</span></div>\n                    {{ with .context }}\n                    <div class=\"context\">\n                        <span class=\"label\">Context:</span><br />\n                        {{ range $k, $v := . }}\n                        <div class=\"attribute\"><span class=\"label\">{{ $k }}:</span> <span>{{ $v }}</span></div>\n                        {{ end }}\n                    </div>\n                    {{ end }}\n                </article>\n            {{ end }}\n        </div>\n        {{ else }}\n        <b>No timers found</b>\n        {{ end }}\n    </div>\n    \n</body>\n</html>"

// Assets returns go-assets FileSystem
var Assets = assets.NewFileSystem(map[string][]string{"/": []string{"assets"}, "/assets": []string{}, "/assets/templates": []string{}, "/assets/templates/admin": []string{"index.gohtml"}}, map[string]*assets.File{
	"/assets": &assets.File{
		Path:     "/assets",
		FileMode: 0x800001ed,
		Mtime:    time.Unix(1691480113, 1691480113725907006),
		Data:     nil,
	}, "/assets/templates": &assets.File{
		Path:     "/assets/templates",
		FileMode: 0x800001ed,
		Mtime:    time.Unix(1691480119, 1691480119598570096),
		Data:     nil,
	}, "/assets/templates/admin": &assets.File{
		Path:     "/assets/templates/admin",
		FileMode: 0x800001ed,
		Mtime:    time.Unix(1691480127, 1691480127850119080),
		Data:     nil,
	}, "/assets/templates/admin/index.gohtml": &assets.File{
		Path:     "/assets/templates/admin/index.gohtml",
		FileMode: 0x1a4,
		Mtime:    time.Unix(1691513327, 1691513327614456567),
		Data:     []byte(_Assets97b9446a0df6936070d46bfa8dbe9801fcb7f8eb),
	}, "/": &assets.File{
		Path:     "/",
		FileMode: 0x800001ed,
		Mtime:    time.Unix(1691482480, 1691482480572147412),
		Data:     nil,
	}}, "")
