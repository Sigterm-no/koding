// services package is used for providing a Service interface. All the third
// party integrations need to implement this
package services

type Service interface {
	// PrepareMessage gets the input and prepares
	// the message body
	PrepareMessage(*ServiceInput) (string, error)

	// PrepareEndpoint prepares the endpoint url with given token
	PrepareEndpoint(string) (string, error)

	// Output extracts data with given input
	Output(*ServiceInput) *ServiceOutput
}

// ServiceInput is used for input objects
type ServiceInput map[string]interface{}

func (si ServiceInput) Key(key string) interface{} {
	val, ok := si[key]
	if !ok {
		return nil
	}

	return val
}

func (si ServiceInput) SetKey(key string, value interface{}) {
	si[key] = value
}

// ServiceOutput is used for extracting data from
// integration services
type ServiceOutput struct {
	Username  string `json:"username"`
	GroupName string `json:"groupName"`
}