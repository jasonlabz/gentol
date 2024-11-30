package metadata

type ServerMeta struct {
	BaseConfig
	ProjectName string
}

func (s *ServerMeta) GenRenderData() map[string]any {
	result := map[string]any{
		"ProjectName": s.ProjectName,
	}
	return result
}
