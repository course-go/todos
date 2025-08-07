package health

type Option func(registry *Registry) error

func WithService(name, version string) Option {
	return func(registry *Registry) error {
		registry.service = name
		registry.version = version

		return nil
	}
}

func WithComponent(component *Component) Option {
	return func(registry *Registry) error {
		registry.components = append(registry.components, component)
		return nil
	}
}
