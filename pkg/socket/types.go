package socket

type CaptureInfo struct {
	ContainerName      string `json:"container_name,omitempty"`
	ContainerNamespace string `json:"container_namespace,omitempty"`
	ContainerID        string `json:"containerID,omitempty"`
	Interface          string `json:"interface,omitempty"`
	FileName           string `json:"file_name,omitempty"`
}

type Capture struct {
	*CaptureInfo `json:"CaptureInfo,omitempty"`
	Stats        Stats `json:"Stats,omitempty"`
}

type Stats struct {
	NbPacket uint `json:"NbPacket,omitempty"`
	NbBytes  uint `json:"NbBytes,omitempty"`
}
