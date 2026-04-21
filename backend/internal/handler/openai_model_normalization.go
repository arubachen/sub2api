package handler

import (
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

func normalizeOpenAICompatModelInBody(body []byte) ([]byte, string, bool, error) {
	if len(body) == 0 {
		return body, "", false, nil
	}

	model := strings.TrimSpace(gjson.GetBytes(body, "model").String())
	if model == "" {
		return body, "", false, nil
	}

	normalizedModel := service.NormalizeOpenAICompatRequestedModel(model)
	if normalizedModel == "" || normalizedModel == model {
		return body, normalizedModel, false, nil
	}

	normalizedBody, err := sjson.SetBytes(body, "model", normalizedModel)
	if err != nil {
		return body, model, false, err
	}
	return normalizedBody, normalizedModel, true, nil
}
