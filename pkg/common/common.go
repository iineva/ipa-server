package common

// get args until arg is not empty
func Def(args ...string) string {
	for _, v := range args {
		if v != "" {
			return v
		}
	}
	return ""
}
