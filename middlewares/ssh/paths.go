package ssh

func GetHomePath(user string) string {
	return "/home/" + user
}

func GetSnipPath(user string) string {
	return GetHomePath(user) + "/.snip"
}

func GetRemotePath(user string, localPath string) string {
	return GetSnipPath(user) + "/" + localPath
}
