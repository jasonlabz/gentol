package embedded

import _ "embed"

//go:embed template.tar.gz
var TemplateData []byte
