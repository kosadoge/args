package args

type Flag struct {
	Long     string
	Short    string
	Usage    string
	Value    Value
	DefValue string
}
