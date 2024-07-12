package model

import "gopkg.in/yaml.v3"

// RawMessage allows for delayed unmarshalling of YAML fields.
// Note: Using this instead of yaml.Node for ease of testing.
type RawMessage struct {
	unmarshal func(any) error
	data      any
}

// UnmarshalYAML is used to set the unmarshal function for a RawMessage
// and set its internal state (data).
func (msg *RawMessage) UnmarshalYAML(unmarshal func(any) error) error {
	msg.unmarshal = unmarshal

	var temp any
	err := unmarshal(&temp)
	if err != nil {
		return err
	}

	msg.data = temp
	return nil
}

// Unmarshal is used to unmarshal a value into a RawMessage.
func (msg *RawMessage) Unmarshal(v any) error {
	return msg.unmarshal(v)
}

// MarshalYAML is used to marshal the RawMessage back to YAML.
func (msg RawMessage) MarshalYAML() (interface{}, error) {
	return msg.data, nil
}

func NewRawMessage(instance any) (*RawMessage, error) {
	data, err := yaml.Marshal(instance)
	if err != nil {
		return nil, err
	}

	var rawMsg RawMessage
	err = yaml.Unmarshal(data, &rawMsg)
	if err != nil {
		return nil, err
	}

	rawMsg.data = instance
	return &rawMsg, nil
}
