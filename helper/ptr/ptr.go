package ptr

func StringToPtr(in string) *string {
	return &in
}

func PtrToString(in *string) string {
	if in == nil {
		return ""
	}
	return *in
}
