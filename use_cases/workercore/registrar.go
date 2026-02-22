package workercore

type IRegistrar interface {
	Add(worker Worker)
}
