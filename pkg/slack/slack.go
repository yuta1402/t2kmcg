package slack

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/yuta1402/t2kmpg/pkg/webparse"
)

type requestData struct {
	Text string `json:"text"`
}

func PostContestAnnouncement(createdContest *webparse.CreatedContest, apiURL string) (*http.Response, error) {
	startTimeStr := createdContest.Options.StartTime.Format("2006/01/02 15:04")
	endTimeStr := createdContest.Options.EndTime.Format("2006/01/02 15:04")

	text := "*「" + createdContest.Options.ContestTitle + "」開催のお知らせ*\n" +
		"日時: " + startTimeStr + " ~ " + endTimeStr + "\n" +
		"会場: " + createdContest.URL + "\n" +
		"\n参加できる方は:ok: 絵文字、参加できない方は:ng: 絵文字でお知らせください。"

	d := requestData{
		Text: text,
	}

	json, err := json.Marshal(d)
	if err != nil {
		return nil, err
	}

	res, err := http.Post(apiURL, "application/json", bytes.NewBuffer([]byte(json)))
	if err != nil {
		return nil, err
	}

	return res, nil
}
