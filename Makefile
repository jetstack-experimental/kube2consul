
codegen:
	mockgen -package=mocks -source=pkg/interfaces/interfaces.go > pkg/mocks/mocks.go
