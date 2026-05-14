package metadata

type ProjectMeta struct {
	ModulePath         string
	ProjectName        string
	ServiceName        string
	ServicePackageName string
	ServiceStructName  string
	ServiceDir         string
}

func (p *ProjectMeta) GenRenderData() map[string]any {
	result := map[string]any{
		"ModulePath":         p.ModulePath,
		"ProjectName":        p.ProjectName,
		"ServiceName":        p.ServiceName,
		"ServiceStructName":  p.ServiceStructName,
		"ServicePackageName": p.ServicePackageName,
		"ServiceDir":         p.ServiceDir,
	}
	return result
}
