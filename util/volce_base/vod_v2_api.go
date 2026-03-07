package volce_base

// https://www.volcengine.com/docs/4/1582324?lang=zh
// https://www.volcengine.com/docs/4/1582324?lang=zh#input

type InputForStartExecutionInput struct {
	Type string `json:"Type,omitempty"`
	Vid  string `json:"Vid,omitempty"`
}

type HighLightModel string
type HightLightMode string

const (
	Miniseries HighLightModel = "Miniseries"
	Game       HighLightModel = "Game"

	StorylineCuts HightLightMode = "StorylineCuts"
	MiniGame      HightLightMode = "MiniGame"
)

// https://www.volcengine.com/docs/4/1582324?lang=zh#operationtaskhighlight
type HighLight struct {
	Model HighLightModel `json:"Model,omitempty"`
	Mode  string         `json:"Mode,omitempty"`
}

type OperationTaskAudioExtract struct {
	Voice bool `json:"Voice,omitempty"` // 人声分离，取值为"true"或"false"
}

// 识别提示语言，取值如下：
// cmn-Hans-CN: 简体中文// cmn-Hant-CN: 繁体中文
// eng-US: 英语// jpn-JP: 日语// kor-KR: 韩语// rus-RU: 俄语// fra-FR: 法语
// por-PT: 葡萄牙语// spa-ES: 西班牙语// vie-VN: 越南语
// mya-MM: 缅甸语// nld-NL: 荷兰语// deu-DE: 德语
// ind-ID: 印尼语// ita-IT: 意大利语// pol-PL: 波兰语
// tha-TH: 泰语// tur-TR: 土耳其语// ara-SA: 阿拉伯语// msa-MY: 马来语
// ron-RO: 罗马尼亚语// fil-PH: 菲律宾语// hin-IN: 印地语
type OperationTaskAsr struct {
	Type            string `json:"Type,omitempty"`            // 语音识别，speech: 对话;singing: 歌唱
	Language        string `json:"Language,omitempty"`        // 语音识别语言，取值为"zh"（中文）或"en"（英文）
	WithSpeakerInfo bool   `json:"WithSpeakerInfo,omitempty"` // 是否开启使说话人识别功能。开启后，会通过返回参数 speaker 返回说话人信息。
	WithConfidence  bool   `json:"WithConfidence,omitempty"`  // 是否返回置信度。如设为 true，会通过返回参数 Confidence 返回置信度。
	Mode            string `json:"Mode,omitempty"`            // 工作模式。缺省为标准模式。。使用其它模式需提交工单联系火山引擎技术支持团队申请。
}

type OperationTaskOcr struct {
	WithImageSet bool   `json:"WithImageSet,omitempty"` // 当输入为 Vid 时，是否对视频中的图集进行 OCR 识别。若视频中无图集，则返回空的识别结果。
	Mode         string `json:"Mode,omitempty"`         // Subtitle：默认模式。Detailed：会在任务结果中输出 OCR 识别的文本类型和位置信息。
}

type EraseAuto struct {
	Type           string `json:"Type,omitempty"` //Subtitle: 擦除 OCR 检测为字幕的文本;Text: (Beta) 擦除除场景文字（如宫殿门牌匾等）以外的字幕及其他文本（如人物介绍等）。
	SubtitleFilter struct {
	} `json:"SubtitleFilter"`
	Locations []struct {
		RatioLocation struct {
			TopLeftX     float64 `json:"TopLeftX,omitempty"`     // 框选区域左上角相对于视频左上角在X轴上的偏移比例，取值范围为[0,1]，其中 0 表示无偏移（与视频左边缘对齐），1 表示完全偏移（与视频右边缘对齐）。
			TopLeftY     float64 `json:"TopLeftY,omitempty"`     // 框选区域左上角相对于视频左上角在 Y 轴上的偏移比例，取值范围为 [0,1]，其中 0 表示无偏移（与视频上边缘对齐），1 表示完全偏移（与视频下边缘对齐）。
			BottomRightX float64 `json:"BottomRightX,omitempty"` // 框选区域右下角相对于视频左上角在 X 轴上的偏移比例，取值范围为 [0,1]，其中 0 表示无偏移（与视频左边缘对齐），1 表示完全偏移（与视频右边缘对齐）。
			BottomRightY float64 `json:"BottomRightY,omitempty"` // 框选区域右下角相对于视频左上角在 Y 轴上的偏移比例，取值范围为 [0,1]，其中 0 表示无偏移（与视频上边缘对齐），1 表示完全偏移（与视频下边缘对齐）。
		} `json:"RatioLocation,omitempty"`
	} `json:"Locations,omitempty"`
}

type OperationTaskErase struct {
	Mode          string     `json:"Mode,omitempty"` // Auto：自动擦除模式。在此模式下，系统将启用 OCR 识别，并依据检测结果进行擦除操作。Manual：(Beta) 手动擦除模式。在此模式下，系统不会启用 OCR 识别，仅擦除白色字幕内容。
	Auto          *EraseAuto `json:"Auto,omitempty"`
	NewVid        bool       `json:"NewVid,omitempty"`        // 是否创建新 Vid。取值为 true 或 false。
	WithEraseInfo bool       `json:"WithEraseInfo,omitempty"` // 是否返回擦除信息。取值为 true 或 false。
}

// https://www.volcengine.com/docs/4/1582324?lang=zh#operationtask
// https://www.volcengine.com/docs/4/1582324?lang=zh#operationtask
// 任务类型：

