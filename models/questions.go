package models

type Question struct {
	Title       string
	Description string
	Url         string
}

type QuestionType int

const (
	QuestionTypeMatching QuestionType = iota
	QuestionTypeRelative
	QuestionTypeRadar
	QuestionTypeThermometer
	QuestionTypePicture
)

var QuestionName = map[QuestionType]string{
	QuestionTypeMatching:    "Matching",
	QuestionTypeRelative:    "Relative",
	QuestionTypeRadar:       "Radar",
	QuestionTypeThermometer: "Thermometer",
	QuestionTypePicture:     "Picture",
}
var matchingQuestionList = []Question{}

var QuestionMap = map[QuestionType][]Question{
	QuestionTypeMatching: {
		Question{
			Title:       "Same Bezirk",
			Description: "Ask if the Hider is in the same Bezirk as you are.",
			Url:         "sameBezirk",
		},
		Question{
			Title:       "Same Ortsteil",
			Description: "Ask if the Hider is in the same Ortsteil as you are.",
			Url:         "sameOrtsteil",
		},
		Question{
			Title:       "Last Letter of Ortsteil",
			Description: "Ask if the Ortsteil the hider is in has the same last letter as the Ortsteil you are in.",
			Url:         "ortsteilLastLetter",
		},
	},
	QuestionTypeRelative: {
		Question{
			Title:       "McDonald's Distance",
			Description: "Ask if the Hider is closer or further away from a McDonald's",
			Url:         "closerToMcDonalds",
		},
	},
	QuestionTypeRadar:       {},
	QuestionTypeThermometer: {},
	QuestionTypePicture:     {},
}

func GetQuestionMap() map[QuestionType][]Question {

	questionMap := make(map[QuestionType][]Question)
	questionMap[QuestionTypeMatching] = matchingQuestionList
	questionMap[QuestionTypeMatching] = matchingQuestionList

	return questionMap
}
