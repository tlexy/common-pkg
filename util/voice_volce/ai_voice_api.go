package voice_volce

type AudioFormat string

const (
	AudioFormatMp3 AudioFormat = "mp3"
	AudioFormatWav AudioFormat = "wav"
)

type AsrRequest struct {
	AudioUrl    string
	Language    string
	AudioFormat AudioFormat
	RequestId   string
	CallbackUrl string
}

type AsrRequestInternalAudio struct {
	Url      string `json:"url"`
	Format   string `json:"format"`
	Language string `json:"language"`
}

type AsrRequestInternalRequest struct {
	EnableSpeakerInfo      bool   `json:"enable_speaker_info"`
	ShowUtterances         bool   `json:"show_utterances"`
	EnableLid              bool   `json:"enable_lid"`
	EnableEmotionDetection bool   `json:"enable_emotion_detection"`
	EnableGenderDetection  bool   `json:"enable_gender_detection"`
	VadSegment             bool   `json:"vad_segment"`
	Callback               string `json:"callback"`
}

type AsrRequestInternal struct {
	Audio   *AsrRequestInternalAudio   `json:"audio"`
	Request *AsrRequestInternalRequest `json:"request"`
}

type AsrResultRes struct {
	TaskId string
}
