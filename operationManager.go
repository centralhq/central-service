package main

type OperationManager struct {
	service StoreResource
}

func NewOperationManager(storeResource StoreResource) *OperationManager {
	return &OperationManager{
		service: storeResource,
	}
}

func (op *OperationManager) executeSync(objectId string) (uint64, error) {
	newCounter, err := op.service.upsert(objectId)
	if err != nil {
		return 0, err
	}

	return newCounter, nil
}