// Highlight: 高光智剪任务。
// AdAudit: 巨量广告预审任务。
// AudioExtract: 人声背景音分离任务。
// Vision: 长视频理解任务。
// Asr: ASR 提取字幕任务。
// Storyline: 故事线分析任务。
// Segment: 场景切分任务。
// Ocr: OCR 提取字幕任务。
// Erase: 精细化字幕擦除任务。
type Task struct {
	Type         string                     `json:"Type,omitempty"`
	Highlight    *HighLight                 `json:"Highlight,omitempty"`
	AudioExtract *OperationTaskAudioExtract `json:"AudioExtract,omitempty"` // 人声分离
	Asr          *OperationTaskAsr          `json:"Asr,omitempty"`          // 语音识别
	Ocr          *OperationTaskOcr          `json:"Ocr,omitempty"`          // 文字识别
	Erase        *OperationTaskErase        `json:"Erase,omitempty"`        // 视频擦除
}

type ConvertOperationForStartExecutionInput struct {
	Type string `json:"Type,omitempty"`
	Task *Task  `json:"Task,omitempty"`
}

type StartExecutionInput struct {
	Input     *InputForStartExecutionInput            `json:"Input,omitempty"`
	Operation *ConvertOperationForStartExecutionInput `json:"Operation,omitempty"`
	SpaceName string                                  `json:"SpaceName,omitempty"`
}

type ResponseMetadata struct {
	RequestId string `json:"RequestId,omitempty"`
	Action    string `json:"Action,omitempty"`
	Version   string `json:"Version,omitempty"`
	Service   string `json:"Service,omitempty"`
	Region    string `json:"Region,omitempty"`
}

type StartExecutionOutput struct {
	ResponseMetadata ResponseMetadata `json:"ResponseMetadata,omitempty"`
	Result           struct {
		RunId string `json:"RunId,omitempty"`
	} `json:"Result,omitempty"`
}

type ExecutionOcrResult struct {
	ResponseMetadata ResponseMetadata `json:"ResponseMetadata,omitempty"`
	Result           struct {
		Status string                       `json:"Status,omitempty"`
		RunId  string                       `json:"RunId,omitempty"`
		Input  *InputForStartExecutionInput `json:"Input,omitempty"`
		Meta   struct {
			CreateTime string `json:"CreateTime,omitempty"`
			EndTime    string `json:"EndTime,omitempty"`
			SpaceName  string `json:"SpaceName,omitempty"`
			StartTime  string `json:"StartTime,omitempty"`
			Trigger    string `json:"Trigger,omitempty"`
		} `json:"Meta,omitempty"`
		Operation *ConvertOperationForStartExecutionInput `json:"Operation,omitempty"`
		Output    struct {
			Type string `json:"Type,omitempty"`
			Task struct {
				Type string `json:"Type,omitempty"`
				Ocr  struct {
					Duration float64 `json:"Duration,omitempty"`
					Texts    []struct {
						Text         string  `json:"Text,omitempty"`
						Start        float64 `json:"Start,omitempty"`
						End          float64 `json:"End,omitempty"`
						DetailedInfo struct {
							Label          string `json:"Label,omitempty"`
							PixelRectangle struct {
								TopLeftX     float64 `json:"TopLeftX,omitempty"`
								TopLeftY     float64 `json:"TopLeftY,omitempty"`
								BottomRightX float64 `json:"BottomRightX,omitempty"`
								BottomRightY float64 `json:"BottomRightY,omitempty"`
							} `json:"PixelRectangle,omitempty"`
						} `json:"DetailedInfo,omitempty"`
					} `json:"Texts,omitempty"`
				} `json:"Ocr,omitempty"`
			} `json:"Task,omitempty"`
		} `json:"Output,omitempty"`
	} `json:"Result,omitempty"`
}

type ExecutionEraseResult struct {
	ResponseMetadata ResponseMetadata `json:"ResponseMetadata,omitempty"`
	Result           struct {
		Status string                       `json:"Status,omitempty"`
		RunId  string                       `json:"RunId,omitempty"`
		Input  *InputForStartExecutionInput `json:"Input,omitempty"`
		Meta   struct {
			CreateTime string `json:"CreateTime,omitempty"`
			EndTime    string `json:"EndTime,omitempty"`
			SpaceName  string `json:"SpaceName,omitempty"`
			StartTime  string `json:"StartTime,omitempty"`
			Trigger    string `json:"Trigger,omitempty"`
		} `json:"Meta,omitempty"`
		Operation *ConvertOperationForStartExecutionInput `json:"Operation,omitempty"`
		Output    struct {
			Type string `json:"Type,omitempty"`
			Task struct {
				Type  string `json:"Type,omitempty"`
				Erase struct {
					Duration float64 `json:"Duration,omitempty"`
					File     struct {
						Size     int64  `json:"Size,omitempty"`
						FileName string `json:"FileName,omitempty"`
						Vid      string `json:"Vid,omitempty"`
					} `json:"File,omitempty"`
					Info struct {
						Width  int32 `json:"Width,omitempty"`
						Height int32 `json:"Height,omitempty"`
						Areas  []struct {
							Start float64 `json:"Start,omitempty"`
							End   float64 `json:"End,omitempty"`
						} `json:"Areas,omitempty"`
					} `json:"Info,omitempty"`
				} `json:"Erase,omitempty"`
			} `json:"Task,omitempty"`
		} `json:"Output,omitempty"`
	} `json:"Result,omitempty"`
}
