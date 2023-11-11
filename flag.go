package args

type Flag struct {
	long     string
	short    string
	usage    string
	value    Value
	defValue string
}
