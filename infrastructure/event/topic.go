package event

type Topic struct {
	id string
}

func TriggerTopic() Topic {
	return Topic{id: "trigger"}
}
