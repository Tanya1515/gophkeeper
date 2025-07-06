package client_storage

// Custom type, that describes operations in application
type Operation string

// OperationInfo - struct, that containts necessary information
// for executing operation again.
type OperationInfo struct {
	OperationID   string    // Operation identificator
	OperationName Operation // Operation name
	OperationTime string    // Time, at which the operation was performed
	Fields        []string  // Data, that are going to be changed after performing the request
}

// Constants, that are related to type Operations.
// The values describe CRUD operations with objects.
const (
	Get    Operation = "get"
	Update Operation = "update"
	Delete Operation = "delete"
	Create Operation = "create"
)
