package azuredevops

const (
	JsonPatchContentType = "application/json-patch+json"
)

// The JSON model for a JSON Patch operation
type JsonPatchOperation struct {
	// The path to copy from for the Move/Copy operation.
	From *string `json:"from,omitempty"`
	// The patch operation
	Op *Operation `json:"op,omitempty"`
	// The path for the operation. In the case of an array, a zero based index can be used to specify the position in the array (e.g. /biscuits/0/name). The "-" character can be used instead of an index to insert at the end of the array (e.g. /biscuits/-).
	Path *string `json:"path,omitempty"`
	// The value for the operation. This is either a primitive or a JToken.
	Value interface{} `json:"value,omitempty"`
}
