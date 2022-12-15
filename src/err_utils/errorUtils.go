package err_utils

func Panic_NonNilErr(err error) {
	if err != nil {
		panic(err)
	}
}
