package play

func GetTreeKeyParts(parent interface{}) []string {
	var parts []string
	for {
		var part string
		switch p := parent.(type) {
		case *LoopRow:
			if p == nil {
				parent = nil
				break
			}
			part = "row." + p.GetKey()
			parent = p.ParentPlay
		case *Play:
			if p == nil {
				parent = nil
				break
			}
			part = "play." + p.GetKey()
			parent = p.ParentLoopRow
		case nil:
			parent = nil
		}
		if parent == nil {
			break
		}

		part = regNormalizeTreeKeyParts.ReplaceAllString(part, "_")
		parts = append([]string{part}, parts...)
	}
	return parts
}
