package socket

type Capture struct {
	ContainerName      string `json:"container_name,omitempty"`
	ContainerNamespace string `json:"container_namespace,omitempty"`
	ContainerID        string `json:"containerID,omitempty"`
	Interface          string `json:"interface,omitempty"`
	FileName           string `json:"file_name,omitempty"`
}
