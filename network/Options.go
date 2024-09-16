package network

type ServeType int32

const (
	TypeClient ServeType = 0
	TypeServer ServeType = 1
)

type IOptions interface {
	GetMaxRetires() int32
	GetTimeMills() int64
	GetData() interface{}
}

type Options struct {
	maxRetries int32
	timeMills  int64
	data       interface{}
}

func WithMax(maxRetries int32, timeMills int64, data interface{}) IOptions {
	return &Options{
		maxRetries: maxRetries,
		timeMills:  timeMills,
		data:       data,
	}
}

func (o *Options) GetMaxRetires() int32 {
	return o.maxRetries
}

func (o *Options) GetTimeMills() int64 {
	return o.timeMills
}

func (o *Options) GetData() interface{} {
	return o.data
}
