package types

// Camera represents a Frigate camera configuration
type Camera struct {
	Name         string        `json:"name"`
	Enabled      bool          `json:"enabled"`
	Entities     []string      `json:"entities"`
	FrameTime    int           `json:"frame_time"`
	Models       *CameraModels `json:"models,omitempty"`
	FFmpegInputs *FFmpegInput  `json:"ffmpeg_inputs,omitempty"`
}

// CameraModels contains model configuration for camera
type CameraModels struct {
	Object *ObjectDetection `json:"object,omitempty"`
}

// ObjectDetection contains object detection config
type ObjectDetection struct {
	Enabled   bool `json:"enabled"`
	MaxLabels int  `json:"max_labels,omitempty"`
}

// FFmpegInput contains ffmpeg input configuration
type FFmpegInput struct {
	Path    string `json:"path"`
	ROSpeed int    `json:"ro_speed,omitempty"`
}

// CamerasResponse represents the response from /api/cameras
type CamerasResponse map[string]Camera

// StreamConfig represents stream configuration
type StreamConfig struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
	Path    string `json:"path,omitempty"`
}

// GetStreamsResponse represents the response from /api/streams
type GetStreamsResponse map[string]StreamConfig

// Go2RTCMedia represents media information in go2rtc
type Go2RTCMedia struct {
	Media     string `json:"media"`
	Direction string `json:"direction"`
	Codec     string `json:"codec"`
}

// Go2RTCCodec represents codec information
type Go2RTCCodec struct {
	CodecName  string `json:"codec_name"`
	CodecType  string `json:"codec_type"`
	Level      int    `json:"level,omitempty"`
	Profile    string `json:"profile,omitempty"`
	SampleRate int    `json:"sample_rate,omitempty"`
}

// Go2RTCProducer represents a producer in go2rtc stream
type Go2RTCProducer struct {
	ID         int              `json:"id"`
	FormatName string           `json:"format_name"`
	Protocol   string           `json:"protocol"`
	RemoteAddr string           `json:"remote_addr,omitempty"`
	Source     string           `json:"source,omitempty"`
	URL        string           `json:"url,omitempty"`
	SDP        string           `json:"sdp,omitempty"`
	UserAgent  string           `json:"user_agent,omitempty"`
	Medias     []string         `json:"medias,omitempty"`
	Receivers  []Go2RTCReceiver `json:"receivers,omitempty"`
}

// Go2RTCReceiver represents receiver information
type Go2RTCReceiver struct {
	ID      int         `json:"id"`
	Codec   Go2RTCCodec `json:"codec"`
	Childs  []int       `json:"childs,omitempty"`
	Bytes   int64       `json:"bytes"`
	Packets int         `json:"packets"`
}

// Go2RTCConsumer represents a consumer in go2rtc stream
type Go2RTCConsumer struct {
	ID         int            `json:"id"`
	FormatName string         `json:"format_name"`
	Protocol   string         `json:"protocol"`
	RemoteAddr string         `json:"remote_addr,omitempty"`
	SDP        string         `json:"sdp,omitempty"`
	UserAgent  string         `json:"user_agent,omitempty"`
	Medias     []string       `json:"medias,omitempty"`
	Senders    []Go2RTCSender `json:"senders,omitempty"`
	BytesSend  int64          `json:"bytes_send,omitempty"`
}

// Go2RTCSender represents sender information
type Go2RTCSender struct {
	ID      int         `json:"id"`
	Codec   Go2RTCCodec `json:"codec"`
	Parent  int         `json:"parent,omitempty"`
	Bytes   int64       `json:"bytes"`
	Packets int         `json:"packets"`
}

// Go2RTCStream represents a single stream from go2rtc
type Go2RTCStream struct {
	Producers []Go2RTCProducer `json:"producers"`
	Consumers []Go2RTCConsumer `json:"consumers"`
}

// Go2RTCStreamsResponse represents the response from /api/go2rtc/streams
type Go2RTCStreamsResponse map[string]Go2RTCStream
