package play

func LoadPlaybook(app App, playbook []interface{}) []*Play {
	playBook := make([]*Play, len(playbook))
	for i, playI := range playbook {
		playBook[i] = ParseInterface(app, playI)
	}
	return playBook
}
