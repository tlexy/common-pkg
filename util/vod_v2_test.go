package utils

import (
	"log"
	"os"
	"testing"

	"github.com/tlexy/common-pkg/util/vod_volce"
)

func TestVodVolce_SubmitOcrTask2(t *testing.T) {
	vod := vod_volce.NewVodVolce("",
		"",
		"")
	runId, err := vod.SubmitOcrTask("", "")
	if err != nil {
		t.Fatal(err)
	}
	log.Printf("runId: %v\n", runId)
}

func TestVodVolce_QueryOcrTaskResult2(t *testing.T) {
	vod := vod_volce.NewVodVolce("",
		"",
		"")
	ocrRes, err := vod.QueryOcrTaskResult("")
	if err != nil {
		t.Fatal(err)
	}
	results := ""
	for _, ocr := range ocrRes.Result.Output.Task.Ocr.Texts {
		if ocr.DetailedInfo.Label == "Subtitle" {
			results += ocr.Text + "\n"
		}
	}
	// 创建一个文件，准备写入
	err = os.WriteFile("ocr_subtitle.json", []byte(results), 0644)
	if err != nil {
		t.Fatal(err)
	}

	//log.Printf("ocrRes: %+v\n", ocrRes)
}

func TestVodVolce_SubmitEraseTask2(t *testing.T) {
	vod := vod_volce.NewVodVolce("",
		"",
		"space-voice-extreme")
	runId, err := vod.SubmitEraseTask("", "")
	if err != nil {
		t.Fatal(err)
	}
	//"qb:f525d9768cfecaa2fee05e5405a797cf"
	log.Printf("runId: %v\n", runId)
}

func TestVodVolce_QueryEraseTask2(t *testing.T) {
	vod := vod_volce.NewVodVolce("",
		"",
		"")
	eraseRes, err := vod.QueryEraseTaskResult("")
	if err != nil {
		t.Fatal(err)
	}
	log.Printf("eraseRes: %+v\n", eraseRes)
}
